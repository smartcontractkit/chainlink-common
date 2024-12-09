package notification_policy

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "notification-policy [actions]",
	Short: "Perform actions on notification policy",
}

func init() {
	Cmd.AddCommand(listCmd, deleteCmd)
}
