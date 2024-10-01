package workflows

import (
	"cmp"
	"slices"
	"strings"
	"text/template"

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
	err := tmpl.Execute(&sb, steps)
	if err != nil {
		return "", err
	}
	return sb.String(), nil
}

var tmpl = template.Must(template.New("").Funcs(map[string]any{
	"replace": strings.ReplaceAll,
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
