package model

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/dominikbraun/graph"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

// stepDefinition is the parsed representation of a step in a workflow.
//
// Within the workflow spec, they are called "Capability Properties".
type stepDefinition struct {
	ID     string         `json:"id" jsonschema:"required"`
	Ref    string         `json:"ref,omitempty" jsonschema:"pattern=^[a-z0-9_]+$"`
	Inputs map[string]any `json:"inputs,omitempty"`
	Config map[string]any `json:"config" jsonschema:"required"`

	CapabilityType capabilities.CapabilityType `json:"-"`
}

// workflowSpec is the parsed representation of a workflow.
type workflowSpec struct {
	Triggers  []stepDefinition `json:"triggers" jsonschema:"required"`
	Actions   []stepDefinition `json:"actions,omitempty"`
	Consensus []stepDefinition `json:"consensus" jsonschema:"required"`
	Targets   []stepDefinition `json:"targets" jsonschema:"required"`
}

func (w *workflowSpec) steps() []stepDefinition {
	s := []stepDefinition{}
	s = append(s, w.Actions...)
	s = append(s, w.Consensus...)
	s = append(s, w.Targets...)
	return s
}

// workflow is a directed graph of nodes, where each node is a step.
//
// triggers are special steps that are stored separately, they're
// treated differently due to their nature of being the starting
// point of a workflow.
type workflow struct {
	id string
	graph.Graph[string, *step]

	triggers []*triggerCapability

	spec *workflowSpec
}

func (w *workflow) walkDo(start string, do func(s *step) error) error {
	var outerErr error
	err := graph.BFS(w.Graph, start, func(ref string) bool {
		n, err := w.Graph.Vertex(ref)
		if err != nil {
			outerErr = err
			return true
		}

		err = do(n)
		if err != nil {
			outerErr = err
			return true
		}

		return false
	})
	if err != nil {
		return err
	}

	return outerErr
}

func (w *workflow) dependents(start string) ([]*step, error) {
	steps := []*step{}
	m, err := w.Graph.AdjacencyMap()
	if err != nil {
		return nil, err
	}

	adj, ok := m[start]
	if !ok {
		return nil, fmt.Errorf("could not find step with ref %s", start)
	}

	for adjacentRef := range adj {
		n, err := w.Graph.Vertex(adjacentRef)
		if err != nil {
			return nil, err
		}

		steps = append(steps, n)
	}

	return steps, nil
}

// step wraps a stepDefinition with additional context for dependencies and execution
type step struct {
	stepDefinition
	dependencies      []string
	capability        capabilities.CallbackCapability
	config            *values.Map
	executionStrategy ExecutionStrategy
}

type triggerCapability struct {
	stepDefinition
	trigger capabilities.TriggerCapability
	config  *values.Map
}

const (
	keywordTrigger = "trigger"
)

func Parse(yamlWorkflow string) (*workflow, error) {
	spec, err := ParseWorkflowSpecYaml(yamlWorkflow)
	if err != nil {
		return nil, err
	}

	// Construct and validate the graph. We instantiate an
	// empty graph with just one starting entry: `trigger`.
	// This provides the starting point for our graph and
	// points to all dependent steps.
	// Note: all triggers are represented by a single step called
	// `trigger`. This is because for workflows with multiple triggers
	// only one trigger will have started the workflow.
	stepHash := func(s *step) string {
		return s.Ref
	}
	g := graph.New(
		stepHash,
		graph.PreventCycles(),
		graph.Directed(),
	)
	err = g.AddVertex(&step{
		stepDefinition: stepDefinition{Ref: keywordTrigger},
	})
	if err != nil {
		return nil, err
	}

	// Next, let's populate the other entries in the graph.
	for _, s := range spec.steps() {
		// TODO: The workflow format spec doesn't always require a `Ref`
		// to be provided (triggers and targets don't have a `Ref` for example).
		// To handle this, we default the `Ref` to the type, but ideally we
		// should find a better long-term way to handle this.
		if s.Ref == "" {
			s.Ref = s.ID
		}

		innerErr := g.AddVertex(&step{stepDefinition: s})
		if innerErr != nil {
			return nil, fmt.Errorf("cannot add vertex %s: %w", s.Ref, innerErr)
		}
	}

	stepRefs, err := g.AdjacencyMap()
	if err != nil {
		return nil, err
	}

	// Next, let's iterate over the steps and populate
	// any edges.
	for stepRef := range stepRefs {
		step, innerErr := g.Vertex(stepRef)
		if innerErr != nil {
			return nil, innerErr
		}

		refs, innerErr := findRefs(step.Inputs)
		if innerErr != nil {
			return nil, innerErr
		}
		step.dependencies = refs

		if stepRef != keywordTrigger && len(refs) == 0 {
			return nil, errors.New("all non-trigger steps must have a dependent ref")
		}

		for _, r := range refs {
			innerErr = g.AddEdge(r, step.Ref)
			if innerErr != nil {
				return nil, innerErr
			}
		}
	}

	triggerSteps := []*triggerCapability{}
	for _, t := range spec.Triggers {
		triggerSteps = append(triggerSteps, &triggerCapability{
			stepDefinition: t,
		})
	}
	wf := &workflow{
		spec:     &spec,
		Graph:    g,
		triggers: triggerSteps,
	}
	return wf, err
}

var (
	interpolationTokenRe = regexp.MustCompile(`^\$\((\S+)\)$`)
)

// findRefs takes an `inputs` map and returns a list of all the step references
// contained within it.
func findRefs(inputs map[string]any) ([]string, error) {
	refs := []string{}
	_, err := deepMap(
		inputs,
		// This function is called for each string in the map
		// for each string, we iterate over each match of the interpolation token
		// - if there are no matches, return no reference
		// - if there is one match, return the reference
		// - if there are multiple matches (in the case of a multi-part state reference), return just the step ref
		func(el string) (any, error) {
			matches := interpolationTokenRe.FindStringSubmatch(el)
			if len(matches) < 2 {
				return el, nil
			}

			m := matches[1]
			parts := strings.Split(m, ".")
			if len(parts) < 1 {
				return nil, fmt.Errorf("invalid ref %s", m)
			}

			refs = append(refs, parts[0])
			return el, nil
		},
	)
	return refs, err
}

// deepMap recursively applies a transformation function
// over each string within:
//
//   - a map[string]any
//   - a []any
//   - a string
func deepMap(input any, transform func(el string) (any, error)) (any, error) {
	// in the case of a string, simply apply the transformation
	// in the case of a map, recurse and apply the transformation to each value
	// in the case of a list, recurse and apply the transformation to each element
	switch tv := input.(type) {
	case string:
		nv, err := transform(tv)
		if err != nil {
			return nil, err
		}

		return nv, nil
	case mapping:
		// coerce mapping to map[string]any
		mp := map[string]any(tv)

		nm := map[string]any{}
		for k, v := range mp {
			nv, err := deepMap(v, transform)
			if err != nil {
				return nil, err
			}

			nm[k] = nv
		}
		return nm, nil
	case map[string]any:
		nm := map[string]any{}
		for k, v := range tv {
			nv, err := deepMap(v, transform)
			if err != nil {
				return nil, err
			}

			nm[k] = nv
		}
		return nm, nil
	case []any:
		a := []any{}
		for _, el := range tv {
			ne, err := deepMap(el, transform)
			if err != nil {
				return nil, err
			}

			a = append(a, ne)
		}
		return a, nil
	}

	return nil, fmt.Errorf("cannot traverse item %+v of type %T", input, input)
}
