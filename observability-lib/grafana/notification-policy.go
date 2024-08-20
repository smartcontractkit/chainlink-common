package grafana

import (
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/grafana/grafana-foundation-sdk/go/cog"
)

type NotificationPolicyOptions struct {
	Continue          *bool
	GroupBy           []string
	GroupInterval     string
	GroupWait         string
	MuteTimeIntervals []string
	Receiver          string // must match name of ContactPoint Name
	ObjectMatchers    []alerting.ObjectMatcher
	Provenance        *alerting.Provenance
	RepeatInterval    string
	Routes            []*alerting.NotificationPolicyBuilder
}

func NewNotificationPolicy(options *NotificationPolicyOptions) *alerting.NotificationPolicyBuilder {
	routes := make([]cog.Builder[alerting.NotificationPolicy], 0)
	for _, route := range options.Routes {
		routes = append(routes, route)
	}

	newNotificationPolicy := alerting.NewNotificationPolicyBuilder().
		GroupBy(options.GroupBy).
		MuteTimeIntervals(options.MuteTimeIntervals).
		Routes(routes)

	if options.Continue != nil {
		newNotificationPolicy.Continue(*options.Continue)
	}

	if options.GroupInterval != "" {
		newNotificationPolicy.GroupInterval(options.GroupInterval)
	}

	if options.GroupWait != "" {
		newNotificationPolicy.GroupWait(options.GroupWait)
	}

	if options.ObjectMatchers != nil {
		newNotificationPolicy.ObjectMatchers(options.ObjectMatchers)
	}

	if options.Provenance != nil {
		newNotificationPolicy.Provenance(*options.Provenance)
	}

	if options.Receiver != "" {
		newNotificationPolicy.Receiver(options.Receiver)
	}

	if options.RepeatInterval != "" {
		newNotificationPolicy.RepeatInterval(options.RepeatInterval)
	}

	return newNotificationPolicy
}
