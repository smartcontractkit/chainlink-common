package pkg

import (
	"errors"
	"fmt"
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type Label struct {
	Name   string
	Type   string
	Values map[string]string
}

func PbLabelToGoLabels(labels map[string]*pb.Label) ([]Label, error) {
	goLabels := make([]Label, 0, len(labels))
	for name, label := range labels {
		lbl := Label{Name: name}
		switch pbLbl := label.Kind.(type) {
		case *pb.Label_StringLabel:
			lbl.Type = "string"
			lbl.Values = mapDefaults(pbLbl.StringLabel.Defaults, true)
		case *pb.Label_Uint32Label:
			lbl.Type = "uint32"
			lbl.Values = mapDefaults(pbLbl.Uint32Label.Defaults, false)
		case *pb.Label_Uint64Label:
			lbl.Type = "uint64"
			lbl.Values = mapDefaults(pbLbl.Uint64Label.Defaults, false)
		case *pb.Label_Int64Label:
			lbl.Type = "int64"
			lbl.Values = mapDefaults(pbLbl.Int64Label.Defaults, false)
		default:
			return nil, fmt.Errorf("unsupported label type: %T", pbLbl)
		}
		goLabels = append(goLabels, lbl)
	}

	return goLabels, nil
}

func AppendGoLabelsToVersion(lbls []Label) (string, error) {
	if len(lbls) == 0 {
		return "", nil
	}

	entries := make([]string, len(lbls))
	for i, lbl := range lbls {
		switch lbl.Type {
		case "string":
			entries[i] = lbl.Name + "=\"+c." + lbl.Name + "()"
		case "uint64":
			entries[i] = lbl.Name + "=\"+strconv.FormatUint(c." + lbl.Name + "(), 10)"
		case "int64":
			entries[i] = lbl.Name + "=\"+strconv.FormatInt(c." + lbl.Name + "(), 10)"
		case "uint32":
			entries[i] = lbl.Name + "=\"+strconv.FormatUint(uint64(c." + lbl.Name + "()), 10)"
		case "int32":
			entries[i] = lbl.Name + "=\"+strconv.FormatInt(int64(c." + lbl.Name + "()), 10)"
		default:
			return "", errors.New("unsupported label type: " + lbl.Type)
		}
	}

	return "+\":" + strings.Join(entries, "&"), nil
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
