package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scottstav/wreccless/internal/state"
)

func TestNewPending(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	configPath = filepath.Join(t.TempDir(), "nonexistent.toml")

	rootCmd.SetArgs([]string{"new", "--dir", "/tmp/testproject", "--task", "fix the bug", "--pending"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output == "" {
		t.Fatal("expected worker ID on stdout")
	}

	workers, err := state.List(dir)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(workers) != 1 {
		t.Fatalf("expected 1 worker, got %d", len(workers))
	}
	w := workers[0]
	if w.Status != state.StatusPending {
		t.Errorf("expected pending, got %s", w.Status)
	}
	if w.Directory != "/tmp/testproject" {
		t.Errorf("directory: %s", w.Directory)
	}
	if w.Task != "fix the bug" {
		t.Errorf("task: %s", w.Task)
	}
}

func TestNewPendingJSON(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	configPath = filepath.Join(t.TempDir(), "nonexistent.toml")

	rootCmd.SetArgs([]string{"new", "--dir", "/tmp/proj", "--task", "test", "--pending", "--json"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("json parse: %v (%s)", err, buf.String())
	}
	if result["id"] == nil {
		t.Error("expected id in JSON output")
	}
}
