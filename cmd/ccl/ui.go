package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/scottstav/wreccless/internal/config"
	"github.com/scottstav/wreccless/internal/tui"
	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch the interactive TUI dashboard",
	RunE:  runUI,
}

func init() {
	rootCmd.AddCommand(uiCmd)
}

func runUI(cmd *cobra.Command, args []string) error {
	app := tui.NewApp(stateDir, configPath)
	p := tea.NewProgram(app, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Check if we need to resume a session
	if a, ok := finalModel.(tui.App); ok && a.ResumeWorker != nil {
		cfg, _ := config.Load(configPath)
		claudeArgs := []string{"--resume", a.ResumeWorker.SessionID}
		if cfg.Claude.SkipPermissions {
			claudeArgs = append(claudeArgs, "--dangerously-skip-permissions")
		}

		os.Chdir(a.ResumeWorker.Directory)

		claudePath, err := exec.LookPath("claude")
		if err != nil {
			return fmt.Errorf("claude not found in PATH")
		}
		return syscall.Exec(claudePath, append([]string{"claude"}, claudeArgs...), os.Environ())
	}

	return nil
}
