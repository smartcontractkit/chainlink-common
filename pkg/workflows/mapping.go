package workflows

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/shopspring/decimal"
)

type mapping map[string]any

func (m *mapping) UnmarshalJSON(b []byte) error {
	mp := map[string]any{}

	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()

	err := d.Decode(&mp)
	if err != nil {
		return err
	}

	nm, err := convertNumbers(mp)
	if err != nil {
		return err
	}

	*m = (mapping)(nm)
	return err
}

// convertNumber detects if a json.Number is an integer or a decimal and converts it to the appropriate type.
//
// Supported type conversions:
// - json.Number -> int64
// - json.Number -> float64 -> decimal.Decimal
func convertNumber(el any) (any, error) {
	switch elv := el.(type) {
	case json.Number:
		if strings.Contains(elv.String(), ".") {
			f, err := elv.Float64()
			if err == nil {
				return decimal.NewFromFloat(f), nil
			}
		}

		return elv.Int64()
	default:
		return el, nil
	}
}

func convertNumbers(m map[string]any) (map[string]any, error) {
	nm := map[string]any{}
	for k, v := range m {
		switch tv := v.(type) {
		case map[string]any:
			cm, err := convertNumbers(tv)
			if err != nil {
				return nil, err
			}

			nm[k] = cm
		case []any:
			na := make([]any, len(tv))
			for i, v := range tv {
				cv, err := convertNumber(v)
				if err != nil {
					return nil, err
				}

				na[i] = cv
			}

			nm[k] = na
		default:
			cv, err := convertNumber(v)
			if err != nil {
				return nil, err
			}

			nm[k] = cv
		}
	}

	return nm, nil
}

func (m mapping) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any(m))
}
