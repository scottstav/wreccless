package tui

import (
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestLoadHistoryMissingFile(t *testing.T) {
	dirs := loadDirHistory("/nonexistent/path/history.json")
	if len(dirs) != 0 {
		t.Errorf("expected empty slice for missing file, got %d items", len(dirs))
	}
}

func TestSaveAndLoadHistory(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "history.json")

	dirs := []string{"~/projects/foo", "~/projects/bar"}
	if err := saveDirHistory(path, dirs); err != nil {
		t.Fatalf("save: %v", err)
	}

	got := loadDirHistory(path)
	if len(got) != 2 {
		t.Fatalf("expected 2 dirs, got %d", len(got))
	}
	if got[0] != "~/projects/foo" || got[1] != "~/projects/bar" {
		t.Errorf("unexpected dirs: %v", got)
	}
}

func TestSaveHistoryCapsAt50(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "history.json")

	dirs := make([]string, 60)
	for i := range dirs {
		dirs[i] = "~/dir" + string(rune('a'+i%26))
	}
	saveDirHistory(path, dirs)
	got := loadDirHistory(path)
	if len(got) != 50 {
		t.Errorf("expected 50 dirs after cap, got %d", len(got))
	}
}

func TestNewDirPicker(t *testing.T) {
	dp := newDirPicker(nil)
	if dp.Value() != "~/" {
		t.Errorf("expected initial value '~/', got %q", dp.Value())
	}
	if dp.open {
		t.Error("expected dropdown to start closed")
	}
	if dp.cursor != -1 {
		t.Errorf("expected cursor at -1, got %d", dp.cursor)
	}
}

func TestDirPickerCandidatesWithHistory(t *testing.T) {
	history := []string{"~/projects/foo", "~/projects/bar", "~/documents"}
	dp := newDirPicker(history)
	dp.input.SetValue("~/pro")
	dp.refreshCandidates()

	found := false
	for _, c := range dp.candidates {
		if c == "~/projects/foo" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected history entry ~/projects/foo in candidates, got %v", dp.candidates)
	}
}

func TestDirPickerCandidatesMaxFive(t *testing.T) {
	dp := newDirPicker(nil)
	dp.input.SetValue("/") // root has many dirs
	dp.refreshCandidates()
	if len(dp.candidates) > 5 {
		t.Errorf("expected max 5 candidates, got %d", len(dp.candidates))
	}
}

func TestDirPickerDownOpensDropdown(t *testing.T) {
	dp := newDirPicker([]string{"~/projects/foo"})
	dp.input.Focus()
	dp.refreshCandidates()

	if dp.open {
		t.Error("dropdown should start closed")
	}

	dp, _ = dp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if !dp.open {
		t.Error("down arrow should open dropdown")
	}
	if dp.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", dp.cursor)
	}
}

func TestDirPickerCursorWraps(t *testing.T) {
	dp := newDirPicker([]string{"~/a", "~/b"})
	dp.input.Focus()
	dp.input.SetValue("~/")
	dp.refreshCandidates()
	dp.open = true
	dp.cursor = len(dp.candidates) - 1

	dp, _ = dp.Update(tea.KeyMsg{Type: tea.KeyDown})
	if dp.cursor != 0 {
		t.Errorf("expected cursor to wrap to 0, got %d", dp.cursor)
	}
}

func TestDirPickerUpFromTop(t *testing.T) {
	dp := newDirPicker([]string{"~/a", "~/b"})
	dp.input.Focus()
	dp.input.SetValue("~/")
	dp.refreshCandidates()
	dp.open = true
	dp.cursor = 0

	dp, _ = dp.Update(tea.KeyMsg{Type: tea.KeyUp})
	if dp.cursor != len(dp.candidates)-1 {
		t.Errorf("expected cursor to wrap to %d, got %d", len(dp.candidates)-1, dp.cursor)
	}
}

func TestDirPickerEscClosesDropdown(t *testing.T) {
	dp := newDirPicker([]string{"~/a"})
	dp.input.Focus()
	dp.input.SetValue("~/")
	dp.refreshCandidates()
	dp.open = true

	dp, cmd := dp.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if dp.open {
		t.Error("esc should close dropdown")
	}
	if cmd != nil {
		t.Error("esc with open dropdown should not produce a command")
	}
}

func TestDirPickerEscCancelWhenClosed(t *testing.T) {
	dp := newDirPicker(nil)
	dp.input.Focus()

	_, cmd := dp.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("esc with closed dropdown should produce cancel command")
	}
}

func TestDirPickerTypingResetsCursor(t *testing.T) {
	dp := newDirPicker([]string{"~/a", "~/b"})
	dp.input.Focus()
	dp.input.SetValue("~/")
	dp.refreshCandidates()
	dp.open = true
	dp.cursor = 1

	dp, _ = dp.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	if dp.cursor != -1 {
		t.Errorf("typing should reset cursor to -1, got %d", dp.cursor)
	}
}

func TestDirPickerTabDrillsIntoDir(t *testing.T) {
	dp := newDirPicker([]string{"~/projects", "~/documents"})
	dp.input.Focus()
	dp.input.SetValue("~/")
	dp.refreshCandidates()
	dp.open = true
	dp.cursor = 0 // ~/projects selected

	dp, cmd := dp.Update(tea.KeyMsg{Type: tea.KeyTab})
	// Should drill into ~/projects/ (not advance field)
	if cmd != nil {
		t.Error("tab with selection should drill, not produce a command")
	}
	if dp.Value() != "~/projects/" {
		t.Errorf("expected value '~/projects/', got %q", dp.Value())
	}
}

func TestDirPickerTabNoSelectionAdvances(t *testing.T) {
	dp := newDirPicker(nil)
	dp.input.Focus()

	_, cmd := dp.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd == nil {
		t.Error("tab with no selection should produce dirPickerNextFieldMsg")
	}
}

func TestDirPickerEnterAccepts(t *testing.T) {
	dp := newDirPicker([]string{"~/projects", "~/documents"})
	dp.input.Focus()
	dp.input.SetValue("~/")
	dp.refreshCandidates()
	dp.open = true
	dp.cursor = 1 // ~/documents selected

	dp, cmd := dp.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Error("enter should produce dirPickerNextFieldMsg")
	}
	if dp.Value() != "~/documents" {
		t.Errorf("expected value '~/documents', got %q", dp.Value())
	}
	if dp.open {
		t.Error("enter should close dropdown")
	}
}

func TestDirPickerViewShowsCandidates(t *testing.T) {
	dp := newDirPicker([]string{"~/projects/foo", "~/projects/bar"})
	dp.input.Focus()
	dp.input.SetValue("~/")
	dp.refreshCandidates()
	dp.open = true
	dp.cursor = 0

	view := dp.CandidatesView()
	if !strings.Contains(view, "projects/foo") {
		t.Error("expected candidates view to contain highlighted candidate")
	}
}

func TestDirPickerViewHiddenWhenClosed(t *testing.T) {
	dp := newDirPicker([]string{"~/projects/foo"})
	dp.input.Focus()
	dp.input.SetValue("~/")
	dp.refreshCandidates()
	dp.open = false

	view := dp.CandidatesView()
	if strings.Contains(view, "projects/foo") {
		t.Error("candidates should not appear when dropdown is closed")
	}
}
