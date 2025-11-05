package supportedSubInstructions

import (
	"fmt"
	"strings"
)

// ────────────────────────────────────────────────────────────────
// Public API types
// ────────────────────────────────────────────────────────────────

/*
type PreConditionRef struct {
	Class  string `json:"Class"`
	Method string `json:"Method"`
}

type SupportedSubInstruction struct {
	Class         string            `json:"Class"`
	Method        string            `json:"Method"`
	Description   string            `json:"Description,omitempty"`
	PreConditions []PreConditionRef `json:"PreConditions,omitempty"`
}

*/

type Catalog struct {
	SupportedSubInstructions []SupportedSubInstruction `json:"SupportedSubInstructions"`
}

type PlanRoot struct {
	Class           string              `json:"Class"`
	ActionMethod    string              `json:"ActionMethod"`
	SubInstructions PlanSubInstructions `json:"SubInstructions"`
}

type PlanSubInstructions struct {
	SubInstructionContainer *PlanContainer `json:"SubInstructionContainer,omitempty"`
}

type PlanContainer struct {
	ExecutionType                   string            `json:"ExecutionType"` // "Serial" | "Parallel"
	SubInstructionContainerChildren []PlanNodeWrapper `json:"SubInstructionContainerChildren"`
}

type PlanNodeWrapper struct {
	SubInstructionContainer *PlanContainer `json:"SubInstructionContainer,omitempty"`
	SubInstruction          *PlanLeaf      `json:"SubInstruction,omitempty"`
}

type PlanLeaf struct {
	Class  string `json:"Class"`
	Method string `json:"Method"`
}

type Result struct {
	OK     bool     `json:"ok"`
	Errors []string `json:"errors"`
}

// ────────────────────────────────────────────────────────────────
// Catalog indexing
// ────────────────────────────────────────────────────────────────

type catalogIndex struct {
	exists map[string]bool
	preqs  map[string][]string
}

//func idOf(c, m string) string { return c + "|" + m }

func buildCatalogIndex(cat Catalog) (*catalogIndex, []string) {
	idx := &catalogIndex{
		exists: make(map[string]bool),
		preqs:  make(map[string][]string),
	}
	var errs []string

	for _, it := range cat.SupportedSubInstructions {
		c := strings.TrimSpace(it.Class)
		m := strings.TrimSpace(it.Method)
		if c == "" || m == "" {
			errs = append(errs, fmt.Sprintf(`Catalog item missing Class/Method: %s`, mustJSON(it)))
			continue
		}
		id := idOf(c, m)
		if idx.exists[id] {
			errs = append(errs, fmt.Sprintf("Duplicate catalog entry: (%s, %s)", c, m))
			continue
		}
		idx.exists[id] = true
	}

	for _, it := range cat.SupportedSubInstructions {
		from := idOf(it.Class, it.Method)
		if !idx.exists[from] {
			continue
		}
		for _, p := range it.PreConditions {
			pc := strings.TrimSpace(p.Class)
			pm := strings.TrimSpace(p.Method)
			if pc == "" || pm == "" {
				errs = append(errs, fmt.Sprintf("Invalid precondition under (%s, %s): missing Class/Method", it.Class, it.Method))
				continue
			}
			to := idOf(pc, pm)
			if !idx.exists[to] {
				errs = append(errs, fmt.Sprintf("Unknown precondition in catalog: (%s, %s) -> (%s, %s)", it.Class, it.Method, pc, pm))
				continue
			}
			if from == to {
				errs = append(errs, fmt.Sprintf("Self-precondition in catalog: (%s, %s)", it.Class, it.Method))
				continue
			}
			idx.preqs[from] = append(idx.preqs[from], to)
		}
	}

	return idx, errs
}

// func mustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }

// ────────────────────────────────────────────────────────────────
// Public API
// ────────────────────────────────────────────────────────────────

/*
ValidatePlan validates the execution tree against the catalog:

  - Every leaf must exist in the catalog.
  - SERIAL: each child may rely on anything executed by earlier siblings.
  - PARALLEL: each child must have all preconditions satisfied BEFORE the group starts
    (no relying on siblings inside the same parallel batch).
  - Nested containers are handled recursively.

An empty SubInstructions object (no top container) is valid.
*/
func ValidatePlan(cat Catalog, plan PlanRoot) Result {
	idx, errs := buildCatalogIndex(cat)
	if len(errs) > 0 {
		return Result{OK: false, Errors: errs}
	}

	// Executed set as we walk the tree
	executed := make(map[string]bool)

	if plan.SubInstructions.SubInstructionContainer == nil {
		return Result{OK: true}
	}

	errs, _ = validateContainer(idx, plan.SubInstructions.SubInstructionContainer, "$.SubInstructions.SubInstructionContainer", executed)
	if len(errs) > 0 {
		return Result{OK: false, Errors: errs}
	}
	return Result{OK: true}
}

