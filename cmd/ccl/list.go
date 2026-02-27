package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/scottstav/wreccless/internal/state"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workers",
	RunE:  runList,
}

var (
	listJSON   bool
	listStatus string
)

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output JSON")
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (pending|working|done|error)")
	rootCmd.AddCommand(listCmd)
}

func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

func runList(cmd *cobra.Command, args []string) error {
	workers, err := state.List(stateDir)
	if err != nil {
		return err
	}

	// Stale detection
	for _, w := range workers {
		if w.Status == state.StatusWorking && w.PID > 0 && !isProcessAlive(w.PID) {
			w.Status = state.StatusError
			state.Write(stateDir, w)
		}
	}

	if listStatus != "" {
		var filtered []*state.Worker
		for _, w := range workers {
			if string(w.Status) == listStatus {
				filtered = append(filtered, w)
			}
		}
		workers = filtered
	}

	if len(workers) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No workers.")
		return nil
	}

	if listJSON {
		data, err := json.Marshal(workers)
		if err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}

	tw := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tSTATUS\tDIRECTORY\tTASK")
	home, _ := os.UserHomeDir()
	for _, w := range workers {
		dir := w.Directory
		if home != "" {
			dir = strings.Replace(dir, home, "~", 1)
		}
		task := w.Task
		if len(task) > 60 {
			task = task[:57] + "..."
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", w.ID, w.Status, dir, task)
	}
	tw.Flush()
	return nil
}
