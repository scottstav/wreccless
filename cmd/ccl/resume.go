package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/scottstav/wreccless/internal/config"
	"github.com/scottstav/wreccless/internal/state"
	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume <id>",
	Short: "Resume a worker's claude session interactively",
	Args:  cobra.ExactArgs(1),
	RunE:  runResume,
}

var resumeDryRun bool

func init() {
	resumeCmd.Flags().BoolVar(&resumeDryRun, "dry-run", false, "Print the resume command instead of executing it")
	rootCmd.AddCommand(resumeCmd)
}

func runResume(cmd *cobra.Command, args []string) error {
	id := args[0]
	w, err := state.Read(stateDir, id)
	if err != nil {
		return fmt.Errorf("worker %s not found", id)
	}
	if w.SessionID == "" {
		return fmt.Errorf("worker %s has no session to resume", id)
	}

	cfg, _ := config.Load(configPath)

	claudeArgs := []string{"--resume", w.SessionID}
	if cfg.Claude.SkipPermissions {
		claudeArgs = append(claudeArgs, "--dangerously-skip-permissions")
	}

	if resumeDryRun {
		fmt.Fprintf(cmd.OutOrStdout(), "cd %s && claude %s\n", w.Directory, strings.Join(claudeArgs, " "))
		return nil
	}

	// Clean up state file since user is taking over
	state.Delete(stateDir, id)

	os.Chdir(w.Directory)

	claudePath, err := exec.LookPath("claude")
	if err != nil {
		return fmt.Errorf("claude not found in PATH")
	}
	return syscall.Exec(claudePath, append([]string{"claude"}, claudeArgs...), os.Environ())
}
