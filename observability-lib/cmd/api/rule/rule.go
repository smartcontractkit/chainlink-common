package rule

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "rule [actions]",
	Short: "Perform actions on dashboard",
}

func init() {
	Cmd.AddCommand(deleteCmd)
}
