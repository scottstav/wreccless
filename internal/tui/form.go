package tui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type cancelMsg struct{}

type createMsg struct {
	dir     string
	task    string
	image   string
	pending bool
}

type form struct {
	dirPicker   dirPicker
	inputs      []textinput.Model // task, image (2 inputs)
	focusIndex  int               // 0=dirpicker, 1=task, 2=image, 3=pending
	pending     bool
	completions []string // for image field only
	width       int
	height      int
}

func newForm(width, height int, historyPath string, history []string) form {
	dp := newDirPicker(historyPath, history)
	dp.Focus()

	taskInput := textinput.New()
	taskInput.Placeholder = "Describe the task..."
	taskInput.CharLimit = 500
	taskInput.Width = 50

	imageInput := textinput.New()
	imageInput.Placeholder = "(optional) path to image"
	imageInput.CharLimit = 256
	imageInput.Width = 50

	return form{
		dirPicker: dp,
		inputs:    []textinput.Model{taskInput, imageInput},
		width:     width,
		height:    height,
	}
}

func (f form) Init() tea.Cmd {
	return textinput.Blink
}

func (f form) Update(msg tea.Msg) (form, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return f.handleKey(msg)
	case dirPickerNextFieldMsg:
		return f.advanceFocus(), nil
	case dirPickerCancelMsg:
		return f, func() tea.Msg { return cancelMsg{} }
	}

	// Update the focused component
	if f.focusIndex == 0 {
		var cmd tea.Cmd
		f.dirPicker, cmd = f.dirPicker.Update(msg)
		return f, cmd
	}
	idx := f.focusIndex - 1
	if idx >= 0 && idx < len(f.inputs) {
		var cmd tea.Cmd
		f.inputs[idx], cmd = f.inputs[idx].Update(msg)
		return f, cmd
	}
	return f, nil
}

func (f form) handleKey(msg tea.KeyMsg) (form, tea.Cmd) {
	switch {
	case key.Matches(msg, formKeys.Cancel):
		if f.focusIndex == 0 && f.dirPicker.open {
			var cmd tea.Cmd
			f.dirPicker, cmd = f.dirPicker.Update(msg)
			return f, cmd
		}
		return f, func() tea.Msg { return cancelMsg{} }

	case key.Matches(msg, formKeys.Submit):
		dir := f.dirPicker.Value()
		task := f.inputs[0].Value()
		if dir == "" || task == "" {
			return f, nil
		}
		if strings.HasPrefix(dir, "~/") {
			home, _ := os.UserHomeDir()
			dir = filepath.Join(home, dir[2:])
		}
		return f, func() tea.Msg {
			return createMsg{
				dir:     dir,
				task:    task,
				image:   f.inputs[1].Value(),
				pending: f.pending,
			}
		}

	case key.Matches(msg, formKeys.NextField):
		if f.focusIndex == 0 && f.dirPicker.open && f.dirPicker.cursor >= 0 {
			var cmd tea.Cmd
			f.dirPicker, cmd = f.dirPicker.Update(msg)
			return f, cmd
		}
		return f.advanceFocus(), nil

	case key.Matches(msg, formKeys.PrevField):
		return f.retreatFocus(), nil

	case key.Matches(msg, formKeys.Toggle):
		if f.focusIndex == 3 {
			f.pending = !f.pending
			return f, nil
		}
	}

	// Delegate to focused component
	if f.focusIndex == 0 {
		var cmd tea.Cmd
		f.dirPicker, cmd = f.dirPicker.Update(msg)
		return f, cmd
	}

	idx := f.focusIndex - 1
	if idx >= 0 && idx < len(f.inputs) {
		var cmd tea.Cmd
		f.inputs[idx], cmd = f.inputs[idx].Update(msg)

		if f.focusIndex == 2 {
			f.updateImageCompletions()
		}

		return f, cmd
	}
	return f, nil
}

