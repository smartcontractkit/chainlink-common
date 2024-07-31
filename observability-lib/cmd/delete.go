package cmd

import (
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete Grafana Dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := cmd.Flag("dashboard-name").Value.String()
		url := cmd.Flag("grafana-url").Value.String()

		dashboard := NewDashboard(
			name,
			cmd.Flag("grafana-token").Value.String(),
			url,
			"",
			DataSources{},
			"",
			dashboard.Dashboard{},
		)

		err := dashboard.Delete()
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	DeleteCmd.Flags().String("dashboard-name", "", "Name of the dashboard to deploy")
	errName := DeleteCmd.MarkFlagRequired("dashboard-name")
	if errName != nil {
		panic(errName)
	}
	DeleteCmd.Flags().String("grafana-url", "", "Grafana URL")
	errURL := DeleteCmd.MarkFlagRequired("grafana-url")
	if errURL != nil {
		panic(errURL)
	}
	DeleteCmd.Flags().String("grafana-token", "", "Grafana API token")
	errToken := DeleteCmd.MarkFlagRequired("grafana-token")
	if errToken != nil {
		panic(errToken)
	}
}
