package contact_point

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "contact-point [actions]",
	Short: "Perform actions on contact point",
}

func init() {
	Cmd.AddCommand(listCmd, deleteCmd)
}
