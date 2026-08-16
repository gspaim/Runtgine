package tui

import (
	"os"

	"charm.land/lipgloss/v2"
)

const (
	Space     = "#070B14"
	Panel     = "#101A2F"
	Starlight = "#8FB8FF"
	Violet    = "#9D8CFF"
	Telemetry = "#73E2C2"
	Amber     = "#FFB86B"
	Anomaly   = "#FF6B8A"
	Muted     = "#64748B"
)

type Theme struct {
	NoColor bool
	ASCII   bool
}

func DetectTheme() Theme {
	return Theme{
		NoColor: os.Getenv("NO_COLOR") != "",
		ASCII:   os.Getenv("TERM") == "dumb" || os.Getenv("RUNTGINE_ASCII") != "",
	}
}

func (t Theme) style(color string) lipgloss.Style {
	s := lipgloss.NewStyle()
	if !t.NoColor {
		s = s.Foreground(lipgloss.Color(color))
	}
	return s
}

func (t Theme) Header() lipgloss.Style {
	return t.style(Starlight).Bold(true)
}

func (t Theme) Muted() lipgloss.Style {
	return t.style(Muted)
}

func (t Theme) Active() lipgloss.Style {
	return t.style(Starlight).Bold(true)
}

func (t Theme) Selected() lipgloss.Style {
	s := t.style(Starlight).Bold(true)
	if !t.NoColor {
		s = s.Background(lipgloss.Color(Panel))
	}
	return s
}

func (t Theme) Panel(focused bool) lipgloss.Style {
	color := Muted
	if focused {
		color = Violet
	}
	border := lipgloss.RoundedBorder()
	if t.ASCII {
		border = lipgloss.ASCIIBorder()
	}
	s := lipgloss.NewStyle().Border(border).Padding(0, 1)
	if !t.NoColor {
		s = s.BorderForeground(lipgloss.Color(color))
	}
	return s
}

func (t Theme) Status(status string) lipgloss.Style {
	color := Muted
	switch status {
	case "running":
		color = Amber
	case "succeeded":
		color = Telemetry
	case "failed", "cancelled", "rejected":
		color = Anomaly
	case "accepted", "planned":
		color = Starlight
	}
	return t.style(color)
}

func (t Theme) Symbol(status string) string {
	if t.ASCII {
		switch status {
		case "running":
			return ">"
		case "succeeded":
			return "+"
		case "failed", "cancelled", "rejected":
			return "!"
		default:
			return "."
		}
	}
	switch status {
	case "running":
		return "◉"
	case "succeeded":
		return "●"
	case "failed", "cancelled", "rejected":
		return "×"
	default:
		return "○"
	}
}

func (t Theme) Star() string {
	if t.ASCII {
		return "*"
	}
	return "✦"
}
