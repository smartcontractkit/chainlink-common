package sdk

import (
	"cmp"
	"iter"
	"slices"
	"strings"
	"text/template"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

type StepInputs struct {
	OutputRef string
	Mapping   map[string]any
}

type Output struct {
	Ref, Name string
}

func (i *StepInputs) HasRef(ref string) bool {
	if i.OutputRef != "" {
		return i.OutputRef == ref
	}
	for s := range flatten(i.Mapping) {
		if parseRef(s) == ref {
			return true
		}
	}
	return false
}

// Outputs returns only outputs from Mapping, sorted and grouped by Ref
func (i *StepInputs) Outputs() []Output {
	m := make(outputs)
	m.add(i.Mapping)
	var s []Output
	for ref, name := range m {
		if len(name) == 0 {
			s = append(s, Output{Ref: ref})
			continue
		}
		slices.Sort(name)
		s = append(s, Output{Ref: ref, Name: strings.Join(name, "<br>")})
	}
	slices.SortFunc(s, func(a, b Output) int {
		return cmp.Or(
			cmp.Compare(a.Ref, b.Ref),
			cmp.Compare(a.Name, b.Name),
		)
	})
	return s
}

type outputs map[string][]string

func flatten(m map[string]any) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, v := range m {
			switch t := v.(type) {
			case []map[string]any:
				for _, m := range t {
					for s := range flatten(m) {
						if !yield(s) {
							return
						}
					}
				}
			case map[string]any:
				for s := range flatten(t) {
					if !yield(s) {
						return
					}
				}
			case []string:
				for _, s := range t {
					if !yield(s) {
						return
					}
				}
			case string:
				if !yield(t) {
					return
				}
			}
		}
	}
}

func (os outputs) add(m map[string]any) {
	for s := range flatten(m) {
		os.addOutput(s)
	}
}

func parseRef(s string) string {
	if strings.HasPrefix(s, "$(") {
		s = s[2 : len(s)-1] // trim $()
		parts := strings.SplitN(s, ".", 2)
		if len(parts) != 2 {
			return s
		}
		return parts[0]
	}
	return ""
}

func (os outputs) addOutput(s string) {
	if strings.HasPrefix(s, "$(") {
		s = s[2 : len(s)-1] // trim $()
		if strings.HasSuffix(s, ".outputs") {
			if _, ok := os[s[:len(s)-len(".outputs")]]; !ok {
				os[s[:len(s)-len(".outputs")]] = []string{}
			}
			return
		}
		parts := strings.SplitN(s, ".outputs.", 2)
		if len(parts) != 2 {
			return
		}
		ref, name := parts[0], parts[1]
		os[ref] = append(os[ref], name)
	}
}

// StepDefinition is the parsed representation of a step in a workflow.
//
// Within the workflow spec, they are called "Capability Properties".
type StepDefinition struct {
	ID        string
	Ref       string
	Condition string
	Inputs    StepInputs
	Config    map[string]any

	CapabilityType capabilities.CapabilityType
}

type WorkflowSpec struct {
	Name      string
	Owner     string
	Triggers  []StepDefinition
	Actions   []StepDefinition
	Consensus []StepDefinition
	Targets   []StepDefinition
}

func (w *WorkflowSpec) Steps() []StepDefinition {
	s := []StepDefinition{}
	s = append(s, w.Actions...)
	s = append(s, w.Consensus...)
	s = append(s, w.Targets...)
	return s
}

func (w *WorkflowSpec) FormatChart() (string, error) {
	var sb strings.Builder
	steps := slices.Clone(w.Triggers)
	steps = append(steps, w.Steps()...)
	slices.SortFunc(steps, func(a, b StepDefinition) int {
		return cmp.Or(
			a.CapabilityType.Compare(b.CapabilityType),
			cmp.Compare(a.Ref, b.Ref),
			cmp.Compare(a.ID, b.ID),
		)
	})
	err := tmpl.Execute(&sb, steps)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

var tmpl = template.Must(template.New("").Funcs(map[string]any{
	"replace":  strings.ReplaceAll,
	"parseRef": parseRef,
}).Parse(`flowchart
{{ range $i, $step := . }}
	{{ $ref := .Ref -}}
	{{ $id := replace .ID "@" "[at]" -}}
	{{ $name := printf "%s<br><i>(%s)</i>" .CapabilityType $id -}}
	{{ if .Ref -}}
		{{ $name = printf "<b>%s</b><br>%s" .Ref $name -}}
	{{ else -}}
		{{ $ref = printf "%s%d" "unnamed" $i -}}
	{{ end -}}
	{{ if eq .CapabilityType "trigger" -}}
	{{ $ref }}[\"{{$name}}"/]
	{{ else if eq .CapabilityType "consensus" -}}
	{{ $ref }}[["{{$name}}"]]
	{{ else if eq .CapabilityType "target" -}}
	{{ $ref }}[/"{{$name}}"\]
	{{ else -}}
	{{ $ref }}["{{$name}}"]
	{{ end -}}
	{{ $condRef := parseRef .Condition -}}
	{{ if $condRef -}}
		{{ $condRef }} -..-> {{ $step.Ref }}
	{{ end -}}
	{{ if .Inputs.OutputRef -}}
	{{ .Inputs.OutputRef }} --> {{ $step.Ref }}
	{{ else -}}
		{{ range $out := .Inputs.Outputs -}}
			{{ if $out.Name -}}
	{{ $out.Ref }} -- {{ $out.Name }} --> {{ $ref}} 
			{{ else -}}
	{{ $out.Ref }} --> {{ $ref}}
			{{ end -}}
		{{ end -}}
	{{ end -}}
{{ end -}}
`))
