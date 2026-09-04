package tui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"
)

type appKeyMap struct {
	Next, Prev, Select, Inspect, Cancel, Approve, Deny key.Binding
	Filter, Refresh, Preview, Submit, JSON             key.Binding
	BlastIntent, BlastLive, Help, Quit                 key.Binding
}

func newAppKeyMap() appKeyMap {
	return appKeyMap{
		Next:        key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next tab")),
		Prev:        key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev tab")),
		Select:      key.NewBinding(key.WithKeys("j", "k", "up", "down"), key.WithHelp("j/k", "select")),
		Inspect:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "inspect")),
		Cancel:      key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "cancel run")),
		Approve:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "approve")),
		Deny:        key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "deny")),
		Filter:      key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Refresh:     key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Preview:     key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("ctrl+p", "preview intent")),
		Submit:      key.NewBinding(key.WithKeys("ctrl+enter"), key.WithHelp("ctrl+enter", "submit intent")),
		JSON:        key.NewBinding(key.WithKeys("ctrl+j"), key.WithHelp("ctrl+j", "JSON mode")),
		BlastIntent: key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("ctrl+b", "blast draft")),
		BlastLive:   key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "blast run")),
		Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit:        key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
}

func (k appKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Next, k.Select, k.Help, k.Quit}
}

func (k appKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Next, k.Prev, k.Select, k.Inspect},
		{k.Cancel, k.Approve, k.Deny, k.Filter, k.Refresh},
		{k.Preview, k.Submit, k.JSON, k.BlastIntent, k.BlastLive},
		{k.Help, k.Quit},
	}
}

func newHelp(theme Theme) help.Model {
	h := help.New()
	h.ShowAll = true
	if theme.NoColor {
		plain := lipgloss.NewStyle()
		h.Styles = help.Styles{
			ShortKey: plain, ShortDesc: plain, ShortSeparator: plain,
			FullKey: plain, FullDesc: plain, FullSeparator: plain,
			Ellipsis: plain,
		}
	}
	return h
}
