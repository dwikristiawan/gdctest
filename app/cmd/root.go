package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"gdctest/app/config"
	"gdctest/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	EnvFilePath string
	cfg         *config.Config

	rootCmd = &cobra.Command{
		Use:   "gdctest",
		Short: "GDC Test Backend Service",
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		slog.Error("execute error", "err", err.Error())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&EnvFilePath, "env", "e", ".env", ".env file path")
}

func initConfig() {
	logger.Init()

	var err error
	cfg, err = config.New(EnvFilePath)
	if err != nil {
		panic(fmt.Errorf("config error: %w", err))
	}
}
