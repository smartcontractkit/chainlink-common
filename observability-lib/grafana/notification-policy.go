package grafana

import "github.com/grafana/grafana-foundation-sdk/go/alerting"

type NotificationPolicyOptions struct {
	Receiver string // must match name of ContactPoint Name
	Matchers alerting.Matchers
}

func NewNotificationPolicy(options *NotificationPolicyOptions) *alerting.NotificationPolicyBuilder {
	return alerting.NewNotificationPolicyBuilder().
		Receiver(options.Receiver).
		Matchers(options.Matchers)
}
