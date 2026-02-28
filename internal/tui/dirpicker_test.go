package tui

import (
	"path/filepath"
	"testing"
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
	dp := newDirPicker("", nil)
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
	dp := newDirPicker("", history)
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
	dp := newDirPicker("", nil)
	dp.input.SetValue("/") // root has many dirs
	dp.refreshCandidates()
	if len(dp.candidates) > 5 {
		t.Errorf("expected max 5 candidates, got %d", len(dp.candidates))
	}
}
