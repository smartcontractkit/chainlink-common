package cmd

import (
	"bytes"
	"embed"
	"io/fs"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

type TemplateWorkflowGeneratorHelper struct {
	Templates map[string]TemplateAndCondition
}

func (t *TemplateWorkflowGeneratorHelper) Generate(info GeneratedInfo) (map[string]string, error) {
	files := map[string]string{}
	if t.Templates == nil {
		return files, nil
	}

	for file, templateAndCondition := range t.Templates {
		if !templateAndCondition.ShouldGenerate(info) {
			continue
		}

		content, err := genFromTemplate(file, templateAndCondition, info)
		if err != nil {
			return nil, err
		}

		// can use a template, but it's simple for now
		fileName, err := genFromString("file name for "+file, file, info)
		if err != nil {
			return nil, err
		}
		files[fileName] = content
	}

	return files, nil
}

func genFromTemplate(name string, tc TemplateAndCondition, info GeneratedInfo) (string, error) {
	tmplFiles, err := getAllFiles(tc.FS())
	if err != nil {
		return "", err
	}

	t, err := template.New(name).Funcs(fns(info)).ParseFS(tc.FS(), tmplFiles...)

	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}
	err = t.ExecuteTemplate(buf, tc.Root(), info)
	return buf.String(), err
}

func genFromString(name string, rawTemplate string, info GeneratedInfo) (string, error) {
	t, err := template.New(name).Funcs(fns(info)).Parse(rawTemplate)
	if err != nil {
		return "", err
	}
	buf := &bytes.Buffer{}
	err = t.Execute(buf, info)
	return buf.String(), err
}

func fns(info GeneratedInfo) template.FuncMap {
	return template.FuncMap{
		"LowerFirst": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToLower(s[:1]) + s[1:]
		},
		"Capitalize": capitalize,
		"CapitalizeCap": func(c capabilities.CapabilityType) string {
			return capitalize(string(c))
		},
		"ToSnake": strcase.ToSnake,
		"Repeat":  strings.Repeat,
		"InputAfterCapability": func() string {
			return info.BaseName + "Input"
		},
		"HasOutputs": func(tpe string) bool {
			return len(info.Types[tpe].Outputs) > 0
		},
		"IsCommon": func(tpe capabilities.CapabilityType) bool {
			return tpe.IsValid() != nil
		},
	}
}

func getAllFiles(templateFs embed.FS) ([]string, error) {
	var tmplFiles []string

	// Walk the embedded filesystem to collect all template files
	err := fs.WalkDir(templateFs, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if it's not a directory and has .tmpl extension
		if !d.IsDir() {
			tmplFiles = append(tmplFiles, path)
		}
		return nil
	})

	return tmplFiles, err
}
