package tui

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// cancelMsg signals form cancellation.
type cancelMsg struct{}

// createMsg signals worker creation with form data.
type createMsg struct {
	dir     string
	task    string
	image   string
	pending bool
}

type form struct {
	inputs      []textinput.Model
	focusIndex  int
	pending     bool
	completions []string
	width       int
	height      int
}

func newForm(width, height int) form {
	dirInput := textinput.New()
	dirInput.Placeholder = "~/projects/my-app"
	dirInput.Focus()
	dirInput.CharLimit = 256
	dirInput.Width = 50

	taskInput := textinput.New()
	taskInput.Placeholder = "Describe the task..."
	taskInput.CharLimit = 500
	taskInput.Width = 50

	imageInput := textinput.New()
	imageInput.Placeholder = "(optional) path to image"
	imageInput.CharLimit = 256
	imageInput.Width = 50

	return form{
		inputs: []textinput.Model{dirInput, taskInput, imageInput},
		width:  width,
		height: height,
	}
}

func (f form) Init() tea.Cmd {
	return textinput.Blink
}

func (f form) Update(msg tea.Msg) (form, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return f.handleKey(msg)
	}

	// Update the focused input
	if f.focusIndex < len(f.inputs) {
		var cmd tea.Cmd
		f.inputs[f.focusIndex], cmd = f.inputs[f.focusIndex].Update(msg)
		return f, cmd
	}
	return f, nil
}

func (f form) handleKey(msg tea.KeyMsg) (form, tea.Cmd) {
	switch {
	case key.Matches(msg, formKeys.Cancel):
		return f, func() tea.Msg { return cancelMsg{} }

	case key.Matches(msg, formKeys.Submit):
		dir := f.inputs[0].Value()
		task := f.inputs[1].Value()
		if dir == "" || task == "" {
			return f, nil // don't submit empty
		}
		// Expand ~
		if strings.HasPrefix(dir, "~/") {
			home, _ := os.UserHomeDir()
			dir = filepath.Join(home, dir[2:])
		}
		return f, func() tea.Msg {
			return createMsg{
				dir:     dir,
				task:    task,
				image:   f.inputs[2].Value(),
				pending: f.pending,
			}
		}

	case key.Matches(msg, formKeys.NextField):
		if f.focusIndex < len(f.inputs) {
			f.inputs[f.focusIndex].Blur()
		}
		f.focusIndex++
		if f.focusIndex > 3 { // 3 inputs + 1 checkbox = 4 fields (0-3)
			f.focusIndex = 0
		}
		if f.focusIndex < len(f.inputs) {
			f.inputs[f.focusIndex].Focus()
		}
		return f, nil

	case key.Matches(msg, formKeys.PrevField):
		if f.focusIndex < len(f.inputs) {
			f.inputs[f.focusIndex].Blur()
		}
		f.focusIndex--
		if f.focusIndex < 0 {
			f.focusIndex = 3
		}
		if f.focusIndex < len(f.inputs) {
			f.inputs[f.focusIndex].Focus()
		}
		return f, nil

	case key.Matches(msg, formKeys.Toggle):
		if f.focusIndex == 3 { // pending checkbox
			f.pending = !f.pending
			return f, nil
		}
	}

	// Pass key to focused input
	if f.focusIndex < len(f.inputs) {
		var cmd tea.Cmd
		f.inputs[f.focusIndex], cmd = f.inputs[f.focusIndex].Update(msg)

		// Update completions when typing in dir or image field
		if f.focusIndex == 0 || f.focusIndex == 2 {
			f.updateCompletions()
		}

		return f, cmd
	}
	return f, nil
}

func (f *form) updateCompletions() {
	if f.focusIndex >= len(f.inputs) {
		return
	}
	val := f.inputs[f.focusIndex].Value()
	if val == "" {
		f.completions = nil
		return
	}

	// Expand ~
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

	labels := []string{"Directory:", "Task:", "Image:"}
	for i, label := range labels {
		style := formLabelStyle
		if i == f.focusIndex {
			style = style.Foreground(colorPrimary)
		} else {
			style = style.Foreground(colorMuted)
		}
		b.WriteString("  " + style.Render(label) + " " + f.inputs[i].View())
		b.WriteString("\n")

		// Show completions for directory/image fields
		if (i == 0 || i == 2) && i == f.focusIndex && len(f.completions) > 0 {
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
	b.WriteString(helpKeyStyle.Render("[Ctrl+S]") + " " + helpDescStyle.Render("create"))
	b.WriteString("  ")
	b.WriteString(helpKeyStyle.Render("[Esc]") + " " + helpDescStyle.Render("cancel"))

	// Wrap in border
	content := b.String()
	boxWidth := 60
	if f.width < boxWidth+4 {
		boxWidth = f.width - 4
	}
	return formBorderStyle.Width(boxWidth).Render(content)
}
