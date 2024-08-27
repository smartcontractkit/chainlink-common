package grafana

import "github.com/grafana/grafana-foundation-sdk/go/alerting"

type ContactPointOptions struct {
	Name     string
	Type     alerting.ContactPointType
	Settings map[string]interface{}
}

func NewContactPoint(options *ContactPointOptions) *alerting.ContactPointBuilder {
	return alerting.NewContactPointBuilder().
		Name(options.Name).
		Type(options.Type).
		Settings(options.Settings)
}
