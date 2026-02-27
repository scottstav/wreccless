package main

import (
	"github.com/scottstav/wreccless/internal/config"
	"github.com/scottstav/wreccless/internal/worker"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:    "run <id>",
	Short:  "Internal: run a worker's claude session (blocking)",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	RunE:   runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	return worker.Run(stateDir, args[0], cfg, "")
}
