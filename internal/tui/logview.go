package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/scottstav/wreccless/internal/logrender"
	"github.com/scottstav/wreccless/internal/state"
)

// backMsg signals the app to return to the dashboard.
type backMsg struct{}

// actionMsg signals the app to perform an action on a worker.
type actionMsg struct {
	action string // "approve", "deny", "kill", "clean", "resume"
	worker *state.Worker
}

type logView struct {
	stateDir   string
	configPath string
	worker     *state.Worker
	viewport   viewport.Model
	content    string
	width      int
	height     int
	lastOffset int64
	atBottom   bool
}

func newLogView(stateDir, configPath string, w *state.Worker, width, height int) logView {
	vp := viewport.New(width, height-4)
	vp.SetContent("")

	lv := logView{
		stateDir:   stateDir,
		configPath: configPath,
		worker:     w,
		viewport:   vp,
		width:      width,
		height:     height,
		atBottom:   true,
	}
	lv.loadLog()
	return lv
}

func (lv logView) Init() tea.Cmd {
	return nil
}

func (lv logView) Update(msg tea.Msg) (logView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return lv.handleKey(msg)
	}

	var cmd tea.Cmd
	lv.viewport, cmd = lv.viewport.Update(msg)
	return lv, cmd
}

func (lv logView) handleKey(msg tea.KeyMsg) (logView, tea.Cmd) {
	switch {
	case key.Matches(msg, logViewKeys.Back):
		return lv, func() tea.Msg { return backMsg{} }
	case key.Matches(msg, logViewKeys.Bottom):
		lv.viewport.GotoBottom()
		lv.atBottom = true
		return lv, nil
	case key.Matches(msg, logViewKeys.Top):
		lv.viewport.GotoTop()
		lv.atBottom = false
		return lv, nil
	case key.Matches(msg, logViewKeys.HalfDown):
		lv.viewport.HalfPageDown()
		lv.atBottom = lv.viewport.AtBottom()
		return lv, nil
	case key.Matches(msg, logViewKeys.HalfUp):
		lv.viewport.HalfPageUp()
		lv.atBottom = false
		return lv, nil
	case key.Matches(msg, logViewKeys.Down):
		lv.viewport.ScrollDown(1)
		lv.atBottom = lv.viewport.AtBottom()
		return lv, nil
	case key.Matches(msg, logViewKeys.Up):
		lv.viewport.ScrollUp(1)
		lv.atBottom = false
		return lv, nil

	// Actions
	case key.Matches(msg, logViewKeys.Approve):
		if lv.worker.Status == state.StatusPending {
			return lv, func() tea.Msg { return actionMsg{action: "approve", worker: lv.worker} }
		}
	case key.Matches(msg, logViewKeys.Deny):
		if lv.worker.Status == state.StatusPending {
			return lv, func() tea.Msg { return actionMsg{action: "deny", worker: lv.worker} }
		}
	case key.Matches(msg, logViewKeys.Kill):
		if lv.worker.Status == state.StatusWorking {
			return lv, func() tea.Msg { return actionMsg{action: "kill", worker: lv.worker} }
		}
	case key.Matches(msg, logViewKeys.Resume):
		return lv, func() tea.Msg { return actionMsg{action: "resume", worker: lv.worker} }
	case key.Matches(msg, logViewKeys.Clean):
		if lv.worker.Status == state.StatusDone || lv.worker.Status == state.StatusError {
			return lv, func() tea.Msg { return actionMsg{action: "clean", worker: lv.worker} }
		}
	}
	return lv, nil
}

