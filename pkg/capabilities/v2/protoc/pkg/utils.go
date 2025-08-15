package pkg

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"

	"github.com/smartcontractkit/chainlink-protos/cre/go/tools/generator"
)

func StringLblValue(method bool) func(string, *generator.Label) (string, error) {
	return func(name string, label *generator.Label) (string, error) {
		if method {
			name += "()"
		}
		switch pbLbl := label.Kind.(type) {
		case *generator.Label_StringLabel:
			return fmt.Sprintf("+ c.%s", name), nil
		case *generator.Label_Uint32Label:
			return fmt.Sprintf("strconv.FormatUint(uint64(c.%s), 10)", name), nil
		case *generator.Label_Int32Label:
			return fmt.Sprintf("strconv.FormatInt(int64(c.%s), 10)", name), nil
		case *generator.Label_Uint64Label:
			return fmt.Sprintf("strconv.FormatUint(c.%s, 10)", name), nil
		case *generator.Label_Int64Label:
			return fmt.Sprintf("strconv.FormatInt(c.%s, 10)", name), nil
		default:
			return "", fmt.Errorf("unsupported label type: %T", pbLbl)
		}
	}
}

func PbLabelToGoLabels(labels map[string]*generator.Label) ([]Label, error) {
	goLabels := make([]Label, 0, len(labels))
	for name, label := range labels {
		lbl := Label{Name: name}
		switch pbLbl := label.Kind.(type) {
		case *generator.Label_StringLabel:
			lbl.Type = "string"
			lbl.DefaultValues = mapDefaults(pbLbl.StringLabel.Defaults, true)
		case *generator.Label_Uint32Label:
			lbl.Type = "uint32"
			lbl.DefaultValues = mapDefaults(pbLbl.Uint32Label.Defaults, false)
		case *generator.Label_Uint64Label:
			lbl.Type = "uint64"
			lbl.DefaultValues = mapDefaults(pbLbl.Uint64Label.Defaults, false)
		case *generator.Label_Int64Label:
			lbl.Type = "int64"
			lbl.DefaultValues = mapDefaults(pbLbl.Int64Label.Defaults, false)
		default:
			return nil, fmt.Errorf("unsupported label type: %T", pbLbl)
		}
		goLabels = append(goLabels, lbl)
	}

	slices.SortFunc(goLabels, func(a, b Label) int { return strings.Compare(a.Name, b.Name) })

	return goLabels, nil
}

func mapDefaults[T any](defaults map[string]T, addQuotes bool) map[string]string {
	mapped := make(map[string]string, len(defaults))
	for k, v := range defaults {
		vStr := fmt.Sprintf("%v", v)
		if addQuotes {
			vStr = fmt.Sprintf("%q", vStr)
		}
		mapped[k] = vStr
	}
	return mapped
}
