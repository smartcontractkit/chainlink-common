package notification_policy

import (
	"errors"
	"strings"

	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/smartcontractkit/chainlink-common/observability-lib/api"
	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [receiver]",
	Short: "Delete notification policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		grafanaClient := api.NewClient(
			cmd.Flag("grafana-url").Value.String(),
			cmd.Flag("grafana-token").Value.String(),
		)

		if len(args) != 1 {
			return errors.New("receiver argument missing")
		}

		matchers, err := cmd.Flags().GetStringArray("matchers")
		if err != nil {
			return err
		}
		if matchers != nil && len(matchers) > 0 {
			objectMatchers := alerting.ObjectMatchers{}
			notificationPolicy := alerting.NotificationPolicy{
				Receiver: grafana.Pointer(args[0]),
			}
			for _, matcher := range matchers {
				objectMatcher := strings.Split(matcher, ",")
				if len(objectMatcher) != 3 {
					return errors.New("invalid matcher format must be key,operator,value")
				}

				objectMatchers = append(objectMatchers, objectMatcher)
			}
			notificationPolicy.ObjectMatchers = &objectMatchers
			errDelete := grafanaClient.DeleteNestedPolicy(notificationPolicy)

			if errDelete != nil {
				return errDelete
			}
		}

		return nil
	},
}

func init() {
	deleteCmd.Flags().StringArray("matchers", []string{}, "Object matchers, in the form of key,operator,value e.g. 'key,=,value'")
	errMatchers := deleteCmd.MarkFlagRequired("matchers")
	if errMatchers != nil {
		panic(errMatchers)
	}
}
