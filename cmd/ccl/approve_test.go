package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/scottstav/wreccless/internal/state"
)

func TestApprove(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	configPath = filepath.Join(t.TempDir(), "nonexistent.toml")
	w := &state.Worker{ID: "600", Status: state.StatusPending, Directory: "/tmp", Task: "pending task"}
	state.Write(dir, w)

	rootCmd.SetArgs([]string{"approve", "600"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	updated, _ := state.Read(dir, "600")
	if updated.Status == state.StatusPending {
		t.Error("worker should not be pending after approve")
	}
}

func TestApproveNotPending(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	w := &state.Worker{ID: "601", Status: state.StatusDone, Directory: "/tmp", Task: "done task"}
	state.Write(dir, w)

	rootCmd.SetArgs([]string{"approve", "601"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when approving non-pending worker")
	}
}

func TestDeny(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	configPath = filepath.Join(t.TempDir(), "nonexistent.toml")
	w := &state.Worker{ID: "700", Status: state.StatusPending, Directory: "/tmp", Task: "deny me"}
	state.Write(dir, w)

	rootCmd.SetArgs([]string{"deny", "700"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.Execute()

	_, err := state.Read(dir, "700")
	if err == nil {
		t.Error("worker should be deleted after deny")
	}
}

func TestDenyNotPending(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	w := &state.Worker{ID: "701", Status: state.StatusWorking, Directory: "/tmp", Task: "working", PID: 1}
	state.Write(dir, w)

	rootCmd.SetArgs([]string{"deny", "701"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when denying non-pending worker")
	}
}
