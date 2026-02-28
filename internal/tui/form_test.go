package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFormFieldNavigation(t *testing.T) {
	f := newForm(80, 24, nil)
	if f.focusIndex != 0 {
		t.Errorf("expected focus at 0, got %d", f.focusIndex)
	}

	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab})
	if f.focusIndex != 1 {
		t.Errorf("expected focus at 1, got %d", f.focusIndex)
	}

	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if f.focusIndex != 0 {
		t.Errorf("expected focus at 0, got %d", f.focusIndex)
	}
}

func TestFormCancel(t *testing.T) {
	f := newForm(80, 24, nil)
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Error("expected Esc to produce a cancel command")
	}
}

func TestFormTogglePending(t *testing.T) {
	f := newForm(80, 24, nil)
	f.focusIndex = 3
	if f.pending {
		t.Error("expected pending to start false")
	}
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !f.pending {
		t.Error("expected pending to be true after toggle")
	}
}

func TestFormDirPickerIntegration(t *testing.T) {
	f := newForm(80, 24, nil)
	if f.focusIndex != 0 {
		t.Errorf("expected focus at 0, got %d", f.focusIndex)
	}
	if f.dirPicker.Value() != "~/" {
		t.Errorf("expected dirpicker value '~/', got %q", f.dirPicker.Value())
	}
}

func TestFormDirPickerNextFieldMsg(t *testing.T) {
	f := newForm(80, 24, nil)
	f, _ = f.Update(dirPickerNextFieldMsg{})
	if f.focusIndex != 1 {
		t.Errorf("expected focus at 1 after dirPickerNextFieldMsg, got %d", f.focusIndex)
	}
}

func TestFormSubmitEmpty(t *testing.T) {
	f := newForm(80, 24, nil)
	f.dirPicker.input.SetValue("")
	_, cmd := f.Update(tea.KeyMsg{Type: tea.KeyCtrlS})
	if cmd != nil {
		t.Error("expected submit with empty fields to produce no command")
	}
}

func TestFormWraparound(t *testing.T) {
	f := newForm(80, 24, nil)
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab}) // -> 1
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab}) // -> 2
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab}) // -> 3 (checkbox)
	f, _ = f.Update(tea.KeyMsg{Type: tea.KeyTab}) // -> 0 (wraparound)
	if f.focusIndex != 0 {
		t.Errorf("expected focus to wrap to 0, got %d", f.focusIndex)
	}
}
