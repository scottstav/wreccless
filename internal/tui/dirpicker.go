package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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
