package dashboard

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/observability-lib/api"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [name] [folderUID]",
	Short: "Delete dashboard by name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return errors.New("please provide dashboard name and folder UID")
		}

		grafanaClient := api.NewClient(
			cmd.Flag("grafana-url").Value.String(),
			cmd.Flag("grafana-token").Value.String(),
		)
		delDashboard, _, err := grafanaClient.GetDashboardByNameFolderUID(args[0], args[1])
		if err != nil {
			return err
		}

		if delDashboard.UID == nil {
			return errors.New("contact point not found")
		}

		_, errDelete := grafanaClient.DeleteDashboardByUID(*delDashboard.UID)
		if errDelete != nil {
			return errDelete
		}

		return nil
	},
}
