package pkg

import (
	"bytes"
	"fmt"
	"path"
	"slices"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/codegen"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/tools/generator"
)

type TemplateGenerator struct {
	Name               string
	Template           string
	FileNameTemplate   string
	Partials           map[string]string
	StringLblValue     func(name string, label *generator.Label) (string, error)
	PbLabelTLangLabels func(labels map[string]*generator.Label) ([]Label, error)
	ExtraFns           template.FuncMap
}

func (t *TemplateGenerator) GenerateFile(
	file *protogen.File,
	plugin *protogen.Plugin,
	args any,
	toolName,
	localPrefix string) error {

	seen := map[string]int{}
	importToPkg := map[protogen.GoImportPath]protogen.GoPackageName{}
	for _, f := range plugin.Files {
		base := string(f.GoPackageName)
		alias := base
		if count, ok := seen[base]; ok {
			count++
			alias = fmt.Sprintf("%s%d", base, count)
			seen[base] = count
		} else {
			seen[base] = 0
		}
		importToPkg[f.GoImportPath] = protogen.GoPackageName(alias)
	}

	fileName, content, err := t.Generate(path.Base(file.GeneratedFilenamePrefix), args, importToPkg, toolName, localPrefix)
	if err != nil {
		return err
	}

	if len(content) > 0 {
		g := plugin.NewGeneratedFile(fileName, "")
		g.P(content)
	}
	return nil
}

func (t *TemplateGenerator) Generate(
	baseFile,
	args any,
	importToPkg map[protogen.GoImportPath]protogen.GoPackageName,
	toolName,
	localPrefix string) (string, string, error) {
	fileName, err := t.runTemplate(t.Name+"_fileName", t.FileNameTemplate, baseFile, t.Partials, importToPkg)
	if err != nil {
		return "", "", err
	}

	file, err := t.runTemplate(t.Name, t.Template, args, t.Partials, importToPkg)
	if err != nil {
		return fileName, "", err
	}

	if strings.TrimSpace(file) == "" {
		return fileName, "", nil
	}

	settings := codegen.PrettySettings{
		Tool: toolName,
		GoPrettySettings: codegen.GoPrettySettings{
			LocalPrefix: localPrefix,
		},
	}

	prettyFile, err := codegen.PrettyFile(fileName, file, settings)
	return fileName, prettyFile, err
}

