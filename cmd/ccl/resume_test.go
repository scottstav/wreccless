package main

import (
	"strings"
	"testing"

	"github.com/scottstav/wreccless/internal/state"
)

func TestResumeDryRun(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	w := &state.Worker{
		ID:        "1200",
		Status:    state.StatusDone,
		Directory: "/tmp/myproject",
		Task:      "done task",
		SessionID: "sess-1200-uuid",
	}
	state.Write(dir, w)

	rootCmd.SetArgs([]string{"resume", "1200", "--dry-run"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "sess-1200-uuid") {
		t.Errorf("expected session ID: %s", output)
	}
	if !strings.Contains(output, "/tmp/myproject") {
		t.Errorf("expected directory: %s", output)
	}
}

func TestResumeNoSession(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	w := &state.Worker{ID: "1201", Status: state.StatusPending, Directory: "/tmp", Task: "no session"}
	state.Write(dir, w)

	rootCmd.SetArgs([]string{"resume", "1201"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for worker without session")
	}
}
