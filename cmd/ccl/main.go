package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	stateDir   = filepath.Join(os.Getenv("HOME"), ".local", "state", "ccl")
	configPath = filepath.Join(os.Getenv("HOME"), ".config", "ccl", "config.toml")
)

var rootCmd = &cobra.Command{
	Use:   "ccl",
	Short: "Claude Code Launcher — manage background Claude workers",
}

func init() {
	if v := os.Getenv("CCL_STATE_DIR"); v != "" {
		stateDir = v
	}
	if v := os.Getenv("CCL_CONFIG"); v != "" {
		configPath = v
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
