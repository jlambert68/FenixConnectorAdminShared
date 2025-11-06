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

// Result for plan validation (without order)
type Result struct {
	OK     bool     `json:"ok"`
	Errors []string `json:"errors"`
}

// An atomic executable leaf in the order
type OrderLeaf struct {
	ID     string `json:"id"` // "Class|Method"
	Class  string `json:"Class"`
	Method string `json:"Method"`
}

// A Stage is a parallel batch: all leaves in the Stage can run together
type Stage struct {
	Parallel []OrderLeaf `json:"parallel"`
}

// ExecutionOrder is the full ordered set of stages
type ExecutionOrder struct {
	Stages []Stage `json:"stages"`
}

// ────────────────────────────────────────────────────────────────
// Catalog indexing (normalized lookups)
// ────────────────────────────────────────────────────────────────

type catalogIndex struct {
	exists map[string]bool
	preqs  map[string][]string
}

// func idOf(c, m string) string { return c + "|" + m }
func normalize(s string) string { return strings.TrimSpace(s) }

//func mustJSON(v any) string { b, _ := json.Marshal(v); return string(b) }

func buildCatalogIndex(cat Catalog) (*catalogIndex, []string) {
	idx := &catalogIndex{
		exists: make(map[string]bool),
		preqs:  make(map[string][]string),
	}
	var errs []string

	// Register items and catch duplicates / empty fields
	for _, it := range cat.SupportedSubInstructions {
		c := normalize(it.Class)
		m := normalize(it.Method)
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

	// Wire preconditions
	for _, it := range cat.SupportedSubInstructions {
		from := idOf(normalize(it.Class), normalize(it.Method))
		if !idx.exists[from] {
			continue
		}
		for _, p := range it.PreConditions {
			pc := normalize(p.Class)
			pm := normalize(p.Method)
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

// ────────────────────────────────────────────────────────────────
// Public: basic validation only
// ────────────────────────────────────────────────────────────────

// ValidatePlan checks existence + preconditions under Serial/Parallel semantics.
// Empty SubInstructions (no container) is considered valid.
func ValidatePlan(cat Catalog, plan PlanRoot) Result {
	idx, errs := buildCatalogIndex(cat)
	if len(errs) > 0 {
		return Result{OK: false, Errors: errs}
	}
	if plan.SubInstructions.SubInstructionContainer == nil {
		return Result{OK: true}
	}
	executed := make(map[string]bool)
	if errs, _ := validateContainer(idx, plan.SubInstructions.SubInstructionContainer, "$.SubInstructions.SubInstructionContainer", executed); len(errs) > 0 {
		return Result{OK: false, Errors: errs}
	}
	return Result{OK: true}
}

// ────────────────────────────────────────────────────────────────
// Public: compute full execution order (stages)
// ────────────────────────────────────────────────────────────────

// ComputeExecutionOrder validates the plan and returns ordered stages.
// Each Stage is a parallel batch. Serial containers concatenate stages;
// Parallel containers merge stages by index (barrier semantics).
func ComputeExecutionOrder(cat Catalog, plan PlanRoot) (ExecutionOrder, []string) {
	idx, errs := buildCatalogIndex(cat)
	if len(errs) > 0 {
		return ExecutionOrder{}, errs
	}
	// No container ⇒ no work
	if plan.SubInstructions.SubInstructionContainer == nil {
		return ExecutionOrder{Stages: nil}, nil
	}

	executed := make(map[string]bool)
	errs, stages, _ := buildStagesForContainer(idx, plan.SubInstructions.SubInstructionContainer, "$.SubInstructions.SubInstructionContainer", executed)
	if len(errs) > 0 {
		return ExecutionOrder{}, errs
	}
	return ExecutionOrder{Stages: stages}, nil
}

// ────────────────────────────────────────────────────────────────
// Core validators that also build stages
// ────────────────────────────────────────────────────────────────

// validateContainer enforces semantics and returns (errs, addedIDs).
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
		pre := copySet(executed)
		groupAdds := make(map[string]bool)
		for i, child := range c.SubInstructionContainerChildren {
			childPath := fmt.Sprintf("%s.SubInstructionContainerChildren[%d]", path, i)
			errs, add := validateNode(idx, &child, childPath, pre) // evaluated vs snapshot
			if len(errs) > 0 {
				return errs, nil
			}
			mergeInto(groupAdds, add)
		}
		return nil, groupAdds
	}
	return nil, nil
}

// buildStagesForContainer validates and *builds* stage list; returns (errs, stages, addedIDs).
func buildStagesForContainer(idx *catalogIndex, c *PlanContainer, path string, executed map[string]bool) ([]string, []Stage, map[string]bool) {
	if c.ExecutionType != "Serial" && c.ExecutionType != "Parallel" {
		return []string{fmt.Sprintf("%s.ExecutionType must be 'Serial' or 'Parallel'", path)}, nil, nil
	}

	switch c.ExecutionType {
	case "Serial":
		local := copySet(executed)
		var out []Stage
		for i, child := range c.SubInstructionContainerChildren {
			childPath := fmt.Sprintf("%s.SubInstructionContainerChildren[%d]", path, i)
			errs, stages, add := buildStagesForNode(idx, &child, childPath, local)
			if len(errs) > 0 {
				return errs, nil, nil
			}
			out = append(out, stages...)
			mergeInto(local, add)
		}
		return nil, out, diffSet(local, executed)

	case "Parallel":
		pre := copySet(executed)
		// Collect child stage lists against the same snapshot
		var childStages [][]Stage
		groupAdds := make(map[string]bool)

		for i, child := range c.SubInstructionContainerChildren {
			childPath := fmt.Sprintf("%s.SubInstructionContainerChildren[%d]", path, i)
			errs, stages, add := buildStagesForNode(idx, &child, childPath, pre)
			if len(errs) > 0 {
				return errs, nil, nil
			}
			childStages = append(childStages, stages)
			mergeInto(groupAdds, add)
		}

		// Merge by index (barrier)
		merged := mergeStagesByIndex(childStages)
		return nil, merged, groupAdds
	}

	return nil, nil, nil
}

// validateNode returns (errs, addedIDs).
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
	// Leaf
	leaf := n.SubInstruction
	lc := normalize(leaf.Class)
	lm := normalize(leaf.Method)
	id := idOf(lc, lm)

	if !idx.exists[id] {
		// Helpful hint (first few entries) during integration
		hints := make([]string, 0, 8)
		count := 0
		for k := range idx.exists {
			if count >= 8 {
				break
			}
			parts := strings.SplitN(k, "|", 2)
			if len(parts) == 2 {
				hints = append(hints, fmt.Sprintf("(%s, %s)", parts[0], parts[1]))
			}
			count++
		}
		hintMsg := ""
		if len(hints) > 0 {
			hintMsg = fmt.Sprintf(" Known catalog examples: %s", strings.Join(hints, "; "))
		}
		return []string{fmt.Sprintf("%s.SubInstruction refers to unknown catalog entry (%s, %s).%s", path, lc, lm, hintMsg)}, nil
	}

	if missing := missingPreconditions(idx, id, executed); len(missing) > 0 {
		return []string{fmt.Sprintf("%s.SubInstruction (%s, %s) missing preconditions: %s", path, lc, lm, strings.Join(missing, ", "))}, nil
	}

	return nil, map[string]bool{id: true}
}

// buildStagesForNode validates and returns (errs, stages, addedIDs).
func buildStagesForNode(idx *catalogIndex, n *PlanNodeWrapper, path string, executed map[string]bool) ([]string, []Stage, map[string]bool) {
	hasContainer := n.SubInstructionContainer != nil
	hasLeaf := n.SubInstruction != nil
	if hasContainer && hasLeaf {
		return []string{fmt.Sprintf("%s must contain either SubInstructionContainer or SubInstruction, not both", path)}, nil, nil
	}
	if !hasContainer && !hasLeaf {
		return []string{fmt.Sprintf("%s must contain either SubInstructionContainer or SubInstruction", path)}, nil, nil
	}
	if hasContainer {
		return buildStagesForContainer(idx, n.SubInstructionContainer, path+".SubInstructionContainer", executed)
	}

	// Leaf => a single-stage with one parallel item
	leaf := n.SubInstruction
	lc := normalize(leaf.Class)
	lm := normalize(leaf.Method)
	id := idOf(lc, lm)

	if !idx.exists[id] {
		return []string{fmt.Sprintf("%s.SubInstruction refers to unknown catalog entry (%s, %s)", path, lc, lm)}, nil, nil
	}
	if missing := missingPreconditions(idx, id, executed); len(missing) > 0 {
		return []string{fmt.Sprintf("%s.SubInstruction (%s, %s) missing preconditions: %s", path, lc, lm, strings.Join(missing, ", "))}, nil, nil
	}

	stage := Stage{Parallel: []OrderLeaf{{ID: id, Class: lc, Method: lm}}}
	return nil, []Stage{stage}, map[string]bool{id: true}
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
// Stage merge helpers for Parallel containers
// ────────────────────────────────────────────────────────────────

func mergeStagesByIndex(children [][]Stage) []Stage {
	// Find max number of stages among children
	max := 0
	for _, st := range children {
		if len(st) > max {
			max = len(st)
		}
	}
	if max == 0 {
		return nil
	}
	out := make([]Stage, max)
	for i := 0; i < max; i++ {
		var batch []OrderLeaf
		for _, st := range children {
			if i < len(st) {
				batch = append(batch, st[i].Parallel...)
			}
		}
		out[i] = Stage{Parallel: batch}
	}
	return out
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

func diffSet(a, b map[string]bool) map[string]bool {
	out := make(map[string]bool)
	for k := range a {
		if !b[k] {
			out[k] = true
		}
	}
	return out
}
