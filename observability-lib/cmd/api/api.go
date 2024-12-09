package api

import (
	"github.com/smartcontractkit/chainlink-common/observability-lib/cmd/api/contact_point"
	"github.com/smartcontractkit/chainlink-common/observability-lib/cmd/api/dashboard"
	"github.com/smartcontractkit/chainlink-common/observability-lib/cmd/api/notification_policy"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "api [resources]",
	Short: "Select resources to perform actions",
}

func init() {
	Cmd.AddCommand(contact_point.Cmd)
	Cmd.AddCommand(dashboard.Cmd)
	Cmd.AddCommand(notification_policy.Cmd)

	Cmd.PersistentFlags().String("grafana-url", "", "Grafana URL")
	errURL := Cmd.MarkPersistentFlagRequired("grafana-url")
	if errURL != nil {
		panic(errURL)
	}

	Cmd.PersistentFlags().String("grafana-token", "", "Grafana API token")
	errToken := Cmd.MarkPersistentFlagRequired("grafana-token")
	if errToken != nil {
		panic(errToken)
	}
}