// ────────────────────────────────────────────────────────────────
/*
Implementation notes:

We return the set of IDs each node/container would "add" to the executed set,
instead of mutating via hidden modes. This keeps SERIAL/PARALLEL logic explicit.

- validateContainer(..., executed) -> (errs, added)
  - SERIAL:
      local := copy(executed)
      for child:
         errsChild, addChild := validateNode(..., local)
         if errsChild -> stop
         merge local with addChild
      added = diff(local, executed)
  - PARALLEL:
      pre := copy(executed)  // snapshot before group
      for child:
         errsChild, addChild := validateNode(..., pre) // eval vs snapshot
         if errsChild -> stop
         merge groupAdds with addChild
      added = groupAdds
- validateNode delegates to either leaf or container.

At the top level we ignore the returned 'added' since we only need validity.
*/
// ────────────────────────────────────────────────────────────────

func validateContainer(idx *catalogIndex, c *PlanContainer, path string, executed map[string]bool) ([]string, map[string]bool) {
	if c.ExecutionType != "Serial" && c.ExecutionType != "Parallel" {
		return []string{fmt.Sprintf("%s.ExecutionType must be 'Serial' or 'Parallel'", path)}, nil
	}

	switch c.ExecutionType {

	case "Serial":
		local := copySet(executed)
		for i, child := range c.SubInstructionContainerChildren {
			childPath := fmt.Sprintf("%s.SubInstructionContainerChildren[%d]", path, i)
			errs, add := validateNode(idx, &child, childPath, local)
			if len(errs) > 0 {
				return errs, nil
			}
			mergeInto(local, add)
		}
		return nil, diffSet(local, executed)

	case "Parallel":
		pre := copySet(executed)           // snapshot
		groupAdds := make(map[string]bool) // what the group contributes
		for i, child := range c.SubInstructionContainerChildren {
			childPath := fmt.Sprintf("%s.SubInstructionContainerChildren[%d]", path, i)
			errs, add := validateNode(idx, &child, childPath, pre) // each vs snapshot
			if len(errs) > 0 {
				return errs, nil
			}
			mergeInto(groupAdds, add)
		}
		return nil, groupAdds
	}

	return nil, nil
}

func validateNode(idx *catalogIndex, n *PlanNodeWrapper, path string, executed map[string]bool) ([]string, map[string]bool) {
	hasContainer := n.SubInstructionContainer != nil
	hasLeaf := n.SubInstruction != nil
	if hasContainer && hasLeaf {
		return []string{fmt.Sprintf("%s must contain either SubInstructionContainer or SubInstruction, not both", path)}, nil
	}
	if !hasContainer && !hasLeaf {
		return []string{fmt.Sprintf("%s must contain either SubInstructionContainer or SubInstruction", path)}, nil
	}

	if hasContainer {
		return validateContainer(idx, n.SubInstructionContainer, path+".SubInstructionContainer", executed)
	}

	leaf := n.SubInstruction
	id := idOf(leaf.Class, leaf.Method)
	if !idx.exists[id] {
		return []string{fmt.Sprintf("%s.SubInstruction refers to unknown catalog entry (%s, %s)", path, leaf.Class, leaf.Method)}, nil
	}

	// Check preconditions against the provided executed set.
	missing := missingPreconditions(idx, id, executed)
	if len(missing) > 0 {
		return []string{fmt.Sprintf("%s.SubInstruction (%s, %s) missing preconditions: %s",
			path, leaf.Class, leaf.Method, strings.Join(missing, ", "))}, nil
	}

	// A leaf adds itself.
	added := map[string]bool{id: true}
	return nil, added
}

func missingPreconditions(idx *catalogIndex, id string, executed map[string]bool) []string {
	var miss []string
	for _, p := range idx.preqs[id] {
		if !executed[p] {
			parts := strings.SplitN(p, "|", 2)
			if len(parts) == 2 {
				miss = append(miss, fmt.Sprintf("(%s, %s)", parts[0], parts[1]))
			} else {
				miss = append(miss, p)
			}
		}
	}
	return miss
}

// ────────────────────────────────────────────────────────────────
// Small set helpers
// ────────────────────────────────────────────────────────────────

func copySet(m map[string]bool) map[string]bool {
	out := make(map[string]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func mergeInto(dst, src map[string]bool) {
	for k := range src {
		dst[k] = true
	}
}

// diffSet returns keys present in a but not in b.
func diffSet(a, b map[string]bool) map[string]bool {
	out := make(map[string]bool)
	for k := range a {
		if !b[k] {
			out[k] = true
		}
	}
	return out
}
