package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFormFieldNavigation(t *testing.T) {
	f := newForm(80, 24)

	// Starts on field 0
	if f.focusIndex != 0 {
		t.Errorf("expected focus at 0, got %d", f.focusIndex)
	}

	// Tab moves to next field
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab})
	if f.focusIndex != 1 {
		t.Errorf("expected focus at 1, got %d", f.focusIndex)
	}

	// Shift+Tab moves back
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if f.focusIndex != 0 {
		t.Errorf("expected focus at 0, got %d", f.focusIndex)
	}
}

func TestFormCancel(t *testing.T) {
	f := newForm(80, 24)
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("expected Esc to produce a cancel command")
	}
}

func TestFormTogglePending(t *testing.T) {
	f := newForm(80, 24)
	// Navigate to pending field (field 3)
	f.focusIndex = 3
	if f.pending {
		t.Error("expected pending to start false")
	}
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !f.pending {
		t.Error("expected pending to be true after toggle")
	}
}

func TestFormDirAutocomplete(t *testing.T) {
	f := newForm(80, 24)
	f.inputs[0].SetValue("/tmp")
	f.updateCompletions()
	// Just verify it doesn't panic
	if f.inputs[0].Value() != "/tmp" {
		t.Errorf("expected /tmp, got %q", f.inputs[0].Value())
	}
}

func TestFormSubmitEmpty(t *testing.T) {
	f := newForm(80, 24)
	// Try to submit with empty fields - should not produce a command
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd != nil {
		t.Error("expected submit with empty fields to produce no command")
	}
}

func TestFormWraparound(t *testing.T) {
	f := newForm(80, 24)
	// Tab through all fields and back to start
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab}) // -> 1
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab}) // -> 2
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab}) // -> 3 (checkbox)
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab}) // -> 0 (wraparound)
	if f.focusIndex != 0 {
		t.Errorf("expected focus to wrap to 0, got %d", f.focusIndex)
	}
}
