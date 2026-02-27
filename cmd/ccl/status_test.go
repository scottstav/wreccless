package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/scottstav/wreccless/internal/state"
)

func TestStatus(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	w := &state.Worker{ID: "500", Status: state.StatusWorking, Directory: "/tmp/proj", Task: "build feature", PID: 1, SessionID: "sess-500"}
	state.Write(dir, w)

	rootCmd.SetArgs([]string{"status", "500"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "500") || !strings.Contains(output, "working") {
		t.Errorf("missing info: %s", output)
	}
}

func TestStatusJSON(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	w := &state.Worker{ID: "501", Status: state.StatusDone, Directory: "/tmp", Task: "done task", SessionID: "sess-501"}
	state.Write(dir, w)

	rootCmd.SetArgs([]string{"status", "501", "--json"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.Execute()

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(buf.String()), &result); err != nil {
		t.Fatalf("json: %v", err)
	}
	if result["id"] != "501" {
		t.Errorf("id: %v", result["id"])
	}
}

func TestStatusNotFound(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir

	rootCmd.SetArgs([]string{"status", "999"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent worker")
	}
}
