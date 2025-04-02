package grafana

import "github.com/grafana/grafana-foundation-sdk/go/alerting"

type ContactPointOptions struct {
	Name                  string
	Type                  alerting.ContactPointType
	Settings              map[string]interface{}
	DisableResolveMessage bool
	Uid                   string
}

func NewContactPoint(options *ContactPointOptions) *alerting.ContactPointBuilder {
	builder := alerting.NewContactPointBuilder().
		Name(options.Name).
		Type(options.Type).
		Settings(options.Settings).
		DisableResolveMessage(options.DisableResolveMessage)

	if options.Uid != "" {
		builder.Uid(options.Uid)
	}

	return builder
}
