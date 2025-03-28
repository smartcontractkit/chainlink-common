package pkg

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/codegen"
)

type templateGenerator struct {
	Name             string
	Template         string
	FileNameTemplate string
}

func (t *templateGenerator) Generate(baseFile, args any) (string, string, error) {
	fileName, err := runTemplate(t.Name+"_fileName", t.FileNameTemplate, baseFile)
	if err != nil {
		return "", "", err
	}

	file, err := runTemplate(t.Name, t.Template, args)
	if err != nil {
		return fileName, "", err
	}

	settings := codegen.PrettySettings{
		Tool: "github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc",
		GoPrettySettings: codegen.GoPrettySettings{
			// TODO make this configurable
			LocalPrefix: "github.com/smartcontractkit",
		},
	}

	prettyFile, err := codegen.PrettyFile(fileName, file, settings)
	return fileName, prettyFile, err
}

func runTemplate(name, t string, args any) (string, error) {
	buf := &bytes.Buffer{}
	templ, err := template.New(name).
		Funcs(template.FuncMap{
			"LowerFirst": func(s string) string {
				if len(s) == 0 {
					return s
				}
				return strings.ToLower(s[:1]) + s[1:]
			},
		}).
		Parse(t)
	if err != nil {
		return "", err
	}

	err = templ.Execute(buf, args)
	return buf.String(), err
}
