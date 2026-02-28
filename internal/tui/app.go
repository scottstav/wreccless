package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/scottstav/wreccless/internal/config"
	"github.com/scottstav/wreccless/internal/hooks"
	"github.com/scottstav/wreccless/internal/state"
	"github.com/scottstav/wreccless/internal/worker"
)

type view int

const (
	viewDashboard view = iota
	viewLogView
	viewForm
)

type tickMsg time.Time
type flashDismissMsg struct{}

// ResumeInfo holds data needed to exec into claude after TUI exits.
type ResumeInfo struct {
	SessionID string
	Directory string
}

// App is the root Bubble Tea model.
type App struct {
	stateDir     string
	configPath   string
	view         view
	dashboard    dashboard
	logView      logView
	form         form
	width        int
	height       int
	showHelp     bool
	ResumeWorker *ResumeInfo // Set when TUI exits for resume
}

// NewApp creates a new TUI application model.
func NewApp(stateDir, configPath string) App {
	return App{
		stateDir:   stateDir,
		configPath: configPath,
		dashboard:  newDashboard(stateDir, configPath),
	}
}

func (a App) Init() tea.Cmd {
	return tea.Batch(
		a.dashboard.Init(),
		tickCmd(),
	)
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func flashCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(time.Time) tea.Msg {
		return flashDismissMsg{}
	})
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.dashboard.width = msg.Width
		a.dashboard.height = msg.Height
		a.dashboard.refreshWorkers()
		a.dashboard.refreshLogPreview()
		return a, nil

	case tickMsg:
		a.dashboard.refreshWorkers()
		a.dashboard.refreshLogPreview()
		if a.view == viewLogView {
			a.logView.refreshLog()
			a.logView.refreshWorker()
		}
		return a, tickCmd()

	case flashDismissMsg:
		a.dashboard.flash = ""
		return a, nil

	case backMsg:
		a.view = viewDashboard
		a.dashboard.refreshWorkers()
		a.dashboard.refreshLogPreview()
		return a, nil

	case cancelMsg:
		a.view = viewDashboard
		return a, nil

	case createMsg:
		flash := a.handleCreate(msg)
		a.view = viewDashboard
		a.dashboard.refreshWorkers()
		a.dashboard.refreshLogPreview()
		return a, flash

	case actionMsg:
		return a.handleAction(msg)
	}

	// Route to active view
	switch a.view {
	case viewDashboard:
		return a.updateDashboard(msg)
	case viewLogView:
		return a.updateLogView(msg)
	case viewForm:
		return a.updateForm(msg)
	}
	return a, nil
}

func (a App) updateDashboard(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		switch {
		case key.Matches(msg, dashboardKeys.Quit):
			return a, tea.Quit
		case key.Matches(msg, dashboardKeys.Help):
			a.showHelp = !a.showHelp
			return a, nil
		case key.Matches(msg, dashboardKeys.Enter):
			if w := a.dashboard.selectedWorker(); w != nil {
				a.logView = newLogView(a.stateDir, a.configPath, w, a.width, a.height)
				a.view = viewLogView
				return a, nil
			}
		case key.Matches(msg, dashboardKeys.New):
			a.form = newForm(a.width, a.height)
			a.view = viewForm
			return a, a.form.Init()

		case key.Matches(msg, dashboardKeys.Approve):
			if w := a.dashboard.selectedWorker(); w != nil && w.Status == state.StatusPending {
				return a, func() tea.Msg { return actionMsg{action: "approve", worker: w} }
			}
		case key.Matches(msg, dashboardKeys.Deny):
			if w := a.dashboard.selectedWorker(); w != nil && w.Status == state.StatusPending {
				return a, func() tea.Msg { return actionMsg{action: "deny", worker: w} }
			}
		case key.Matches(msg, dashboardKeys.Kill):
			if w := a.dashboard.selectedWorker(); w != nil && w.Status == state.StatusWorking {
				return a, func() tea.Msg { return actionMsg{action: "kill", worker: w} }
			}
		case key.Matches(msg, dashboardKeys.Resume):
			if w := a.dashboard.selectedWorker(); w != nil {
				return a, func() tea.Msg { return actionMsg{action: "resume", worker: w} }
			}
		case key.Matches(msg, dashboardKeys.Clean):
			if w := a.dashboard.selectedWorker(); w != nil && (w.Status == state.StatusDone || w.Status == state.StatusError) {
				return a, func() tea.Msg { return actionMsg{action: "clean", worker: w} }
			}
		case key.Matches(msg, dashboardKeys.CleanAll):
			return a, func() tea.Msg { return actionMsg{action: "cleanall", worker: nil} }
		case key.Matches(msg, dashboardKeys.Filter):
			filters := []string{"", "pending", "working", "done", "error"}
			cur := 0
			for i, f := range filters {
				if f == a.dashboard.filter {
					cur = i
					break
				}
			}
			a.dashboard.filter = filters[(cur+1)%len(filters)]
			a.dashboard.refreshWorkers()
			a.dashboard.refreshLogPreview()
			return a, nil
		}
	}

	var cmd tea.Cmd
	a.dashboard, cmd = a.dashboard.Update(msg)
	return a, cmd
}

