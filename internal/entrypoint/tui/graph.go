package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gspaim/Runtgine/internal/core/graph"
)

type graphListItem struct {
	kind string
	id   string
}

func (g graphListItem) Title() string       { return g.kind }
func (g graphListItem) Description() string { return g.id }
func (g graphListItem) FilterValue() string { return g.kind + " " + g.id }

var graphKindOrder = []string{
	graph.KindPlayer,
	graph.KindCapability,
	graph.KindTask,
	graph.KindRun,
	graph.KindPath,
	graph.KindSymbol,
}

func (m Model) maybeLoadGraph() tea.Cmd {
	if m.tab != tabGraph || m.graphLoaded {
		return nil
	}
	return m.graphLoadCmd(false)
}

func (m Model) graphLoadCmd(refresh bool) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if refresh {
			if err := m.core.RefreshGraph(ctx); err != nil {
				return graphMsg{err: err}
			}
		}
		snap, err := m.core.GetGraphSnapshot(ctx)
		return graphMsg{snap: snap, err: err}
	}
}

func kindRank(kind string) int {
	for i, k := range graphKindOrder {
		if kind == k {
			return i
		}
	}
	return len(graphKindOrder)
}

func sortedGraphNodes(nodes []graph.Node) []graph.Node {
	out := append([]graph.Node(nil), nodes...)
	sort.Slice(out, func(i, j int) bool {
		ri, rj := kindRank(out[i].Kind), kindRank(out[j].Kind)
		if ri != rj {
			return ri < rj
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func (m Model) filteredGraphNodes() []graph.Node {
	filter := strings.ToLower(strings.TrimSpace(m.graphFilter))
	nodes := sortedGraphNodes(m.graph.Nodes)
	if filter == "" {
		return nodes
	}
	out := make([]graph.Node, 0, len(nodes))
	for _, n := range nodes {
		haystack := strings.ToLower(n.Kind + " " + n.ID)
		if strings.Contains(haystack, filter) {
			out = append(out, n)
		}
	}
	return out
}

func (m *Model) clampGraphSelection() {
	n := len(m.filteredGraphNodes())
	if n == 0 {
		m.graphSelected = 0
		return
	}
	if m.graphSelected >= n {
		m.graphSelected = n - 1
	}
	if m.graphSelected < 0 {
		m.graphSelected = 0
	}
}

func (m Model) renderGraph() string {
	nodes := m.filteredGraphNodes()
	header := m.graphHeader()
	if len(m.graph.Nodes) == 0 && m.graphErr == nil {
		body := header + "\n\nNo graph nodes."
		if m.graphFilter != "" || m.filtering {
			body += "\n\n" + m.graphFilterLine()
		}
		return m.theme.Panel(true).Render(body)
	}
	if len(nodes) == 0 {
		body := header + "\n\nNo matching nodes."
		body += "\n\n" + m.graphFilterLine()
		return m.theme.Panel(true).Render(body)
	}

	items := make([]list.Item, 0, len(nodes))
	var selected graph.Node
	for i, n := range nodes {
		if i == m.graphSelected {
			selected = n
		}
		items = append(items, graphListItem{kind: n.Kind, id: n.ID})
	}
	m.graphList.SetItems(items)
	if len(items) > 0 {
		m.graphList.Select(m.graphSelected)
	}
	m.graphList.SetSize(max(20, m.width/2-6), max(6, m.height-14))
	listBody := header + "\n" + m.graphList.View() + "\n" + m.graphFilterLine()
	if m.graphErr != nil {
		listBody += "\n" + m.theme.Status("failed").Render("error: "+m.graphErr.Error())
	}

	showDetail := m.width >= 80 || m.graphInspect
	if !showDetail || selected.Kind == "" {
		return m.theme.Panel(true).Render(listBody)
	}
	detailBody := m.graphDetail(selected)
	if m.width >= 120 {
		left := m.theme.Panel(true).Width(m.width/2 - 4).Render(listBody)
		right := m.theme.Panel(false).Width(m.width/2 - 4).Render(detailBody)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	}
	listPane := m.theme.Panel(true).Render(listBody)
	detail := m.theme.Panel(false).Render(detailBody)
	return lipgloss.JoinVertical(lipgloss.Left, listPane, detail)
}

func (m Model) graphHeader() string {
	counts := make(map[string]int, len(graphKindOrder))
	for _, n := range m.graph.Nodes {
		counts[n.Kind]++
	}
	parts := []string{
		fmt.Sprintf("GRAPH  nodes=%d edges=%d", len(m.graph.Nodes), len(m.graph.Edges)),
	}
	for _, kind := range graphKindOrder {
		parts = append(parts, fmt.Sprintf("%s=%d", kind, counts[kind]))
	}
	return strings.Join(parts, "  ")
}

func (m Model) graphFilterLine() string {
	query := "filter: " + m.graphFilter
	if m.filtering && m.tab == tabGraph {
		query += "▌"
	}
	return query
}

func (m Model) graphDetail(n graph.Node) string {
	attrs := "{}"
	if len(n.Attrs) > 0 {
		raw, err := json.MarshalIndent(n.Attrs, "", "  ")
		if err == nil {
			attrs = string(raw)
		}
	}
	arrow := "→"
	if m.theme.ASCII {
		arrow = "->"
	}
	var edges []string
	for _, e := range m.graph.Edges {
		if (e.From.Kind == n.Kind && e.From.ID == n.ID) || (e.To.Kind == n.Kind && e.To.ID == n.ID) {
			edges = append(edges, fmt.Sprintf("%s %s:%s %s %s:%s",
				e.Kind, e.From.Kind, e.From.ID, arrow, e.To.Kind, e.To.ID))
		}
	}
	if len(edges) == 0 {
		edges = []string{"(none)"}
	}
	return fmt.Sprintf("NODE\nkind  %s\nid    %s\n\nattrs\n%s\n\nedges\n%s",
		n.Kind, n.ID, truncateLines(attrs, max(3, m.height/4)), strings.Join(edges, "\n"))
}
