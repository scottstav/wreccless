package worker

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/scottstav/wreccless/internal/config"
	"github.com/scottstav/wreccless/internal/state"
)

func writeMockClaude(t *testing.T, dir string, exitCode int) string {
	t.Helper()
	script := filepath.Join(dir, "mock-claude")
	content := fmt.Sprintf(`#!/bin/sh
echo '{"type":"system","subtype":"init"}'
echo '{"type":"assistant","content":"I fixed the bug."}'
echo '{"type":"result","subtype":"success"}'
exit %d
`, exitCode)
	os.WriteFile(script, []byte(content), 0755)
	return script
}

func TestRunSuccess(t *testing.T) {
	stateDir := t.TempDir()
	binDir := t.TempDir()
	mockClaude := writeMockClaude(t, binDir, 0)

	w := &state.Worker{ID: "1000", Status: state.StatusWorking, Directory: t.TempDir(), Task: "fix bug", SessionID: "test-session"}
	state.Write(stateDir, w)

	cfg := config.Defaults()

	err := Run(stateDir, "1000", cfg, mockClaude)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	updated, err := state.Read(stateDir, "1000")
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if updated.Status != state.StatusDone {
		t.Errorf("expected done, got %s", updated.Status)
	}
	if updated.FinishedAt == nil {
		t.Error("finished_at should be set")
	}

	logPath := filepath.Join(stateDir, "1000.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log: %v", err)
	}
	if len(data) == 0 {
		t.Error("log should not be empty")
	}
}

func TestRunFailure(t *testing.T) {
	stateDir := t.TempDir()
	binDir := t.TempDir()
	mockClaude := writeMockClaude(t, binDir, 1)

	w := &state.Worker{ID: "1001", Status: state.StatusWorking, Directory: t.TempDir(), Task: "fail task", SessionID: "test-session"}
	state.Write(stateDir, w)

	cfg := config.Defaults()

	err := Run(stateDir, "1001", cfg, mockClaude)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	updated, _ := state.Read(stateDir, "1001")
	if updated.Status != state.StatusError {
		t.Errorf("expected error, got %s", updated.Status)
	}
}
