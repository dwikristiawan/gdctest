package cmd

import (
	_ "gdctest/app/modules"
	"gdctest/internal/server"

	"github.com/spf13/cobra"
)

var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "Start REST API server",
	Run: func(cmd *cobra.Command, args []string) {
		server.Run(cfg)
	},
}

func init() {
	rootCmd.AddCommand(restCmd)
}
