package pkg

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/codegen"
)

type templateGenerator struct {
	Name             string
	Template         string
	FileNameTemplate string
	Partials         map[string]string
}

func (t *templateGenerator) Generate(baseFile, args any) (string, string, error) {
	fileName, err := runTemplate(t.Name+"_fileName", t.FileNameTemplate, baseFile, t.Partials)
	if err != nil {
		return "", "", err
	}

	file, err := runTemplate(t.Name, t.Template, args, t.Partials)
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

func runTemplate(name, tmplText string, args any, partials map[string]string) (string, error) {
	buf := &bytes.Buffer{}
	imports := map[string]bool{}
	templ := template.New(name).Funcs(template.FuncMap{
		"LowerFirst": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToLower(s[:1]) + s[1:]
		},
		"dict": func(values ...any) (map[string]any, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict requires even number of arguments")
			}
			m := make(map[string]any, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings, got %T", values[i])
				}
				m[key] = values[i+1]
			}
			return m, nil
		},
		"isTrigger": isTrigger,
		"addImport": func(name protogen.GoImportPath, ignore string) string {
			if ignore != name.String() {
				imports[name.String()] = true
			}

			return ""
		},
		"allimports": func() []string {
			var allImports []string
			for i := range imports {
				allImports = append(allImports, i)
			}
			return allImports
		},
		"name": func(ident protogen.GoIdent, ignore string) string {
			importPath := ident.GoImportPath.String()
			if ignore == importPath {
				return ident.GoName
			}

			// remove quotes
			importPath = importPath[1 : len(importPath)-1]
			parts := strings.Split(importPath, "/")
			return fmt.Sprintf("%s.%s", parts[len(parts)-1], ident.GoName)
		},
	})

	// Register partials
	if partials != nil {
		for name, pt := range partials {
			_, err := templ.New(name).Parse(pt)
			if err != nil {
				return "", err
			}
		}
	}

	// Parse the main template
	templ, err := templ.Parse(tmplText)
	if err != nil {
		return "", err
	}

	err = templ.Execute(buf, args)
	return buf.String(), err
}

func isTrigger(method *protogen.Method) bool {
	opts := method.Desc.Options().(*descriptorpb.MethodOptions)
	if opts == nil {
		return false
	}

	callType := proto.GetExtension(opts, pb.E_CallType).(pb.CallType)

	return callType == pb.CallType_Trigger
}
