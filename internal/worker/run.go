package worker

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/scottstav/wreccless/internal/config"
	"github.com/scottstav/wreccless/internal/hooks"
	"github.com/scottstav/wreccless/internal/state"
)

// Run executes a worker's claude session. This is a blocking call.
// claudeBin allows overriding the claude binary path for testing.
// Pass "" to use the default "claude" from PATH.
func Run(stateDir, id string, cfg *config.Config, claudeBin string) error {
	w, err := state.Read(stateDir, id)
	if err != nil {
		return fmt.Errorf("read worker: %w", err)
	}

	if claudeBin == "" {
		claudeBin = "claude"
	}

	// Build claude arguments — hardcoded flags that ccl depends on
	args := []string{
		"-p",
		"--output-format", "stream-json",
		"--verbose",
		"--session-id", w.SessionID,
	}
	if cfg.Claude.SkipPermissions {
		args = append(args, "--dangerously-skip-permissions")
	}
	if cfg.Claude.SystemPrompt != "" {
		args = append(args, "--append-system-prompt", cfg.Claude.SystemPrompt)
	}
	args = append(args, cfg.Claude.ExtraFlags...)

	// Build task text (prepend image reference if set)
	task := w.Task
	if w.Image != "" {
		task = fmt.Sprintf("Read and reference this image: %s\n\n%s", w.Image, task)
	}
	args = append(args, task)

	// Open log file
	logPath := filepath.Join(stateDir, id+".log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("create log: %w", err)
	}
	defer logFile.Close()

	// Build and start command
	cmd := exec.Command(claudeBin, args...)
	cmd.Dir = w.Directory
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil

	// Forward SIGTERM to child
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	go func() {
		<-sigCh
		if cmd.Process != nil {
			cmd.Process.Signal(syscall.SIGTERM)
		}
	}()

	if err := cmd.Start(); err != nil {
		markError(stateDir, w, cfg)
		return fmt.Errorf("start claude: %w", err)
	}

	// Update PID in state
	w.PID = cmd.Process.Pid
	state.Write(stateDir, w)

	// Wait for completion
	runErr := cmd.Wait()
	signal.Stop(sigCh)

	now := time.Now()
	w.FinishedAt = &now

	if runErr != nil {
		markError(stateDir, w, cfg)
	} else {
		w.Status = state.StatusDone
		state.Write(stateDir, w)
		vars := hooks.Vars{ID: w.ID, Task: w.Task, Dir: w.Directory, Status: "done", SessionID: w.SessionID}
		hooks.Fire(cfg.Hooks.OnDone, vars)
	}

	return nil
}

func markError(stateDir string, w *state.Worker, cfg *config.Config) {
	now := time.Now()
	w.Status = state.StatusError
	w.FinishedAt = &now
	state.Write(stateDir, w)
	vars := hooks.Vars{ID: w.ID, Task: w.Task, Dir: w.Directory, Status: "error", SessionID: w.SessionID}
	hooks.Fire(cfg.Hooks.OnError, vars)
}
