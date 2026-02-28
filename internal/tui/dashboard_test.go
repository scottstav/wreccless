package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/scottstav/wreccless/internal/state"
)

func setupTestWorkers(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	now := time.Now()
	workers := []*state.Worker{
		{ID: "100", Status: state.StatusWorking, Directory: "/tmp/proj-a", Task: "Build API", CreatedAt: &now, StartedAt: &now, PID: 9999},
		{ID: "101", Status: state.StatusPending, Directory: "/tmp/proj-b", Task: "Fix login bug", CreatedAt: &now},
		{ID: "102", Status: state.StatusDone, Directory: "/tmp/proj-c", Task: "Add tests", CreatedAt: &now, StartedAt: &now, FinishedAt: &now},
	}
	for _, w := range workers {
		state.Write(dir, w)
	}
	return dir
}

func TestDashboardCursorMovement(t *testing.T) {
	dir := setupTestWorkers(t)
	d := newDashboard(dir, "")

	// Load workers
	d.refreshWorkers()
	if len(d.workers) != 3 {
		t.Fatalf("expected 3 workers, got %d", len(d.workers))
	}

	// Cursor starts at 0
	if d.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", d.cursor)
	}

	// Move down with 'j'
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if d.cursor != 1 {
		t.Errorf("expected cursor at 1 after j, got %d", d.cursor)
	}

	// Move down again
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if d.cursor != 2 {
		t.Errorf("expected cursor at 2, got %d", d.cursor)
	}

	// Move down at bottom — should stay
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if d.cursor != 2 {
		t.Errorf("expected cursor to stay at 2, got %d", d.cursor)
	}

	// Move up with 'k'
	d, _ = d.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if d.cursor != 1 {
		t.Errorf("expected cursor at 1 after k, got %d", d.cursor)
	}
}

func TestDashboardSelectedWorker(t *testing.T) {
	dir := setupTestWorkers(t)
	d := newDashboard(dir, "")
	d.refreshWorkers()

	w := d.selectedWorker()
	if w == nil || w.ID != "100" {
		t.Errorf("expected selected worker '100', got %+v", w)
	}

	d.cursor = 1
	w = d.selectedWorker()
	if w == nil || w.ID != "101" {
		t.Errorf("expected selected worker '101', got %+v", w)
	}
}

func TestDashboardLogPreview(t *testing.T) {
	dir := setupTestWorkers(t)
	logPath := filepath.Join(dir, "100.log")
	os.WriteFile(logPath, []byte("{\"type\":\"assistant\",\"content\":\"Working on it...\"}\n"), 0644)

	d := newDashboard(dir, "")
	d.refreshWorkers()
	d.refreshLogPreview()

	if d.logContent == "" {
		t.Error("expected log content to be non-empty")
	}
}

func TestDashboardEmptyWorkers(t *testing.T) {
	dir := t.TempDir()
	d := newDashboard(dir, "")
	d.refreshWorkers()

	if len(d.workers) != 0 {
		t.Errorf("expected 0 workers, got %d", len(d.workers))
	}

	w := d.selectedWorker()
	if w != nil {
		t.Errorf("expected nil selected worker, got %+v", w)
	}
}

func TestDashboardFilter(t *testing.T) {
	dir := setupTestWorkers(t)
	d := newDashboard(dir, "")
	d.filter = "pending"
	d.refreshWorkers()

	if len(d.workers) != 1 {
		t.Fatalf("expected 1 pending worker, got %d", len(d.workers))
	}
	if d.workers[0].ID != "101" {
		t.Errorf("expected worker 101, got %s", d.workers[0].ID)
	}
}
