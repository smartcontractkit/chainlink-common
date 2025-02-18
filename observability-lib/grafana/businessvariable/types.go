package businessvariable

import (
	"encoding/json"

	"github.com/grafana/grafana-foundation-sdk/go/cog/variants"
)

type DisplayMode string

const (
	DisplayModeTable    DisplayMode = "table"
	DisplayModeMinimize DisplayMode = "minimize"
)

type Options struct {
	DisplayMode DisplayMode `json:"displayMode"`
	Padding     int         `json:"padding"`
	ShowLabel   bool        `json:"showLabel"`
	Variable    string      `json:"variable"`
}

func VariantConfig() variants.PanelcfgConfig {
	return variants.PanelcfgConfig{
		Identifier: "volkovlabs-variable-panel",
		OptionsUnmarshaler: func(raw []byte) (any, error) {
			options := Options{}

			if err := json.Unmarshal(raw, &options); err != nil {
				return nil, err
			}

			return options, nil
		},
	}
}
