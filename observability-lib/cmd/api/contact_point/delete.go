package contact_point

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/observability-lib/api"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete contact point by name",
	RunE: func(cmd *cobra.Command, args []string) error {
		grafanaClient := api.NewClient(
			cmd.Flag("grafana-url").Value.String(),
			cmd.Flag("grafana-token").Value.String(),
		)

		contactPoint, err := grafanaClient.GetContactPointByName(args[0])
		if err != nil {
			return err
		}

		if contactPoint == nil {
			return errors.New("contact point not found")
		}

		_, _, errDelete := grafanaClient.DeleteContactPoint(*contactPoint.Uid)
		if errDelete != nil {
			return errDelete
		}

		return nil
	},
}
