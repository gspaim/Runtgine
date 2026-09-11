package tui

import (
	"fmt"
	"strings"

	"github.com/gspaim/Runtgine/internal/entrypoint/liveout"
)

func (m Model) renderResult() string {
	lines := []string{"RESULT"}
	if errText := strings.TrimSpace(m.snapshot.Error); errText != "" {
		lines = append(lines, "", m.theme.Status("failed").Render(truncate(errText, max(24, m.width-8))))
	}
	results := liveout.FromEvents(m.snapshot.Events)
	if len(results) == 0 {
		if len(lines) == 1 {
			switch m.snapshot.Status {
			case "succeeded":
				lines = append(lines, "", "No step output.")
			case "failed", "cancelled":
				if m.snapshot.Error == "" {
					lines = append(lines, "", m.snapshot.Status)
				}
			default:
				lines = append(lines, "", m.theme.Muted().Render("waiting for step output…"))
			}
		}
		return strings.Join(lines, "\n")
	}
	limit := 16
	if m.height < 24 || m.width < 80 {
		limit = 8
	}
	used := 0
	for _, r := range results {
		if used >= limit {
			lines = append(lines, "…")
			break
		}
		lines = append(lines, "")
		mark := m.theme.Symbol("succeeded")
		if !r.OK {
			mark = m.theme.Symbol("failed")
		}
		lines = append(lines, fmt.Sprintf("%s %s", mark, r.StepID))
		used++
		body := truncateLines(r.Text, max(2, limit-used))
		for _, line := range strings.Split(body, "\n") {
			if used >= limit {
				lines = append(lines, "…")
				used++
				break
			}
			lines = append(lines, truncate(line, max(20, m.width-6)))
			used++
		}
	}
	return strings.Join(lines, "\n")
}
