package main

import (
	"fmt"
	"syscall"

	"github.com/scottstav/wreccless/internal/config"
	"github.com/scottstav/wreccless/internal/hooks"
	"github.com/scottstav/wreccless/internal/state"
	"github.com/spf13/cobra"
)

var killCmd = &cobra.Command{
	Use:   "kill <id>",
	Short: "Kill a running worker",
	Args:  cobra.ExactArgs(1),
	RunE:  runKill,
}

func init() {
	rootCmd.AddCommand(killCmd)
}

func runKill(cmd *cobra.Command, args []string) error {
	id := args[0]
	w, err := state.Read(stateDir, id)
	if err != nil {
		return fmt.Errorf("worker %s not found", id)
	}

	cfg, _ := config.Load(configPath)

	if w.PID > 0 {
		syscall.Kill(w.PID, syscall.SIGTERM)
	}

	state.Delete(stateDir, id)

	vars := hooks.Vars{ID: id, Task: w.Task, Dir: w.Directory, Status: "killed"}
	hooks.Fire(cfg.Hooks.OnKill, vars)

	fmt.Fprintf(cmd.OutOrStdout(), "Killed worker %s\n", id)
	return nil
}
