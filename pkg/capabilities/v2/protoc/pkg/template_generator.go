package pkg

import (
	"bytes"
	"fmt"
	"path"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/codegen"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type templateGenerator struct {
	Name             string
	Template         string
	FileNameTemplate string
	Partials         map[string]string
}

func (t *templateGenerator) GenerateFile(file *protogen.File, plugin *protogen.Plugin, args any) error {
	fileName, content, err := t.Generate(path.Base(file.GeneratedFilenamePrefix), args)
	if err != nil {
		return err
	}

	g := plugin.NewGeneratedFile(fileName, "")
	g.P(content)
	return nil
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
		Tool: "github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc",
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
		"ToLower": func(s string) string {
			return strings.ToLower(s)
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
		"isTrigger": func(m *protogen.Method) bool { return m.Desc.IsStreamingServer() },
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
		"CapabilityId": func(s *protogen.Service) (string, error) {
			// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-797 ID should be allowed to require a parameter.
			md, err := getCapabilityMetadata(s)
			if err != nil {
				return "", err
			}
			return md.CapabilityId, nil
		},
		"Mode": func(s *protogen.Service) (string, error) {
			md, err := getCapabilityMetadata(s)
			if err != nil {
				return "", err
			}

			switch md.Mode {
			case sdkpb.Mode_Node:
				return "Node", nil
			case sdkpb.Mode_DON:
				return "Don", nil
			default:
				return "", fmt.Errorf("unsupported mode: %s", md.Mode)
			}
		},
		"ConfigType": func(s *protogen.Service) (string, error) {
			md, err := getCapabilityMetadata(s)
			if err != nil {
				return "", err
			}
			_ = md

			return "emptypb.Empty", nil
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

func getCapabilityMetadata(service *protogen.Service) (*pb.CapabilityMetadata, error) {
	opts := service.Desc.Options().(*descriptorpb.ServiceOptions)
	if proto.HasExtension(opts, pb.E_Capability) {
		ext := proto.GetExtension(opts, pb.E_Capability)
		if meta, ok := ext.(*pb.CapabilityMetadata); ok {
			return meta, nil
		}
		return nil, fmt.Errorf("invalid type for CapabilityMetadata")
	}
	return nil, nil
}