func (a App) updateLogView(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(msg, logViewKeys.Quit) {
			return a, tea.Quit
		}
	}
	var cmd tea.Cmd
	a.logView, cmd = a.logView.Update(msg)
	return a, cmd
}

func (a App) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	a.form, cmd = a.form.Update(msg)
	return a, cmd
}

func (a *App) handleCreate(msg createMsg) tea.Cmd {
	cfg, _ := config.Load(a.configPath)

	now := time.Now()
	id := strconv.FormatInt(now.Unix(), 10)

	status := state.StatusWorking
	if msg.pending {
		status = state.StatusPending
	}

	w := &state.Worker{
		ID:        id,
		Status:    status,
		Directory: msg.dir,
		Task:      msg.task,
		Image:     msg.image,
		SessionID: uuid.New().String(),
		CreatedAt: &now,
	}

	if err := state.Write(a.stateDir, w); err != nil {
		a.dashboard.flash = fmt.Sprintf("Error: %v", err)
		a.dashboard.flashErr = true
		return flashCmd()
	}

	vars := hooks.Vars{ID: id, Task: msg.task, Dir: msg.dir, Status: string(status)}
	if msg.pending {
		hooks.Fire(cfg.Hooks.OnPending, vars)
		a.dashboard.flash = fmt.Sprintf("Worker %s created (pending)", id)
	} else {
		startedAt := time.Now()
		w.StartedAt = &startedAt
		state.Write(a.stateDir, w)

		cclBin, _ := os.Executable()
		if err := worker.SpawnRun(id, cclBin, a.configPath, a.stateDir); err != nil {
			a.dashboard.flash = fmt.Sprintf("Error spawning: %v", err)
			a.dashboard.flashErr = true
			return flashCmd()
		}
		hooks.Fire(cfg.Hooks.OnStart, vars)
		a.dashboard.flash = fmt.Sprintf("Worker %s created", id)
	}
	a.dashboard.flashErr = false
	return flashCmd()
}

