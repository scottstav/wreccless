package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/scottstav/wreccless/internal/config"
	"github.com/scottstav/wreccless/internal/hooks"
	"github.com/scottstav/wreccless/internal/state"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new worker",
	RunE:  runNew,
}

var (
	newDir     string
	newTask    string
	newImage   string
	newPending bool
	newJSON    bool
)

func init() {
	newCmd.Flags().StringVar(&newDir, "dir", "", "Project directory (required)")
	newCmd.Flags().StringVar(&newTask, "task", "", "Task description (required)")
	newCmd.Flags().StringVar(&newImage, "image", "", "Image path for claude to reference")
	newCmd.Flags().BoolVar(&newPending, "pending", false, "Create as pending (require manual approval)")
	newCmd.Flags().BoolVar(&newJSON, "json", false, "Output JSON")
	newCmd.MarkFlagRequired("dir")
	newCmd.MarkFlagRequired("task")
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	now := time.Now()
	id := strconv.FormatInt(now.Unix(), 10)

	status := state.StatusWorking
	if newPending {
		status = state.StatusPending
	}

	w := &state.Worker{
		ID:        id,
		Status:    status,
		Directory: newDir,
		Task:      newTask,
		Image:     newImage,
		SessionID: uuid.New().String(),
		CreatedAt: &now,
	}

	if err := state.Write(stateDir, w); err != nil {
		return fmt.Errorf("write state: %w", err)
	}

	vars := hooks.Vars{ID: id, Task: newTask, Dir: newDir, Status: string(status)}
	if newPending {
		hooks.Fire(cfg.Hooks.OnPending, vars)
	} else {
		startedAt := time.Now()
		w.StartedAt = &startedAt
		state.Write(stateDir, w)
		hooks.Fire(cfg.Hooks.OnStart, vars)
		// TODO: Task 11 — spawn detached ccl run here
	}

	if newJSON {
		data, _ := json.Marshal(map[string]string{"id": id, "status": string(status)})
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	} else {
		fmt.Fprintln(cmd.OutOrStdout(), id)
	}
	return nil
}
