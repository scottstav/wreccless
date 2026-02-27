package main

import (
	"fmt"

	"github.com/scottstav/wreccless/internal/config"
	"github.com/scottstav/wreccless/internal/hooks"
	"github.com/scottstav/wreccless/internal/state"
	"github.com/spf13/cobra"
)

var denyCmd = &cobra.Command{
	Use:   "deny <id>",
	Short: "Remove a pending worker",
	Args:  cobra.ExactArgs(1),
	RunE:  runDeny,
}

func init() {
	rootCmd.AddCommand(denyCmd)
}

func runDeny(cmd *cobra.Command, args []string) error {
	id := args[0]
	w, err := state.Read(stateDir, id)
	if err != nil {
		return fmt.Errorf("worker %s not found", id)
	}
	if w.Status != state.StatusPending {
		return fmt.Errorf("worker %s is %s, not pending", id, w.Status)
	}

	cfg, _ := config.Load(configPath)

	if err := state.Delete(stateDir, id); err != nil {
		return err
	}

	vars := hooks.Vars{ID: id, Task: w.Task, Dir: w.Directory, Status: "denied"}
	hooks.Fire(cfg.Hooks.OnKill, vars)

	fmt.Fprintf(cmd.OutOrStdout(), "Denied worker %s\n", id)
	return nil
}
