package main

import (
	"encoding/json"
	"fmt"

	"github.com/scottstav/wreccless/internal/state"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <id>",
	Short: "Show detailed status of a worker",
	Args:  cobra.ExactArgs(1),
	RunE:  runStatus,
}

var statusJSON bool

func init() {
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "Output JSON")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	w, err := state.Read(stateDir, args[0])
	if err != nil {
		return fmt.Errorf("worker %s not found", args[0])
	}

	if statusJSON {
		data, _ := json.MarshalIndent(w, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "ID:         %s\n", w.ID)
	fmt.Fprintf(cmd.OutOrStdout(), "Status:     %s\n", w.Status)
	fmt.Fprintf(cmd.OutOrStdout(), "Directory:  %s\n", w.Directory)
	fmt.Fprintf(cmd.OutOrStdout(), "Task:       %s\n", w.Task)
	if w.Image != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Image:      %s\n", w.Image)
	}
	if w.PID > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "PID:        %d\n", w.PID)
	}
	if w.SessionID != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Session:    %s\n", w.SessionID)
	}
	if w.CreatedAt != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Created:    %s\n", w.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	if w.StartedAt != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Started:    %s\n", w.StartedAt.Format("2006-01-02 15:04:05"))
	}
	if w.FinishedAt != nil {
		fmt.Fprintf(cmd.OutOrStdout(), "Finished:   %s\n", w.FinishedAt.Format("2006-01-02 15:04:05"))
	}
	return nil
}
