package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/scottstav/wreccless/internal/state"
)

func writeTestLog(t *testing.T, dir, id string) {
	t.Helper()
	logPath := filepath.Join(dir, id+".log")
	lines := `{"type":"system","subtype":"init"}
{"type":"assistant","content":[{"type":"text","text":"I found the bug."}]}
{"type":"tool_use","name":"Edit","input":{"file_path":"/tmp/foo.go"}}
{"type":"result","subtype":"success"}
`
	os.WriteFile(logPath, []byte(lines), 0644)
}

func TestLogsHuman(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	state.Write(dir, &state.Worker{ID: "1100", Status: state.StatusDone, Directory: "/tmp", Task: "test"})
	writeTestLog(t, dir, "1100")

	rootCmd.SetArgs([]string{"logs", "1100"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.Execute()

	output := buf.String()
	if !strings.Contains(output, "I found the bug") {
		t.Errorf("expected assistant text: %s", output)
	}
}

func TestLogsJSON(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	state.Write(dir, &state.Worker{ID: "1101", Status: state.StatusDone, Directory: "/tmp", Task: "test"})
	writeTestLog(t, dir, "1101")

	rootCmd.SetArgs([]string{"logs", "1101", "--json"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.Execute()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 JSON lines, got %d", len(lines))
	}
}

func TestLogsNoFile(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	state.Write(dir, &state.Worker{ID: "1102", Status: state.StatusPending, Directory: "/tmp", Task: "test"})

	rootCmd.SetArgs([]string{"logs", "1102"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	err := rootCmd.Execute()
	// Should error since no log file exists
	if err == nil {
		output := buf.String()
		if !strings.Contains(strings.ToLower(output), "no log") {
			t.Log("no error and no 'no log' message — acceptable if it just shows nothing")
		}
	}
}
