package main

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/scottstav/wreccless/internal/config"
	"github.com/scottstav/wreccless/internal/hooks"
	"github.com/scottstav/wreccless/internal/state"
	"github.com/spf13/cobra"
)

var approveCmd = &cobra.Command{
	Use:   "approve <id>",
	Short: "Start a pending worker",
	Args:  cobra.ExactArgs(1),
	RunE:  runApprove,
}

func init() {
	rootCmd.AddCommand(approveCmd)
}

func runApprove(cmd *cobra.Command, args []string) error {
	id := args[0]
	w, err := state.Read(stateDir, id)
	if err != nil {
		return fmt.Errorf("worker %s not found", id)
	}
	if w.Status != state.StatusPending {
		return fmt.Errorf("worker %s is %s, not pending", id, w.Status)
	}

	cfg, _ := config.Load(configPath)

	now := time.Now()
	w.Status = state.StatusWorking
	w.StartedAt = &now
	if w.SessionID == "" {
		w.SessionID = uuid.New().String()
	}
	if err := state.Write(stateDir, w); err != nil {
		return err
	}

	// TODO: Task 11 — spawn detached ccl run here
	vars := hooks.Vars{ID: id, Task: w.Task, Dir: w.Directory, Status: string(w.Status)}
	hooks.Fire(cfg.Hooks.OnStart, vars)

	fmt.Fprintf(cmd.OutOrStdout(), "Approved worker %s\n", id)
	return nil
}