func (a App) handleAction(msg actionMsg) (tea.Model, tea.Cmd) {
	cfg, _ := config.Load(a.configPath)
	w := msg.worker

	switch msg.action {
	case "approve":
		now := time.Now()
		w.Status = state.StatusWorking
		w.StartedAt = &now
		if w.SessionID == "" {
			w.SessionID = uuid.New().String()
		}
		state.Write(a.stateDir, w)

		cclBin, _ := os.Executable()
		if err := worker.SpawnRun(w.ID, cclBin, a.configPath, a.stateDir); err != nil {
			a.dashboard.flash = fmt.Sprintf("Error: %v", err)
			a.dashboard.flashErr = true
		} else {
			vars := hooks.Vars{ID: w.ID, Task: w.Task, Dir: w.Directory, Status: "working"}
			hooks.Fire(cfg.Hooks.OnStart, vars)
			a.dashboard.flash = fmt.Sprintf("Worker %s approved", w.ID)
			a.dashboard.flashErr = false
		}

	case "deny":
		state.Delete(a.stateDir, w.ID)
		vars := hooks.Vars{ID: w.ID, Task: w.Task, Dir: w.Directory, Status: "denied"}
		hooks.Fire(cfg.Hooks.OnKill, vars)
		a.dashboard.flash = fmt.Sprintf("Worker %s denied", w.ID)
		a.dashboard.flashErr = false
		if a.view == viewLogView {
			a.view = viewDashboard
		}

	case "kill":
		if w.PID > 0 {
			syscall.Kill(w.PID, syscall.SIGTERM)
		}
		state.Delete(a.stateDir, w.ID)
		vars := hooks.Vars{ID: w.ID, Task: w.Task, Dir: w.Directory, Status: "killed"}
		hooks.Fire(cfg.Hooks.OnKill, vars)
		a.dashboard.flash = fmt.Sprintf("Worker %s killed", w.ID)
		a.dashboard.flashErr = false
		if a.view == viewLogView {
			a.view = viewDashboard
		}

	case "clean":
		state.Delete(a.stateDir, w.ID)
		a.dashboard.flash = fmt.Sprintf("Worker %s cleaned", w.ID)
		a.dashboard.flashErr = false
		if a.view == viewLogView {
			a.view = viewDashboard
		}

	case "cleanall":
		workers, _ := state.List(a.stateDir)
		count := 0
		for _, w := range workers {
			if w.Status == state.StatusDone || w.Status == state.StatusError {
				state.Delete(a.stateDir, w.ID)
				count++
			}
		}
		a.dashboard.flash = fmt.Sprintf("Cleaned %d worker(s)", count)
		a.dashboard.flashErr = false

	case "resume":
		if w.SessionID == "" {
			a.dashboard.flash = "No session to resume"
			a.dashboard.flashErr = true
		} else {
			a.ResumeWorker = &ResumeInfo{
				SessionID: w.SessionID,
				Directory: w.Directory,
			}
			state.Delete(a.stateDir, w.ID)
			return a, tea.Quit
		}
	}

	a.dashboard.refreshWorkers()
	a.dashboard.refreshLogPreview()
	return a, flashCmd()
}

func (a App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	if a.showHelp {
		return a.renderHelp()
	}

	switch a.view {
	case viewDashboard:
		return a.dashboard.View()
	case viewLogView:
		return a.logView.View()
	case viewForm:
		formView := a.form.View()
		formHeight := lipgloss.Height(formView)
		padTop := (a.height - formHeight) / 2
		if padTop < 0 {
			padTop = 0
		}
		return strings.Repeat("\n", padTop) + lipgloss.PlaceHorizontal(a.width, lipgloss.Center, formView)
	}
	return ""
}

func (a App) renderHelp() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("wreccless — Keyboard Shortcuts"))
	b.WriteString("\n\n")

	section := func(title string, bindings [][2]string) {
		b.WriteString(headerStyle.Render(title))
		b.WriteString("\n")
		for _, bind := range bindings {
			b.WriteString(fmt.Sprintf("  %s  %s\n",
				helpKeyStyle.Width(20).Render(bind[0]),
				helpDescStyle.Render(bind[1]),
			))
		}
		b.WriteString("\n")
	}

	section("Dashboard", [][2]string{
		{"j / ctrl+n / ↓", "Next worker"},
		{"k / ctrl+p / ↑", "Previous worker"},
		{"Enter", "Open log viewer"},
		{"n", "New worker"},
		{"/", "Cycle status filter"},
		{"a", "Approve pending worker"},
		{"d", "Deny pending worker"},
		{"x", "Kill working worker"},
		{"r", "Resume worker session"},
		{"c", "Clean done/error worker"},
		{"C", "Clean all done/error"},
	})

	section("Log Viewer", [][2]string{
		{"j / ctrl+n", "Scroll down"},
		{"k / ctrl+p", "Scroll up"},
		{"ctrl+d", "Half page down"},
		{"ctrl+u", "Half page up"},
		{"g", "Jump to top"},
		{"G", "Jump to bottom"},
		{"Esc", "Back to dashboard"},
	})

	section("Global", [][2]string{
		{"q / ctrl+c", "Quit"},
		{"?", "Toggle this help"},
	})

	b.WriteString(mutedStyle.Render("Press ? to close"))
	return b.String()
}
