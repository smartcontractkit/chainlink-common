package cmd

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-common/observability-lib/grafana"
	"github.com/spf13/cobra"
)

var GenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Grafana Dashboard JSON",
	RunE: func(cmd *cobra.Command, args []string) error {
		dashboard, err := BuildDashboardWithType(&BuildOptions{
			Name:              cmd.Flag("dashboard-name").Value.String(),
			Platform:          grafana.TypePlatform(cmd.Flag("platform").Value.String()),
			TypeDashboard:     TypeDashboard(cmd.Flag("type").Value.String()),
			MetricsDataSource: grafana.NewDataSource(cmd.Flag("metrics-datasource").Value.String(), ""),
			LogsDataSource:    grafana.NewDataSource(cmd.Flag("logs-datasource").Value.String(), ""),
		})

		if err != nil {
			return err
		}

		dashboardJSON, errDashboardJSON := dashboard.GenerateJSON()
		if errDashboardJSON != nil {
			return err
		}

		fmt.Print(string(dashboardJSON))

		return nil
	},
}

func init() {
	GenerateCmd.Flags().String("dashboard-name", "", "Name of the dashboard to deploy")
	errName := GenerateCmd.MarkFlagRequired("dashboard-name")
	if errName != nil {
		panic(errName)
	}
	GenerateCmd.Flags().String("metrics-datasource", "Prometheus", "Metrics datasource name")
	GenerateCmd.Flags().String("logs-datasource", "", "Logs datasource name")
	GenerateCmd.Flags().String("platform", "docker", "Platform where the dashboard is deployed (docker or kubernetes)")
	GenerateCmd.Flags().String("type", "core-node", "Dashboard type can be either core-node | core-node-components | core-node-resources | don-ocr | don-ocr2 | don-ocr3")
}
