package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const maxHistory = 50

func loadDirHistory(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var dirs []string
	if err := json.Unmarshal(data, &dirs); err != nil {
		return nil
	}
	return dirs
}

func saveDirHistory(path string, dirs []string) error {
	if len(dirs) > maxHistory {
		dirs = dirs[:maxHistory]
	}
	data, err := json.Marshal(dirs)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// dirPickerNextFieldMsg signals the form should advance to the next field.
type dirPickerNextFieldMsg struct{}

// dirPickerCancelMsg signals the form should cancel.
type dirPickerCancelMsg struct{}

type dirPicker struct {
	input       textinput.Model
	candidates  []string
	cursor      int
	open        bool
	history     []string
	historyPath string
}

func newDirPicker(historyPath string, history []string) dirPicker {
	ti := textinput.New()
	ti.Placeholder = "~/projects/my-app"
	ti.CharLimit = 256
	ti.Width = 50
	ti.SetValue("~/")
	ti.SetCursor(len("~/"))

	return dirPicker{
		input:       ti,
		cursor:      -1,
		history:     history,
		historyPath: historyPath,
	}
}

func (d *dirPicker) Value() string {
	return d.input.Value()
}

func (d *dirPicker) Focus() {
	d.input.Focus()
}

func (d *dirPicker) Blur() {
	d.input.Blur()
	d.open = false
}

func (d *dirPicker) refreshCandidates() {
	val := d.input.Value()

	home, _ := os.UserHomeDir()

	expandTilde := func(p string) string {
		if strings.HasPrefix(p, "~/") {
			return filepath.Join(home, p[2:])
		}
		return p
	}

	collapseTilde := func(p string) string {
		if home != "" {
			return strings.Replace(p, home, "~", 1)
		}
		return p
	}

	query := strings.ToLower(val)

	seen := make(map[string]bool)
	var candidates []string

	// 1. History matches
	for _, h := range d.history {
		if len(candidates) >= 5 {
			break
		}
		if query == "" || strings.Contains(strings.ToLower(h), query) {
			if !seen[h] {
				seen[h] = true
				candidates = append(candidates, h)
			}
		}
	}

	// 2. Filesystem matches
	if len(candidates) < 5 {
		expanded := expandTilde(val)

		var matches []string
		if strings.HasSuffix(val, "/") {
			entries, _ := os.ReadDir(expanded)
			for _, e := range entries {
				if e.IsDir() {
					matches = append(matches, filepath.Join(expanded, e.Name()))
				}
			}
		} else {
			globMatches, _ := filepath.Glob(expanded + "*")
			for _, m := range globMatches {
				info, err := os.Stat(m)
				if err == nil && info.IsDir() {
					matches = append(matches, m)
				}
			}
		}

		for _, m := range matches {
			if len(candidates) >= 5 {
				break
			}
			display := collapseTilde(m)
			if !seen[display] {
				seen[display] = true
				candidates = append(candidates, display)
			}
		}
	}

	d.candidates = candidates
}

func (d dirPicker) Init() tea.Cmd {
	return textinput.Blink
}

func (d dirPicker) Update(msg tea.Msg) (dirPicker, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return d.handleKey(msg)
	}

	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)
	return d, cmd
}

func (d dirPicker) handleKey(msg tea.KeyMsg) (dirPicker, tea.Cmd) {
	switch msg.Type {
	case tea.KeyDown, tea.KeyCtrlN:
		if len(d.candidates) == 0 {
			d.refreshCandidates()
		}
		if len(d.candidates) > 0 {
			d.open = true
			d.cursor++
			if d.cursor >= len(d.candidates) {
				d.cursor = 0
			}
		}
		return d, nil

	case tea.KeyUp, tea.KeyCtrlP:
		if d.open && len(d.candidates) > 0 {
			d.cursor--
			if d.cursor < 0 {
				d.cursor = len(d.candidates) - 1
			}
		}
		return d, nil

	case tea.KeyEsc:
		if d.open {
			d.open = false
			d.cursor = -1
			return d, nil
		}
		return d, func() tea.Msg { return dirPickerCancelMsg{} }

	case tea.KeyTab:
		if d.open && d.cursor >= 0 && d.cursor < len(d.candidates) {
			selected := d.candidates[d.cursor]
			if !strings.HasSuffix(selected, "/") {
				selected += "/"
			}
			d.input.SetValue(selected)
			d.input.SetCursor(len(selected))
			d.cursor = -1
			d.refreshCandidates()
			return d, nil
		}
		return d, func() tea.Msg { return dirPickerNextFieldMsg{} }

	case tea.KeyEnter:
		if d.open && d.cursor >= 0 && d.cursor < len(d.candidates) {
			d.input.SetValue(d.candidates[d.cursor])
		}
		d.open = false
		d.cursor = -1
		return d, func() tea.Msg { return dirPickerNextFieldMsg{} }
	}

	prevVal := d.input.Value()
	var cmd tea.Cmd
	d.input, cmd = d.input.Update(msg)

	if d.input.Value() != prevVal {
		d.cursor = -1
		d.open = true
		d.refreshCandidates()
	}

	return d, cmd
}

func (d dirPicker) View() string {
	return d.input.View()
}

func (d dirPicker) CandidatesView() string {
	if !d.open || len(d.candidates) == 0 {
		return ""
	}

	var b strings.Builder
	for i, c := range d.candidates {
		if i == d.cursor {
			b.WriteString("    " + lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render("> "+c) + "\n")
		} else {
			b.WriteString("    " + mutedStyle.Render("  "+c) + "\n")
		}
	}
	return b.String()
}
