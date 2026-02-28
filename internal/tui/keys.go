package tui

import "github.com/charmbracelet/bubbles/key"

type dashboardKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	New      key.Binding
	Filter   key.Binding
	Approve  key.Binding
	Deny     key.Binding
	Kill     key.Binding
	Resume   key.Binding
	Clean    key.Binding
	CleanAll key.Binding
	Help     key.Binding
	Quit     key.Binding
}

var dashboardKeys = dashboardKeyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up", "ctrl+p"),
		key.WithHelp("k/ctrl+p", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down", "ctrl+n"),
		key.WithHelp("j/ctrl+n", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "logs"),
	),
	New: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Approve: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "approve"),
	),
	Deny: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "deny"),
	),
	Kill: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "kill"),
	),
	Resume: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "resume"),
	),
	Clean: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clean"),
	),
	CleanAll: key.NewBinding(
		key.WithKeys("C"),
		key.WithHelp("C", "clean all"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type logViewKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	HalfUp   key.Binding
	HalfDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Back     key.Binding
	Approve  key.Binding
	Deny     key.Binding
	Kill     key.Binding
	Resume   key.Binding
	Clean    key.Binding
	Quit     key.Binding
}

var logViewKeys = logViewKeyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up", "ctrl+p"),
		key.WithHelp("k/ctrl+p", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down", "ctrl+n"),
		key.WithHelp("j/ctrl+n", "down"),
	),
	HalfUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "half page up"),
	),
	HalfDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "half page down"),
	),
	Top: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "bottom"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Approve: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "approve"),
	),
	Deny: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "deny"),
	),
	Kill: key.NewBinding(
		key.WithKeys("x"),
		key.WithHelp("x", "kill"),
	),
	Resume: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "resume"),
	),
	Clean: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "clean"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

type formKeyMap struct {
	NextField key.Binding
	PrevField key.Binding
	Submit    key.Binding
	Cancel    key.Binding
	Toggle    key.Binding
}

var formKeys = formKeyMap{
	NextField: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	PrevField: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev field"),
	),
	Submit: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "create"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
}