func (lv *logView) loadLog() {
	logPath := filepath.Join(lv.stateDir, lv.worker.ID+".log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		lv.content = mutedStyle.Render("No logs yet.")
		lv.viewport.SetContent(lv.content)
		return
	}

	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		events := logrender.ParseLine([]byte(line))
		for _, e := range events {
			switch e.Type {
			case logrender.EventText:
				lines = append(lines, e.Text)
			case logrender.EventTool:
				lines = append(lines, toolStyle.Render(fmt.Sprintf("[tool: %s]", e.ToolName)))
			case logrender.EventResult:
				lines = append(lines, resultStyle.Render(fmt.Sprintf("[result: %s]", e.SubType)))
			}
		}
	}

	lv.content = strings.Join(lines, "\n")
	lv.viewport.SetContent(lv.content)
	if lv.atBottom {
		lv.viewport.GotoBottom()
	}
	lv.lastOffset = int64(len(data))
}

// refreshLog reads new data appended since last read.
func (lv *logView) refreshLog() {
	logPath := filepath.Join(lv.stateDir, lv.worker.ID+".log")
	f, err := os.Open(logPath)
	if err != nil {
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil || info.Size() <= lv.lastOffset {
		return
	}

	f.Seek(lv.lastOffset, 0)
	newData := make([]byte, info.Size()-lv.lastOffset)
	n, err := f.Read(newData)
	if err != nil || n == 0 {
		return
	}

	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(string(newData[:n])), "\n") {
		if line == "" {
			continue
		}
		events := logrender.ParseLine([]byte(line))
		for _, e := range events {
			switch e.Type {
			case logrender.EventText:
				lines = append(lines, e.Text)
			case logrender.EventTool:
				lines = append(lines, toolStyle.Render(fmt.Sprintf("[tool: %s]", e.ToolName)))
			case logrender.EventResult:
				lines = append(lines, resultStyle.Render(fmt.Sprintf("[result: %s]", e.SubType)))
			}
		}
	}

	if len(lines) > 0 {
		if lv.content != "" {
			lv.content += "\n"
		}
		lv.content += strings.Join(lines, "\n")
		lv.viewport.SetContent(lv.content)
		if lv.atBottom {
			lv.viewport.GotoBottom()
		}
	}
	lv.lastOffset = info.Size()
}

// refreshWorker re-reads the worker state from disk.
func (lv *logView) refreshWorker() {
	w, err := state.Read(lv.stateDir, lv.worker.ID)
	if err != nil {
		return
	}
	lv.worker = w
}

func (lv logView) View() string {
	var b strings.Builder

	// Header
	task := lv.worker.Task
	maxTask := lv.width - 30
	if maxTask > 0 && len(task) > maxTask {
		task = task[:maxTask-3] + "..."
	}
	header := fmt.Sprintf(" LOGS: %s │ %s", lv.worker.ID, task)
	escHint := helpKeyStyle.Render("[Esc]") + " " + helpDescStyle.Render("back")
	padding := lv.width - lipgloss.Width(header) - lipgloss.Width(escHint) - 2
	if padding < 1 {
		padding = 1
	}
	b.WriteString(titleStyle.Render(header + strings.Repeat(" ", padding) + escHint))
	b.WriteString("\n")

	// Separator
	b.WriteString(mutedStyle.Render(strings.Repeat("─", lv.width)))
	b.WriteString("\n")

	// Viewport
	b.WriteString(lv.viewport.View())
	b.WriteString("\n")

	// Footer
	b.WriteString(mutedStyle.Render(strings.Repeat("─", lv.width)))
	b.WriteString("\n")
	b.WriteString(lv.renderHelp())

	return b.String()
}

func (lv logView) renderHelp() string {
	var parts []string
	add := func(k, desc string) {
		parts = append(parts, helpKeyStyle.Render(k)+" "+helpDescStyle.Render(desc))
	}

	switch lv.worker.Status {
	case state.StatusPending:
		add("[a]", "approve")
		add("[d]", "deny")
	case state.StatusWorking:
		add("[x]", "kill")
	case state.StatusDone, state.StatusError:
		add("[r]", "resume")
		add("[c]", "clean")
	}

	add("[Esc]", "back")
	return "  " + strings.Join(parts, "  ")
}
