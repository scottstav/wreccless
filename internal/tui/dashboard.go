package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/scottstav/wreccless/internal/logrender"
	"github.com/scottstav/wreccless/internal/state"
)

type dashboard struct {
	stateDir   string
	configPath string
	workers    []*state.Worker
	cursor     int
	width      int
	height     int
	spinner    spinner.Model
	logContent string
	filter     string
	flash      string
	flashErr   bool
}

func newDashboard(stateDir, configPath string) dashboard {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorPrimary)
	return dashboard{
		stateDir:   stateDir,
		configPath: configPath,
		spinner:    s,
	}
}

func (d dashboard) Init() tea.Cmd {
	return d.spinner.Tick
}

func (d dashboard) Update(msg tea.Msg) (dashboard, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return d.handleKey(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		d.spinner, cmd = d.spinner.Update(msg)
		return d, cmd
	}
	return d, nil
}

func (d dashboard) handleKey(msg tea.KeyMsg) (dashboard, tea.Cmd) {
	switch {
	case key.Matches(msg, dashboardKeys.Down):
		if d.cursor < len(d.workers)-1 {
			d.cursor++
			d.refreshLogPreview()
		}
	case key.Matches(msg, dashboardKeys.Up):
		if d.cursor > 0 {
			d.cursor--
			d.refreshLogPreview()
		}
	}
	return d, nil
}

func (d *dashboard) refreshWorkers() {
	workers, err := state.List(d.stateDir)
	if err != nil {
		return
	}
	// Stale detection
	for _, w := range workers {
		if w.Status == state.StatusWorking && w.PID > 0 && !isAlive(w.PID) {
			w.Status = state.StatusError
			state.Write(d.stateDir, w)
		}
	}
	// Apply filter
	if d.filter != "" {
		var filtered []*state.Worker
		for _, w := range workers {
			if string(w.Status) == d.filter {
				filtered = append(filtered, w)
			}
		}
		workers = filtered
	}
	d.workers = workers
	if d.cursor >= len(d.workers) && len(d.workers) > 0 {
		d.cursor = len(d.workers) - 1
	}
	if d.cursor < 0 {
		d.cursor = 0
	}
}

func isAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

