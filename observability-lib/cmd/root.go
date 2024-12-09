package cmd

import (
	"log"

	"github.com/smartcontractkit/chainlink-common/observability-lib/cmd/api"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "observability-lib [command]",
	Short: "observability-lib CLI to perform actions on observability resources",
}

func init() {
	rootCmd.AddCommand(api.Cmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalln(err)
	}
}
