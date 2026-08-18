package rule

import (
	"github.com/smartcontractkit/chainlink-common/observability-lib/api"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete rule by UID",
	RunE: func(cmd *cobra.Command, args []string) error {
		grafanaClient := api.NewClient(
			cmd.Flag("grafana-url").Value.String(),
			cmd.Flag("grafana-token").Value.String(),
		)

		_, _, errDelete := grafanaClient.DeleteAlertRule(args[0])
		if errDelete != nil {
			return errDelete
		}

		return nil
	},
}
