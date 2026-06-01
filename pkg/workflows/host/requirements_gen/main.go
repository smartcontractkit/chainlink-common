package main

import (
	"bytes"
	_ "embed"
	"log"
	"os"
	"reflect"
	"text/template"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/codegen"
	sdk "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

//go:embed requirements_helper.go.tmpl
var tmplSrc string

type fieldInfo struct {
	Name string
	Type string
}

type templateData struct {
	Fields []fieldInfo
}

func main() {
	requirementsType := reflect.TypeOf(sdk.Requirements{})

	var fields []fieldInfo
	for i := 0; i < requirementsType.NumField(); i++ {
		f := requirementsType.Field(i)
		if !f.IsExported() {
			continue
		}
		fields = append(fields, fieldInfo{
			Name: f.Name,
			Type: f.Type.String(),
		})
	}

	tmpl, err := template.New("requirements_helper").Parse(tmplSrc)
	if err != nil {
		log.Fatalf("failed to parse template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData{Fields: fields}); err != nil {
		log.Fatalf("failed to execute template: %v", err)
	}

	const outFile = "requirements_helper_gen.go"
	settings := codegen.PrettySettings{
		Tool: "requirements_gen",
		GoPrettySettings: codegen.GoPrettySettings{
			LocalPrefix: "github.com/smartcontractkit/chainlink-common",
		},
	}

	content, err := codegen.PrettyFile(outFile, buf.String(), settings)
	if err != nil {
		log.Fatalf("failed to format generated code: %v\n%s", err, buf.String())
	}

	if err := os.WriteFile(outFile, []byte(content), 0644); err != nil {
		log.Fatalf("failed to write output: %v", err)
	}
}
