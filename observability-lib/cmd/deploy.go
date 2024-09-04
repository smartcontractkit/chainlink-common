package cmd

import (
	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy Grafana dashboard and associated alerts",
	RunE: func(cmd *cobra.Command, args []string) error {
		alertsTags, errAlertsTags := cmd.Flags().GetStringToString("alerts-tags")
		if errAlertsTags != nil {
			return errAlertsTags
		}

		err := NewDashboard(&CommandOptions{
			GrafanaURL:            cmd.Flag("grafana-url").Value.String(),
			GrafanaToken:          cmd.Flag("grafana-token").Value.String(),
			Name:                  cmd.Flag("dashboard-name").Value.String(),
			FolderName:            cmd.Flag("dashboard-folder").Value.String(),
			Platform:              cmd.Flag("platform").Value.String(),
			TypeDashboard:         cmd.Flag("type").Value.String(),
			MetricsDataSourceName: cmd.Flag("metrics-datasource").Value.String(),
			LogsDataSourceName:    cmd.Flag("logs-datasource").Value.String(),
			EnableAlerts:          cmd.Flag("enable-alerts").Value.String() == "true",
			AlertsTags:            alertsTags,
			NotificationTemplates: cmd.Flag("notification-templates").Value.String(),
			SlackChannel:          cmd.Flag("slack-channel").Value.String(),
			SlackWebhookURL:       cmd.Flag("slack-webhook").Value.String(),
			SlackToken:            cmd.Flag("slack-token").Value.String(),
		})

		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	DeployCmd.Flags().String("dashboard-name", "", "Name of the dashboard to deploy")
	errName := DeployCmd.MarkFlagRequired("dashboard-name")
	if errName != nil {
		panic(errName)
	}
	DeployCmd.Flags().String("dashboard-folder", "", "Dashboard folder")
	errFolder := DeployCmd.MarkFlagRequired("dashboard-folder")
	if errFolder != nil {
		panic(errFolder)
	}
	DeployCmd.Flags().String("grafana-url", "", "Grafana URL")
	errURL := DeployCmd.MarkFlagRequired("grafana-url")
	if errURL != nil {
		panic(errURL)
	}
	DeployCmd.Flags().String("grafana-token", "", "Grafana API token")
	errToken := DeployCmd.MarkFlagRequired("grafana-token")
	if errToken != nil {
		panic(errToken)
	}
	DeployCmd.Flags().String("metrics-datasource", "Prometheus", "Metrics datasource name")
	DeployCmd.Flags().String("logs-datasource", "", "Logs datasource name")
	DeployCmd.Flags().String("platform", "docker", "Platform where the dashboard is deployed (docker or kubernetes)")
	DeployCmd.Flags().String("type", "core-node", "Dashboard type can be either core-node | core-node-components | core-node-resources | don-ocr | don-ocr2 | don-ocr3 | nop-ocr2 | nop-ocr3")
	DeployCmd.Flags().Bool("enable-alerts", false, "Deploy alerts")
	DeployCmd.Flags().StringToString("alerts-tags", map[string]string{
		"team": "chainlink-team",
	}, "Alerts tags")
	DeployCmd.Flags().String("notification-templates", "", "Filepath in yaml format, will create notification templates depending on key-value pairs in the yaml file")
	DeployCmd.Flags().String("slack-channel", "", "Slack channel, required when setting up slack contact points")
	DeployCmd.Flags().String("slack-webhook", "", "Slack webhook URL, required when setting up slack contact points")
	DeployCmd.Flags().String("slack-token", "", "Slack token, required when setting up slack contact points and slack webhook is not provided")
}
