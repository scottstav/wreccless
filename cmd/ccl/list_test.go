package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/scottstav/wreccless/internal/state"
)

func seedWorkers(t *testing.T, dir string) {
	t.Helper()
	workers := []*state.Worker{
		{ID: "100", Status: state.StatusPending, Directory: "/tmp/a", Task: "task a"},
		{ID: "200", Status: state.StatusWorking, Directory: "/tmp/b", Task: "task b", PID: 99999},
		{ID: "300", Status: state.StatusDone, Directory: "/tmp/c", Task: "task c", SessionID: "sess-c"},
	}
	for _, w := range workers {
		if err := state.Write(dir, w); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

func resetListFlags() {
	listJSON = false
	listStatus = ""
}

func TestListHuman(t *testing.T) {
	resetListFlags()
	dir := t.TempDir()
	stateDir = dir
	seedWorkers(t, dir)

	rootCmd.SetArgs([]string{"list"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	output := buf.String()
	if !strings.Contains(output, "100") || !strings.Contains(output, "pending") {
		t.Errorf("missing worker 100: %s", output)
	}
}

func TestListJSON(t *testing.T) {
	resetListFlags()
	dir := t.TempDir()
	stateDir = dir
	seedWorkers(t, dir)

	rootCmd.SetArgs([]string{"list", "--json"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	var workers []map[string]interface{}
	if err := json.Unmarshal([]byte(buf.String()), &workers); err != nil {
		t.Fatalf("json: %v (%s)", err, buf.String())
	}
	if len(workers) != 3 {
		t.Errorf("expected 3, got %d", len(workers))
	}
}

func TestListFilterStatus(t *testing.T) {
	resetListFlags()
	dir := t.TempDir()
	stateDir = dir
	seedWorkers(t, dir)

	rootCmd.SetArgs([]string{"list", "--json", "--status", "pending"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	var workers []map[string]interface{}
	json.Unmarshal([]byte(buf.String()), &workers)
	if len(workers) != 1 {
		t.Errorf("expected 1 pending, got %d", len(workers))
	}
}

func TestListEmpty(t *testing.T) {
	resetListFlags()
	dir := t.TempDir()
	stateDir = dir

	rootCmd.SetArgs([]string{"list"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	output := strings.TrimSpace(buf.String())
	if output != "No workers." {
		t.Errorf("expected 'No workers.', got: %q", output)
	}
}

func TestListStaleDetection(t *testing.T) {
	resetListFlags()
	dir := t.TempDir()
	stateDir = dir
	w := &state.Worker{ID: "400", Status: state.StatusWorking, Directory: "/tmp", Task: "stale", PID: 99999}
	state.Write(dir, w)

	rootCmd.SetArgs([]string{"list", "--json"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.Execute()

	var workers []map[string]interface{}
	json.Unmarshal([]byte(buf.String()), &workers)
	if len(workers) != 1 {
		t.Fatalf("expected 1, got %d", len(workers))
	}
	if workers[0]["status"] != "error" {
		t.Errorf("stale worker should be error, got %s", workers[0]["status"])
	}
}
