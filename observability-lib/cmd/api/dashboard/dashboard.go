package dashboard

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "dashboard [actions]",
	Short: "Perform actions on dashboard",
}

func init() {
	Cmd.AddCommand(deleteCmd)
}
