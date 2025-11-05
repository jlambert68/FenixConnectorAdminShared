package supportedSubInstructions

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ────────────────────────────────────────────────────────────────
// Public types
// ────────────────────────────────────────────────────────────────

// PreConditionRef references another instruction by (Class, Method).
type PreConditionRef struct {
	Class  string `json:"Class"`
	Method string `json:"Method"`
}

// SupportedSubInstruction describes one instruction and its prerequisites.
type SupportedSubInstruction struct {
	Class         string            `json:"Class"`
	Method        string            `json:"Method"`
	Description   string            `json:"Description,omitempty"`
	PreConditions []PreConditionRef `json:"PreConditions,omitempty"`
}

// SupportedSubInstructionsDocument is the root document shape.
type SupportedSubInstructionsDocument struct {
	SupportedSubInstructions []SupportedSubInstruction `json:"SupportedSubInstructions"`
}

// OrderItem represents one node in a valid execution order.
type OrderItem struct {
	ID     string `json:"id"`
	Class  string `json:"Class"`
	Method string `json:"Method"`
}

// ValidationResult is returned by ValidatePreconditions.
type ValidationResult struct {
	OK       bool        `json:"ok"`
	Errors   []string    `json:"errors"`
	Warnings []string    `json:"warnings,omitempty"`
	Order    []OrderItem `json:"order,omitempty"`
}

// ────────────────────────────────────────────────────────────────
// Internal graph model
// ────────────────────────────────────────────────────────────────

type node struct {
	ID      string
	Class   string
	Method  string
	Prereqs []string // edges to prerequisite IDs
}

func idOf(class, method string) string {
	return class + "|" + method
}

type graph struct {
	nodes map[string]*node // id -> node
}

// buildGraph constructs the node map and performs basic checks:
// - Class/Method must be non-empty strings
// - no duplicates
// - unknown references
// - self-dependency
func buildGraph(doc SupportedSubInstructionsDocument) (*graph, []string) {
	g := &graph{nodes: make(map[string]*node)}
	var errs []string

	list := doc.SupportedSubInstructions
	// 1) Register nodes & detect duplicates / invalids
	for _, it := range list {
		if strings.TrimSpace(it.Class) == "" || strings.TrimSpace(it.Method) == "" {
			errs = append(errs, fmt.Sprintf(`Invalid item: missing non-empty "Class" or "Method" in %s`, mustJSON(it)))
			continue
		}
		id := idOf(it.Class, it.Method)
		if _, exists := g.nodes[id]; exists {
			errs = append(errs, fmt.Sprintf("Duplicate definition: (%s, %s)", it.Class, it.Method))
			continue
		}
		g.nodes[id] = &node{
			ID:      id,
			Class:   it.Class,
			Method:  it.Method,
			Prereqs: make([]string, 0),
		}
	}

	// 2) Wire edges & check unknown/self references
	for _, it := range list {
		fromID := idOf(it.Class, it.Method)
		n := g.nodes[fromID]
		// If node was invalid/duplicate we won't have it in the map
		if n == nil {
			continue
		}

		for _, pre := range it.PreConditions {
			if strings.TrimSpace(pre.Class) == "" || strings.TrimSpace(pre.Method) == "" {
				errs = append(errs, fmt.Sprintf(`Invalid precondition under (%s, %s): missing non-empty Class/Method`, it.Class, it.Method))
				continue
			}
			toID := idOf(pre.Class, pre.Method)
			if fromID == toID {
				errs = append(errs, fmt.Sprintf("Self-dependency: (%s, %s) depends on itself", it.Class, it.Method))
				continue
			}
			if _, ok := g.nodes[toID]; !ok {
				errs = append(errs, fmt.Sprintf("Unknown precondition: (%s, %s) -> (%s, %s)", it.Class, it.Method, pre.Class, pre.Method))
				continue
			}
			n.Prereqs = append(n.Prereqs, toID)
		}
	}

	return g, errs
}

