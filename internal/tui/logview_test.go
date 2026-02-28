package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/scottstav/wreccless/internal/state"
)

func TestLogViewLoadLog(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	w := &state.Worker{ID: "200", Status: state.StatusWorking, Directory: "/tmp", Task: "test task", CreatedAt: &now, PID: 1}
	state.Write(dir, w)

	// Write a log with many lines
	var content string
	for i := 0; i < 100; i++ {
		content += fmt.Sprintf("{\"type\":\"assistant\",\"content\":\"Line %d\"}\n", i)
	}
	os.WriteFile(filepath.Join(dir, "200.log"), []byte(content), 0644)

	lv := newLogView(dir, "", w, 80, 24)

	if lv.content == "" {
		t.Fatal("expected log content to be non-empty")
	}

	if lv.viewport.TotalLineCount() == 0 {
		t.Error("expected viewport to have lines")
	}
}

func TestLogViewBack(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	w := &state.Worker{ID: "201", Status: state.StatusDone, Directory: "/tmp", Task: "test", CreatedAt: &now}
	state.Write(dir, w)
	os.WriteFile(filepath.Join(dir, "201.log"), []byte("{\"type\":\"assistant\",\"content\":\"done\"}\n"), 0644)

	lv := newLogView(dir, "", w, 80, 24)

	_, cmd := lv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("expected Esc to produce a command (back navigation)")
	}
}

func TestLogViewRefreshLog(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	w := &state.Worker{ID: "202", Status: state.StatusWorking, Directory: "/tmp", Task: "test", CreatedAt: &now, PID: 1}
	state.Write(dir, w)

	logPath := filepath.Join(dir, "202.log")
	os.WriteFile(logPath, []byte("{\"type\":\"assistant\",\"content\":\"First line\"}\n"), 0644)

	lv := newLogView(dir, "", w, 80, 24)
	initialContent := lv.content

	// Append more data
	f, _ := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString("{\"type\":\"assistant\",\"content\":\"Second line\"}\n")
	f.Close()

	lv.refreshLog()

	if lv.content == initialContent {
		t.Error("expected content to update after refreshLog")
	}
	if !containsString(lv.content, "Second line") {
		t.Errorf("expected content to contain 'Second line', got: %s", lv.content)
	}
}

func containsString(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || len(haystack) > 0 && containsSubstring(haystack, needle))
}

func containsSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestLogViewActionApprove(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	w := &state.Worker{ID: "203", Status: state.StatusPending, Directory: "/tmp", Task: "test", CreatedAt: &now}
	state.Write(dir, w)
	os.WriteFile(filepath.Join(dir, "203.log"), []byte(""), 0644)

	lv := newLogView(dir, "", w, 80, 24)

	_, cmd := lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if cmd == nil {
		t.Error("expected 'a' on pending worker to produce an approve action command")
	}
}

func TestLogViewActionKillOnlyWorking(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	w := &state.Worker{ID: "204", Status: state.StatusDone, Directory: "/tmp", Task: "test", CreatedAt: &now}
	state.Write(dir, w)
	os.WriteFile(filepath.Join(dir, "204.log"), []byte(""), 0644)

	lv := newLogView(dir, "", w, 80, 24)

	_, cmd := lv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd != nil {
		t.Error("expected 'x' on done worker to NOT produce a kill command")
	}
}

func TestLogViewNoLogFile(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	w := &state.Worker{ID: "205", Status: state.StatusPending, Directory: "/tmp", Task: "test", CreatedAt: &now}
	state.Write(dir, w)
	// No log file written

	lv := newLogView(dir, "", w, 80, 24)
	// Should not panic, should show "No logs yet." message
	if lv.content == "" {
		t.Error("expected content to be 'No logs yet.' message, got empty")
	}
}
