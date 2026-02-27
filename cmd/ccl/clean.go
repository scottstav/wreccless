package main

import (
	"fmt"

	"github.com/scottstav/wreccless/internal/state"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove done/error workers",
	RunE:  runClean,
}

var cleanAll bool

func init() {
	cleanCmd.Flags().BoolVar(&cleanAll, "all", false, "Remove all workers including working")
	rootCmd.AddCommand(cleanCmd)
}

func runClean(cmd *cobra.Command, args []string) error {
	workers, err := state.List(stateDir)
	if err != nil {
		return err
	}

	removed := 0
	for _, w := range workers {
		if cleanAll || w.Status == state.StatusDone || w.Status == state.StatusError {
			state.Delete(stateDir, w.ID)
			removed++
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Cleaned %d worker(s).\n", removed)
	return nil
}