// detectCycles performs DFS to find back-edges indicating cycles.
// Returns all cycle paths as slices of IDs.
func detectCycles(g *graph) (hasCycle bool, cycles [][]string) {
	const (
		unseen   = 0
		visiting = 1
		visited  = 2
	)
	state := make(map[string]int)
	stack := make([]string, 0)

	var dfs func(u string)
	dfs = func(u string) {
		state[u] = visiting
		stack = append(stack, u)

		for _, v := range g.nodes[u].Prereqs {
			switch state[v] {
			case unseen:
				dfs(v)
			case visiting:
				// Found back-edge u -> v, capture cycle v..u plus v
				start := 0
				for i := range stack {
					if stack[i] == v {
						start = i
						break
					}
				}
				cycle := append(append([]string{}, stack[start:]...), v)
				cycles = append(cycles, cycle)
			}
		}

		// pop u
		stack = stack[:len(stack)-1]
		state[u] = visited
	}

	for id := range g.nodes {
		if state[id] == unseen {
			dfs(id)
		}
	}
	return len(cycles) > 0, cycles
}

// topoSort returns a valid execution order using Kahn's algorithm.
// Nodes appear only after all their prerequisites.
func topoSort(g *graph) ([]string, error) {
	indeg := make(map[string]int, len(g.nodes))
	dependents := make(map[string][]string, len(g.nodes))

	// Initialize indegrees and dependents
	for id := range g.nodes {
		indeg[id] = 0
	}
	for id, n := range g.nodes {
		indeg[id] = len(n.Prereqs)
		for _, p := range n.Prereqs {
			dependents[p] = append(dependents[p], id)
		}
	}

	// queue of zero-indegree nodes
	q := make([]string, 0)
	for id, d := range indeg {
		if d == 0 {
			q = append(q, id)
		}
	}

	order := make([]string, 0, len(g.nodes))
	for len(q) > 0 {
		// pop front
		u := q[0]
		q = q[1:]
		order = append(order, u)

		for _, v := range dependents[u] {
			indeg[v]--
			if indeg[v] == 0 {
				q = append(q, v)
			}
		}
	}

	if len(order) != len(g.nodes) {
		return nil, errors.New("cycle detected (topological sort failed)")
	}
	return order, nil
}

// mustJSON marshals v to JSON for error messages; never panics in practice.
func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// ────────────────────────────────────────────────────────────────
// Public API
// ────────────────────────────────────────────────────────────────

// ValidatePreconditions checks structure, references, self-deps, and cycles.
// If includeOrder is true, it also computes a safe execution order.
func ValidatePreconditions(doc SupportedSubInstructionsDocument, includeOrder bool) ValidationResult {
	g, errs := buildGraph(doc)
	if len(errs) > 0 {
		return ValidationResult{OK: false, Errors: errs}
	}

	// Cycle detection
	if has, cyc := detectCycles(g); has {
		for _, c := range cyc {
			var parts []string
			for _, id := range c {
				// id is Class|Method
				s := strings.SplitN(id, "|", 2)
				if len(s) == 2 {
					parts = append(parts, fmt.Sprintf("(%s, %s)", s[0], s[1]))
				} else {
					parts = append(parts, id)
				}
			}
			errs = append(errs, "Cycle detected: "+strings.Join(parts, " -> "))
		}
		return ValidationResult{OK: false, Errors: errs}
	}

	res := ValidationResult{OK: true, Errors: nil, Warnings: nil}

	if includeOrder {
		ids, err := topoSort(g)
		if err != nil {
			res.OK = false
			res.Errors = []string{err.Error()}
			return res
		}
		order := make([]OrderItem, 0, len(ids))
		for _, id := range ids {
			n := g.nodes[id]
			order = append(order, OrderItem{
				ID:     id,
				Class:  n.Class,
				Method: n.Method,
			})
		}
		res.Order = order
	}

	return res
}

// GetExecutionOrder returns a valid execution order or an error with details.
func GetExecutionOrder(doc SupportedSubInstructionsDocument) ([]OrderItem, error) {
	res := ValidatePreconditions(doc, true)
	if !res.OK {
		return nil, fmt.Errorf("preconditions validation failed:\n - %s", strings.Join(res.Errors, "\n - "))
	}
	return res.Order, nil
}