func (t *TemplateGenerator) runTemplate(name, tmplText string, args any, partials map[string]string, importToPkg map[protogen.GoImportPath]protogen.GoPackageName) (string, error) {
	buf := &bytes.Buffer{}
	imports := map[string]bool{}
	var orderedImports []string
	if t.ExtraFns == nil {
		t.ExtraFns = template.FuncMap{}
	}

	templ := template.New(name).Funcs(template.FuncMap{
		"ImportAlias": func(importPath protogen.GoImportPath) string {
			return string(importToPkg[importPath])
		},
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
		"MapToUntypedAPI": func(m *protogen.Method) (bool, error) {
			md, err := getCapabilityMethodMetadata(m)
			if err != nil {
				return false, err
			}

			if md == nil {
				return false, nil
			} else {
				return md.MapToUntypedApi, nil
			}
		},
		"addImport": func(importPath protogen.GoImportPath, ignore string) string {
			importName := importPath.String()
			if ignore == importName {
				return ""
			}

			// add package name alias if path is mismatched with the package name
			if !isDirNamePackageName(importPath, importToPkg) {
				importName = fmt.Sprintf("%s %s", importToPkg[importPath], importName)
			}

			if !imports[importName] {
				orderedImports = append(orderedImports, importName)
				imports[importName] = true
			}

			return ""
		},
		"allimports": func() []string {
			allImports := make([]string, len(imports))
			copy(allImports, orderedImports)
			return allImports
		},
		"name": func(ident protogen.GoIdent, ignore string) string {
			importPath := ident.GoImportPath.String()
			if ignore == importPath {
				return ident.GoName
			}

			packageName := path.Base(strings.Trim(importPath, `"`))

			// use package name alias if package is mismatched with the package name
			if !isDirNamePackageName(ident.GoImportPath, importToPkg) {
				packageName = string(importToPkg[ident.GoImportPath])
			}

			return fmt.Sprintf("%s.%s", packageName, ident.GoName)
		},
		"CapabilityId": func(s *protogen.Service) (string, error) {
			md, err := getCapabilityMetadata(s)
			if err != nil {
				return "", err
			}
			return md.CapabilityId, nil
		},
		"FullCapabilityId": func(s *protogen.Service) (string, error) {
			md, err := getCapabilityMetadata(s)
			if err != nil {
				return "", err
			}

			if len(md.Labels) == 0 {
				return fmt.Sprintf(`"%s"`, md.CapabilityId), nil
			}

			// The format for labels is: "capabilityName + ':labelName:' + labelValue" for each label.
			// An example evm:ChainSelector:5009297550715157269@1.0.0 would be the EVM for Ethereum mainnet.

			orderedLabels := make([]*namedLabel, 0, len(md.Labels))
			for lblName, label := range md.Labels {
				orderedLabels = append(orderedLabels, &namedLabel{name: lblName, label: label})
			}
			slices.SortFunc(orderedLabels, func(a, b *namedLabel) int {
				return strings.Compare(a.name, b.name)
			})
			idParts := strings.Split(md.CapabilityId, "@")
			if len(idParts) != 2 {
				return "", fmt.Errorf("invalid capability ID format: %s", md.CapabilityId)
			}
			var fullName = fmt.Sprintf(`"%s"`, idParts[0])
			for _, lbl := range orderedLabels {
				lblValStr, err := t.StringLblValue(lbl.name, lbl.label)
				if err != nil {
					return "", fmt.Errorf("failed to stringify receving label %s: %w", lbl.name, err)
				}
				fullName = fmt.Sprintf(`%s + ":%s:" + %s`, fullName, lbl.name, lblValStr)
			}

			return fmt.Sprintf(`%s+"@%s"`, fullName, idParts[1]), nil
		},
		"Labels": func(s *protogen.Service) ([]Label, error) {
			md, err := getCapabilityMetadata(s)
			if err != nil {
				return nil, err
			}

			if len(md.Labels) == 0 {
				return nil, nil
			}

			return t.PbLabelTLangLabels(md.Labels)
		},
		"Mode": func(s *protogen.Service) (string, error) {
			md, err := getCapabilityMetadata(s)
			if err != nil {
				return "", err
			}

			switch md.Mode {
			case sdk.Mode_MODE_NODE:
				return "Node", nil
			case sdk.Mode_MODE_DON:
				return "", nil
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
		"CleanComments": func(line string) string {
			line = strings.TrimSpace(line)
			switch {
			case strings.HasPrefix(line, "//"):
				return strings.TrimSpace(strings.TrimPrefix(line, "//"))
			case strings.HasPrefix(line, "/*"):
				line = strings.TrimPrefix(line, "/*")
				line = strings.TrimSuffix(line, "*/")
				return strings.TrimSpace(line)
			default:
				return line
			}
		},
	}).Funcs(t.ExtraFns)

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

func isDirNamePackageName(importPath protogen.GoImportPath, importToPkg map[protogen.GoImportPath]protogen.GoPackageName) bool {
	packageName := importToPkg[importPath]
	dirName := path.Base(strings.Trim(importPath.String(), `"`))
	return dirName == string(packageName)
}

func getCapabilityMetadata(service *protogen.Service) (*generator.CapabilityMetadata, error) {
	opts := service.Desc.Options().(*descriptorpb.ServiceOptions)
	if proto.HasExtension(opts, generator.E_Capability) {
		ext := proto.GetExtension(opts, generator.E_Capability)
		if meta, ok := ext.(*generator.CapabilityMetadata); ok {
			return meta, nil
		}
		return nil, fmt.Errorf("invalid type for CapabilityMetadata")
	}
	return nil, nil
}

func getCapabilityMethodMetadata(m *protogen.Method) (*generator.CapabilityMethodMetadata, error) {
	opts := m.Desc.Options().(*descriptorpb.MethodOptions)
	if proto.HasExtension(opts, generator.E_Method) {
		ext := proto.GetExtension(opts, generator.E_Method)
		if meta, ok := ext.(*generator.CapabilityMethodMetadata); ok {
			return meta, nil
		}
		return nil, fmt.Errorf("invalid type for CapabilityMethodMetadata")
	}
	return nil, nil
}

type namedLabel struct {
	name  string
	label *generator.Label
}