func (f form) advanceFocus() form {
	if f.focusIndex == 0 {
		f.dirPicker.Blur()
	} else {
		idx := f.focusIndex - 1
		if idx >= 0 && idx < len(f.inputs) {
			f.inputs[idx].Blur()
		}
	}

	f.focusIndex++
	if f.focusIndex > 3 {
		f.focusIndex = 0
	}

	if f.focusIndex == 0 {
		f.dirPicker.Focus()
	} else {
		idx := f.focusIndex - 1
		if idx >= 0 && idx < len(f.inputs) {
			f.inputs[idx].Focus()
		}
	}
	return f
}

func (f form) retreatFocus() form {
	if f.focusIndex == 0 {
		f.dirPicker.Blur()
	} else {
		idx := f.focusIndex - 1
		if idx >= 0 && idx < len(f.inputs) {
			f.inputs[idx].Blur()
		}
	}

	f.focusIndex--
	if f.focusIndex < 0 {
		f.focusIndex = 3
	}

	if f.focusIndex == 0 {
		f.dirPicker.Focus()
	} else {
		idx := f.focusIndex - 1
		if idx >= 0 && idx < len(f.inputs) {
			f.inputs[idx].Focus()
		}
	}
	return f
}

func (f *form) updateImageCompletions() {
	val := f.inputs[1].Value()
	if val == "" {
		f.completions = nil
		return
	}

	expanded := val
	if strings.HasPrefix(expanded, "~/") {
		home, _ := os.UserHomeDir()
		expanded = filepath.Join(home, expanded[2:])
	}

	matches, _ := filepath.Glob(expanded + "*")
	var dirs []string
	for _, m := range matches {
		info, err := os.Stat(m)
		if err == nil && info.IsDir() {
			dirs = append(dirs, m)
		}
	}
	if len(dirs) > 5 {
		dirs = dirs[:5]
	}
	f.completions = dirs
}

func (f form) View() string {
	var b strings.Builder

	title := titleStyle.Render("New Worker")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Field 0: Directory (dirpicker)
	style := formLabelStyle
	if f.focusIndex == 0 {
		style = style.Foreground(colorPrimary)
	} else {
		style = style.Foreground(colorMuted)
	}
	b.WriteString("  " + style.Render("Directory:") + " " + f.dirPicker.View())
	b.WriteString("\n")
	if f.focusIndex == 0 {
		b.WriteString(f.dirPicker.CandidatesView())
	}

	// Fields 1-2: Task, Image
	labels := []string{"Task:", "Image:"}
	for i, label := range labels {
		style := formLabelStyle
		fieldIdx := i + 1
		if fieldIdx == f.focusIndex {
			style = style.Foreground(colorPrimary)
		} else {
			style = style.Foreground(colorMuted)
		}
		b.WriteString("  " + style.Render(label) + " " + f.inputs[i].View())
		b.WriteString("\n")

		if i == 1 && f.focusIndex == 2 && len(f.completions) > 0 {
			home, _ := os.UserHomeDir()
			for _, c := range f.completions {
				display := c
				if home != "" {
					display = strings.Replace(display, home, "~", 1)
				}
				b.WriteString("    " + mutedStyle.Render(display) + "\n")
			}
		}
	}

	// Pending checkbox
	checkStyle := formLabelStyle
	if f.focusIndex == 3 {
		checkStyle = checkStyle.Foreground(colorPrimary)
	} else {
		checkStyle = checkStyle.Foreground(colorMuted)
	}
	check := "[ ]"
	if f.pending {
		check = "[x]"
	}
	b.WriteString("  " + checkStyle.Render("Pending:") + " " + check)
	b.WriteString("\n\n")

	// Help
	b.WriteString("  ")
	b.WriteString(helpKeyStyle.Render("[Tab]") + " " + helpDescStyle.Render("next"))
	b.WriteString("  ")
	b.WriteString(helpKeyStyle.Render("[↓/↑]") + " " + helpDescStyle.Render("browse"))
	b.WriteString("  ")
	b.WriteString(helpKeyStyle.Render("[Ctrl+S]") + " " + helpDescStyle.Render("create"))
	b.WriteString("  ")
	b.WriteString(helpKeyStyle.Render("[Esc]") + " " + helpDescStyle.Render("cancel"))

	content := b.String()
	boxWidth := 60
	if f.width < boxWidth+4 {
		boxWidth = f.width - 4
	}
	return formBorderStyle.Width(boxWidth).Render(content)
}
