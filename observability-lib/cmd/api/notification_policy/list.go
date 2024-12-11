package notification_policy

import (
	"github.com/grafana/grafana-foundation-sdk/go/alerting"
	"github.com/smartcontractkit/chainlink-common/observability-lib/api"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List notification policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		grafanaClient := api.NewClient(
			cmd.Flag("grafana-url").Value.String(),
			cmd.Flag("grafana-token").Value.String(),
		)

		notificationPolicyTree, _, err := grafanaClient.GetNotificationPolicy()
		if err != nil {
			return err
		}

		api.PrintPolicyTree(alerting.NotificationPolicy(notificationPolicyTree), 0)
		return nil
	},
}
