package workflows

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strings"
	"text/template"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
)

func (g *DependencyGraph) FormatChart() (string, error) {
	var sb strings.Builder
	steps := slices.Clone(g.Spec.Triggers)
	steps = append(steps, g.Spec.Steps()...)
	slices.SortFunc(steps, func(a, b sdk.StepDefinition) int {
		return cmp.Or(
			a.CapabilityType.Compare(b.CapabilityType),
			cmp.Compare(a.Ref, b.Ref),
			cmp.Compare(a.ID, b.ID),
		)
	})
	preds, err := g.Graph.PredecessorMap()
	if err != nil {
		return "", fmt.Errorf("failed to get graph predecessors: %w", err)
	}
	type stepAndOutput struct {
		Step   sdk.StepDefinition
		Inputs []string
	}
	nodes := make([]stepAndOutput, len(steps))
	for i, step := range steps {
		inputs := slices.Collect(maps.Keys(preds[step.Ref]))
		if step.CapabilityType != capabilities.CapabilityTypeTrigger {
			inputs = append(inputs, KeywordTrigger)
		}
		slices.Sort(inputs)
		nodes[i] = stepAndOutput{Step: step, Inputs: inputs}
	}
	err = tmpl.Execute(&sb, nodes)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

var tmpl = template.Must(template.New("").Funcs(map[string]any{
	"replace": strings.ReplaceAll,
}).Parse(`flowchart
{{ range $i, $e := . }}
	{{ $ref := .Step.Ref -}}
	{{ $id := replace .Step.ID "@" "[at]" -}}
	{{ $name := printf "%s<br><i>(%s)</i>" .Step.CapabilityType $id -}}
	{{ if .Step.Ref -}}
		{{ $name = printf "<b>%s</b><br>%s" .Step.Ref $name -}}
	{{ else -}}
		{{ $ref = printf "%s%d" "unnamed" $i -}}
	{{ end -}}
	{{ if eq .Step.CapabilityType "trigger" -}}
	{{ $ref }}[\"{{$name}}"/]
	{{ else if eq .Step.CapabilityType "consensus" -}}
	{{ $ref }}[["{{$name}}"]]
	{{ else if eq .Step.CapabilityType "target" -}}
	{{ $ref }}[/"{{$name}}"\]
	{{ else -}}
	{{ $ref }}["{{$name}}"]
	{{ end -}}
	{{ if .Step.Inputs.OutputRef -}}
	{{ .Step.Inputs.OutputRef }} --> {{ .Step.Ref }}
	{{ else -}}
		{{ range $out := .Inputs -}}
	{{ $out }} --> {{ $ref }}
		{{ end -}}
	{{ end -}}
{{ end -}}
`))
