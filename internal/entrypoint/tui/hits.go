package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/graph"
)

type hitRow struct {
	Source string
	Kind   string
	ID     string
	Score  int
}

func hitsFromEvents(events []event.Event) []hitRow {
	for i := len(events) - 1; i >= 0; i-- {
		if rows := hitsFromPayload(events[i].Payload); len(rows) > 0 {
			return rows
		}
	}
	return nil
}

func hitsFromPayload(payload map[string]any) []hitRow {
	if payload == nil {
		return nil
	}
	if nested, ok := payload["context_pack"].(map[string]any); ok {
		if rows := hitsFromPayload(nested); len(rows) > 0 {
			return rows
		}
	}
	var rows []hitRow
	rows = append(rows, decodeHitList("graph", payload["graph_hits"])...)
	rows = append(rows, decodeHitList("memory", payload["memory_hits"])...)
	rows = append(rows, decodeHitList("playbook", payload["playbook_hits"])...)
	return rows
}

func decodeHitList(source string, raw any) []hitRow {
	if raw == nil {
		return nil
	}
	if boxed, ok := raw.(map[string]any); ok {
		if items, exists := boxed["items"]; exists {
			raw = items
		}
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var items []map[string]any
	if err := json.Unmarshal(b, &items); err != nil {
		return nil
	}
	out := make([]hitRow, 0, len(items))
	for _, item := range items {
		kind, _ := item["kind"].(string)
		id, _ := item["id"].(string)
		if id == "" {
			id, _ = item["title"].(string)
		}
		if kind == "" && id == "" {
			continue
		}
		score := 0
		switch v := item["score"].(type) {
		case float64:
			score = int(v)
		case int:
			score = v
		}
		out = append(out, hitRow{Source: source, Kind: kind, ID: id, Score: score})
	}
	return out
}

func hitsFromGraph(hits graph.Hits) []hitRow {
	out := make([]hitRow, 0, len(hits.Items))
	for _, h := range hits.Items {
		out = append(out, hitRow{Source: "graph", Kind: h.Kind, ID: h.ID, Score: h.Score})
	}
	return out
}

func (m Model) renderHits(title string, rows []hitRow) string {
	if len(rows) == 0 {
		return title + "\n\nNo hits."
	}
	lines := []string{title}
	limit := 12
	if m.height < 24 {
		limit = 6
	}
	for i, row := range rows {
		if i >= limit {
			lines = append(lines, "…")
			break
		}
		score := ""
		if row.Score != 0 {
			score = fmt.Sprintf("  %d", row.Score)
		}
		lines = append(lines, fmt.Sprintf("%s  %s  %s%s", row.Source, row.Kind, row.ID, score))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderBlast() string {
	if m.blastErr != nil {
		return "BLAST\n\n" + m.theme.Status("failed").Render(truncate(m.blastErr.Error(), 80))
	}
	if m.blastRep == nil {
		return "BLAST\n\n" + m.theme.Muted().Render("Ctrl+b draft · b selected run")
	}
	rep := *m.blastRep
	risk := string(rep.Risk)
	if risk == "" {
		risk = "none"
	}
	head := fmt.Sprintf("BLAST  %s %s", m.theme.RiskSymbol(risk), strings.ToUpper(risk))
	head = m.theme.Risk(risk).Render(head)
	lines := []string{head, "", "touches"}
	if len(rep.Touches) == 0 {
		lines = append(lines, m.theme.Muted().Render("(none)"))
	} else {
		for i, tch := range rep.Touches {
			if i >= 8 {
				lines = append(lines, "…")
				break
			}
			lines = append(lines, fmt.Sprintf("%s  %s  %s  %s", tch.Mode, tch.Kind, tch.Key, tch.Capability))
		}
	}
	lines = append(lines, "", "conflicts")
	if len(rep.Conflicts) == 0 {
		lines = append(lines, m.theme.Muted().Render("(none)"))
	} else {
		for _, c := range rep.Conflicts {
			lines = append(lines, fmt.Sprintf("%s  %s  holder %s", c.Kind, c.Key, shortID(c.HolderRunID)))
		}
	}
	lines = append(lines, "", "affected")
	if len(rep.Affected) == 0 {
		lines = append(lines, m.theme.Muted().Render("(none)"))
	} else {
		for i, a := range rep.Affected {
			if i >= 8 {
				lines = append(lines, "…")
				break
			}
			lines = append(lines, fmt.Sprintf("%s  %s  %s", a.Kind, a.ID, a.Reason))
		}
	}
	return strings.Join(lines, "\n")
}
