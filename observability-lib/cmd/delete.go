package cmd

import (
	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete Grafana Dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		return DeleteDashboard(&CommandOptions{
			GrafanaURL:   cmd.Flag("grafana-url").Value.String(),
			GrafanaToken: cmd.Flag("grafana-token").Value.String(),
			Name:         cmd.Flag("dashboard-name").Value.String(),
		})
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