func (d *dashboard) refreshLogPreview() {
	w := d.selectedWorker()
	if w == nil {
		d.logContent = ""
		return
	}
	logPath := filepath.Join(d.stateDir, w.ID+".log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		d.logContent = mutedStyle.Render("No logs yet.")
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
	d.logContent = strings.Join(lines, "\n")
}

func (d dashboard) selectedWorker() *state.Worker {
	if d.cursor < 0 || d.cursor >= len(d.workers) {
		return nil
	}
	return d.workers[d.cursor]
}

func (d dashboard) View() string {
	if d.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Title bar
	title := titleStyle.Render("wreccless")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Calculate fixed overhead lines:
	//   title(1) + blank(1) + blank_before_log(1) + log_title(1) +
	//   log_border_top(1) + log_border_bottom(1) + help(1) = 7
	// When flash is visible, it reuses the blank_before_log line
	// (becomes blank_before_flash) and adds just the flash text line.
	fixedLines := 7
	if d.flash != "" {
		fixedLines += 1 // flash text line
	}

	// Budget for table + log content (excluding borders)
	available := d.height - fixedLines
	if available < 4 {
		available = 4
	}

	// Table gets up to 40% of screen, but never more rows than we have
	maxTableRows := d.height * 4 / 10
	if len(d.workers) == 0 {
		// "No workers" placeholder is 1 line
		maxTableRows = 1
	} else {
		// header + worker rows
		actualTable := len(d.workers) + 1
		if actualTable < maxTableRows {
			maxTableRows = actualTable
		}
	}

	// Log content gets the rest (minimum 1 line)
	logHeight := available - maxTableRows
	if logHeight < 1 {
		logHeight = 1
		// Shrink table to fit
		maxTableRows = available - logHeight
		if maxTableRows < 2 {
			maxTableRows = 2
		}
	}

	// Worker table
	if len(d.workers) == 0 {
		b.WriteString(mutedStyle.Render("  No workers. Press [n] to create one."))
		b.WriteString("\n")
	} else {
		// maxTableRows includes the header, so visible worker rows = maxTableRows - 1
		b.WriteString(d.renderTable(maxTableRows - 1))
	}

	// Flash message
	if d.flash != "" {
		b.WriteString("\n")
		if d.flashErr {
			b.WriteString(flashErrorStyle.Render(d.flash))
		} else {
			b.WriteString(flashStyle.Render(d.flash))
		}
	}

	// Log preview pane
	logTitle := "LOGS"
	if w := d.selectedWorker(); w != nil {
		logTitle = fmt.Sprintf("LOGS (%s)", w.ID)
	}

	logBox := logBorderStyle.
		Width(d.width - 4).
		Height(logHeight).
		Render(d.truncateLog(logHeight))

	b.WriteString("\n")
	b.WriteString(logTitleStyle.Render(logTitle))
	b.WriteString("\n")
	b.WriteString(logBox)

	// Help bar
	b.WriteString("\n")
	b.WriteString(d.renderHelp())

	return b.String()
}

func (d dashboard) renderTable(maxRows int) string {
	var b strings.Builder
	home, _ := os.UserHomeDir()

	// Header
	header := fmt.Sprintf("  %-8s %-10s %-24s %s", "ID", "STATUS", "DIRECTORY", "TASK")
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")

	// Determine visible window of workers around the cursor
	start := 0
	end := len(d.workers)
	if maxRows > 0 && len(d.workers) > maxRows {
		// Keep cursor visible within the window
		start = d.cursor - maxRows/2
		if start < 0 {
			start = 0
		}
		end = start + maxRows
		if end > len(d.workers) {
			end = len(d.workers)
			start = end - maxRows
		}
	}

	for i := start; i < end; i++ {
		w := d.workers[i]
		dir := w.Directory
		if home != "" {
			dir = strings.Replace(dir, home, "~", 1)
		}
		if len(dir) > 22 {
			dir = "..." + dir[len(dir)-19:]
		}

		task := w.Task
		maxTask := d.width - 50
		if maxTask < 10 {
			maxTask = 10
		}
		if len(task) > maxTask {
			task = task[:maxTask-3] + "..."
		}

		status := d.renderStatus(w)
		id := w.ID
		if i == d.cursor {
			id = selectedStyle.Render(id)
		}
		row := fmt.Sprintf("  %-8s %-10s %-24s %s", id, status, dir, task)
		b.WriteString(normalRowStyle.Render(row))
		b.WriteString("\n")
	}
	return b.String()
}

func (d dashboard) renderStatus(w *state.Worker) string {
	switch w.Status {
	case state.StatusWorking:
		return statusWorking.Render(d.spinner.View() + " working")
	case state.StatusPending:
		return statusPending.Render("◔ pending")
	case state.StatusDone:
		return statusDone.Render("✓ done")
	case state.StatusError:
		return statusError.Render("✗ error")
	}
	return string(w.Status)
}

func (d dashboard) renderHelp() string {
	var parts []string
	add := func(k, desc string) {
		parts = append(parts, helpKeyStyle.Render(k)+" "+helpDescStyle.Render(desc))
	}

	if w := d.selectedWorker(); w != nil {
		switch w.Status {
		case state.StatusPending:
			add("[a]", "approve")
			add("[d]", "deny")
		case state.StatusWorking:
			add("[x]", "kill")
			add("[r]", "resume")
		case state.StatusDone, state.StatusError:
			add("[r]", "resume")
			add("[c]", "clean")
		}
		add("[enter]", "logs")
	}

	add("[n]", "new")
	if d.filter != "" {
		add("[/]", "filter:"+d.filter)
	} else {
		add("[/]", "filter")
	}
	add("[q]", "quit")
	add("[?]", "help")
	return "  " + strings.Join(parts, "  ")
}

func (d dashboard) truncateLog(maxLines int) string {
	if d.logContent == "" {
		return mutedStyle.Render("No logs.")
	}
	lines := strings.Split(d.logContent, "\n")
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return strings.Join(lines, "\n")
}
