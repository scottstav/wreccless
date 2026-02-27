package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/scottstav/wreccless/internal/state"
)

func TestKill(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	configPath = filepath.Join(t.TempDir(), "nonexistent.toml")
	w := &state.Worker{ID: "800", Status: state.StatusWorking, Directory: "/tmp", Task: "kill me", PID: 99999}
	state.Write(dir, w)

	rootCmd.SetArgs([]string{"kill", "800"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.Execute()

	_, err := state.Read(dir, "800")
	if err == nil {
		t.Error("worker should be deleted after kill")
	}
}

func TestClean(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	state.Write(dir, &state.Worker{ID: "900", Status: state.StatusDone, Directory: "/tmp", Task: "done"})
	state.Write(dir, &state.Worker{ID: "901", Status: state.StatusError, Directory: "/tmp", Task: "error"})
	state.Write(dir, &state.Worker{ID: "902", Status: state.StatusWorking, Directory: "/tmp", Task: "working", PID: 1})

	rootCmd.SetArgs([]string{"clean"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.Execute()

	workers, _ := state.List(dir)
	if len(workers) != 1 {
		t.Errorf("expected 1 remaining (working), got %d", len(workers))
	}
	if workers[0].ID != "902" {
		t.Errorf("wrong surviving worker: %s", workers[0].ID)
	}
}

func TestCleanAll(t *testing.T) {
	dir := t.TempDir()
	stateDir = dir
	state.Write(dir, &state.Worker{ID: "910", Status: state.StatusDone, Directory: "/tmp", Task: "done"})
	state.Write(dir, &state.Worker{ID: "911", Status: state.StatusWorking, Directory: "/tmp", Task: "working", PID: 1})

	rootCmd.SetArgs([]string{"clean", "--all"})
	buf := new(strings.Builder)
	rootCmd.SetOut(buf)
	rootCmd.Execute()

	workers, _ := state.List(dir)
	if len(workers) != 0 {
		t.Errorf("expected 0 remaining, got %d", len(workers))
	}
}
