package contact_point

import (
	"github.com/smartcontractkit/chainlink-common/observability-lib/api"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List contact point",
	RunE: func(cmd *cobra.Command, args []string) error {
		grafanaClient := api.NewClient(
			cmd.Flag("grafana-url").Value.String(),
			cmd.Flag("grafana-token").Value.String(),
		)

		contactPoints, _, err := grafanaClient.GetContactPoints()
		if err != nil {
			return err
		}

		for _, contactPoint := range contactPoints {
			cmd.Printf("| Name: %s | UID: %s\n", *contactPoint.Name, *contactPoint.Uid)
		}

		return nil
	},
}
