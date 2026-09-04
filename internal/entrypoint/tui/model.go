package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/gspaim/Runtgine/internal/core/api"
	"github.com/gspaim/Runtgine/internal/core/blast"
	"github.com/gspaim/Runtgine/internal/core/event"
	"github.com/gspaim/Runtgine/internal/core/graph"
	"github.com/gspaim/Runtgine/internal/core/runner"
	"github.com/gspaim/Runtgine/internal/core/task"
)

const (
	tabIntent = iota
	tabRuns
	tabLive
	tabBoard
	tabEvents
	tabGraph
	tabConfig
	tabCount
	intentHistoryCap = 10
)

var tabNames = []string{"INTENT", "RUNS", "LIVE", "BOARD", "EVENTS", "GRAPH", "CONFIG"}

type CoreAPI interface {
	ListRuns(context.Context, int) ([]api.RunSummary, error)
	GetRun(context.Context, string) (api.RunSnapshot, error)
	ListRecentEvents(context.Context, int) ([]event.Event, error)
	ConfigSnapshot() api.ConfigSnapshot
	Subscribe(int) (<-chan event.Event, func())
	CancelRun(string) error
	ApproveRun(string, string) error
	GetGraphSnapshot(context.Context) (graph.Snapshot, error)
	RefreshGraph(context.Context) error
	CompileIntent(context.Context, string, string, string) (task.Task, string, error)
	SubmitIntent(context.Context, string, string, string) (string, string, error)
	SubmitTask(context.Context, task.Task) (string, error)
	QueryHits(context.Context, graph.Query) graph.Hits
	BlastTask(context.Context, task.Task) (blast.Report, error)
}

type intentHistoryItem struct {
	RunID   string
	Summary string
}

type Model struct {
	core          CoreAPI
	theme         Theme
	width         int
	height        int
	tab           int
	selected      int
	runs          []api.RunSummary
	snapshot      api.RunSnapshot
	events        []event.Event
	config        api.ConfigSnapshot
	eventCh       <-chan event.Event
	filtering     bool
	filter        string
	graphFilter   string
	graph         graph.Snapshot
	graphSelected int
	graphInspect  bool
	graphLoaded   bool
	graphErr      error
	confirm       bool
	err           error
	spinner       spinner.Model
	progress      progress.Model
	intentDraft   string
	intentJSON    bool
	intentPreview string
	intentMethod  string
	intentDirty   bool
	intentConfirm bool
	intentBusy    bool
	intentHistory []intentHistoryItem
	textarea      textarea.Model
	runsTable     table.Model
	eventsTable   table.Model
	previewVP     viewport.Model
	detailVP      viewport.Model
	graphList     list.Model
	help          help.Model
	keys          appKeyMap
	helpOpen      bool
	intentHits    []hitRow
	blastRep      *blast.Report
	blastErr      error
	blastOpen     bool
}

type refreshMsg struct {
	runs     []api.RunSummary
	snapshot api.RunSnapshot
	events   []event.Event
	config   api.ConfigSnapshot
	err      error
}

type streamEventMsg event.Event
type cancelMsg struct{ err error }
type approveMsg struct{ err error }
type tickMsg time.Time
type streamClosedMsg struct{}
type graphMsg struct {
	snap graph.Snapshot
	err  error
}
type intentResultMsg struct {
	preview bool
	pretty  string
	method  string
	runID   string
	summary string
	hits    []hitRow
	err     error
}

type blastMsg struct {
	report blast.Report
	err    error
}

func New(core CoreAPI) (Model, func()) {
	theme := DetectTheme()
	spin := spinner.New(
		spinner.WithSpinner(spinner.MiniDot),
		spinner.WithStyle(theme.style(Amber)),
	)
	bar := progress.New(
		progress.WithFillCharacters('━', '─'),
		progress.WithColors(lipgloss.Color(Telemetry), lipgloss.Color(Violet)),
		progress.WithoutPercentage(),
	)
	if theme.ASCII {
		bar.Full, bar.Empty = '=', '-'
	}
	if theme.NoColor {
		bar.FullColor, bar.EmptyColor = lipgloss.NoColor{}, lipgloss.NoColor{}
	}
	ch, unsubscribe := core.Subscribe(256)
	ta := textarea.New()
	ta.Placeholder = "type intent (NL) or paste Task IR in JSON mode"
	ta.SetHeight(8)
	ta.SetWidth(40)
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	_ = ta.Focus()
	km := table.DefaultKeyMap()
	km.PageUp = key.NewBinding(key.WithKeys("pgup"), key.WithHelp("pgup", "page up"))
	km.HalfPageDown = key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "½ page down"))
	runsTbl := table.New(
		table.WithColumns([]table.Column{
			{Title: "ST", Width: 3},
			{Title: "STATUS", Width: 16},
			{Title: "RUN ID", Width: 14},
			{Title: "INTENT", Width: 36},
			{Title: "ELAPSED", Width: 8},
		}),
		table.WithFocused(true),
		table.WithHeight(12),
		table.WithKeyMap(km),
	)
	eventsTbl := table.New(
		table.WithColumns([]table.Column{
			{Title: "UTC", Width: 12},
			{Title: "TYPE", Width: 20},
			{Title: "RUN", Width: 13},
			{Title: "STEP", Width: 16},
			{Title: "PLAYER", Width: 12},
		}),
		table.WithFocused(false),
		table.WithHeight(10),
		table.WithKeyMap(km),
	)
	styles := table.DefaultStyles()
	styles.Header = theme.Muted().Bold(true)
	styles.Selected = theme.Selected()
	if theme.NoColor {
		styles.Header = lipgloss.NewStyle().Bold(true)
		styles.Selected = lipgloss.NewStyle().Bold(true)
		styles.Cell = lipgloss.NewStyle()
	}
	runsTbl.SetStyles(styles)
	eventsTbl.SetStyles(styles)
	delegate := list.NewDefaultDelegate()
	if theme.NoColor {
		plain := lipgloss.NewStyle()
		delegate.Styles = list.DefaultItemStyles{
			NormalTitle: plain, NormalDesc: plain,
			SelectedTitle: plain.Bold(true), SelectedDesc: plain.Bold(true),
			DimmedTitle: plain, DimmedDesc: plain, FilterMatch: plain,
		}
	}
	gl := list.New([]list.Item{}, delegate, 40, 12)
	gl.SetShowHelp(false)
	gl.SetShowStatusBar(false)
	gl.SetShowTitle(false)
	gl.SetShowPagination(false)
	gl.SetFilteringEnabled(false)
	gl.SetShowFilter(false)
	if theme.NoColor {
		plain := lipgloss.NewStyle()
		st := gl.Styles
		st.Title = plain
		st.TitleBar = plain
		st.PaginationStyle = plain
		st.ActivePaginationDot = plain
		st.InactivePaginationDot = plain
		st.DividerDot = plain
		gl.Styles = st
		ta.SetStyles(textarea.Styles{
			Focused: textarea.StyleState{Base: plain, Text: plain, CursorLine: plain, Placeholder: plain, Prompt: plain},
			Blurred: textarea.StyleState{Base: plain, Text: plain, CursorLine: plain, Placeholder: plain, Prompt: plain},
		})
	}
	return Model{
		core: core, theme: theme, width: 100, height: 30,
		eventCh: ch, spinner: spin, progress: bar,
		textarea: ta, runsTable: runsTbl, eventsTable: eventsTbl,
		previewVP: viewport.New(viewport.WithWidth(40), viewport.WithHeight(10)),
		detailVP:  viewport.New(viewport.WithWidth(40), viewport.WithHeight(10)),
		graphList: gl, help: newHelp(theme), keys: newAppKeyMap(),
	}, unsubscribe
}

func Run(core CoreAPI) error {
	model, unsubscribe := New(core)
	defer unsubscribe()
	_, err := tea.NewProgram(model).Run()
	return err
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.refreshCmd(), waitEvent(m.eventCh), m.spinner.Tick, nextTick())
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.progress.SetWidth(max(10, min(60, msg.Width-12)))
		m.resizeComponents()
	case refreshMsg:
		if msg.err != nil {
			m.err = msg.err
			break
		}
		m.err = nil
		m.runs, m.events, m.config = msg.runs, msg.events, msg.config
		if msg.snapshot.RunID != "" {
			m.snapshot = msg.snapshot
			for i, run := range m.runs {
				if run.RunID == msg.snapshot.RunID {
					m.selected = i
					break
				}
			}
		} else if m.selected >= len(m.runs) {
			m.selected = max(0, len(m.runs)-1)
		}
		if msg.snapshot.RunID == "" && len(m.runs) == 0 {
			m.snapshot = api.RunSnapshot{}
		}
	case graphMsg:
		if msg.err != nil {
			m.graphErr = msg.err
			break
		}
		m.graphErr = nil
		m.graph = msg.snap
		m.graphLoaded = true
		m.clampGraphSelection()
	case streamEventMsg:
		e := event.Event(msg)
		m.events = append([]event.Event{e}, m.events...)
		return m, tea.Batch(waitEvent(m.eventCh), m.refreshCmd())
	case streamClosedMsg:
		m.eventCh = nil
	case tickMsg:
		return m, tea.Batch(nextTick(), m.refreshCmd())
	case cancelMsg:
		m.confirm = false
		m.err = msg.err
		return m, m.refreshCmd()
	case approveMsg:
		m.err = msg.err
		return m, m.refreshCmd()
	case intentResultMsg:
		m.intentBusy = false
		if msg.err != nil {
			m.err = msg.err
			m.intentPreview = ""
			m.intentMethod = ""
			return m, nil
		}
		m.err = nil
		m.intentPreview = msg.pretty
		m.intentMethod = msg.method
		m.intentDirty = false
		if msg.preview {
			m.intentHits = msg.hits
			return m, nil
		}
		if msg.runID == "" {
			return m, nil
		}
		m.intentHistory = append([]intentHistoryItem{{RunID: msg.runID, Summary: msg.summary}}, m.intentHistory...)
		if len(m.intentHistory) > intentHistoryCap {
			m.intentHistory = m.intentHistory[:intentHistoryCap]
		}
		m.tab = tabLive
		return m, m.selectRunCmd(msg.runID)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd
	case blastMsg:
		if msg.err != nil {
			m.err = msg.err
			m.blastOpen = false
			m.blastRep = nil
			m.blastErr = msg.err
			return m, nil
		}
		m.err = nil
		m.blastErr = nil
		rep := msg.report
		m.blastRep = &rep
		m.blastOpen = true
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if m.filtering {
		target := &m.filter
		if m.tab == tabGraph {
			target = &m.graphFilter
		}
		switch key {
		case "esc", "enter":
			m.filtering = false
		case "backspace":
			if *target != "" {
				_, size := utf8.DecodeLastRuneInString(*target)
				*target = (*target)[:len(*target)-size]
			}
		default:
			if msg.Key().Text != "" {
				*target += msg.Key().Text
			}
		}
		if m.tab == tabGraph {
			m.clampGraphSelection()
		}
		return m, nil
	}

	if m.helpOpen {
		if key == "q" || key == "ctrl+c" {
			return m, tea.Quit
		}
		if key == "?" || key == "esc" {
			m.helpOpen = false
		}
		return m, nil
	}
	if key == "?" {
		m.helpOpen = true
		m.help.ShowAll = true
		return m, nil
	}

	if key == "q" || key == "ctrl+c" {
		return m, tea.Quit
	}
	if key == "tab" {
		m.tab = (m.tab + 1) % tabCount
		m.confirm = false
		m.filtering = false
		m.intentConfirm = false
		return m, m.maybeLoadGraph()
	}
	if key == "shift+tab" {
		m.tab = (m.tab + tabCount - 1) % tabCount
		m.confirm = false
		m.filtering = false
		m.intentConfirm = false
		return m, m.maybeLoadGraph()
	}

	if m.tab == tabIntent {
		return m.handleIntentKey(msg)
	}

	if key == "b" && m.tab == tabLive {
		return m, m.liveBlastCmd()
	}

	switch key {
	case "right", "l":
		m.tab = (m.tab + 1) % tabCount
		m.confirm = false
		m.filtering = false
		return m, m.maybeLoadGraph()
	case "left", "h":
		m.tab = (m.tab + tabCount - 1) % tabCount
		m.confirm = false
		m.filtering = false
		return m, m.maybeLoadGraph()
	case "up", "k":
		if m.tab == tabGraph {
			if m.graphSelected > 0 {
				m.graphSelected--
				m.graphInspect = false
			}
			return m, nil
		}
		if m.selected > 0 {
			m.selected--
			m.confirm = false
			return m, m.refreshCmd()
		}
	case "down", "j":
		if m.tab == tabGraph {
			if m.graphSelected+1 < len(m.filteredGraphNodes()) {
				m.graphSelected++
				m.graphInspect = false
			}
			return m, nil
		}
		if m.selected+1 < len(m.runs) {
			m.selected++
			m.confirm = false
			return m, m.refreshCmd()
		}
	case "enter":
		if m.tab == tabGraph {
			m.graphInspect = true
			return m, nil
		}
		if len(m.runs) > 0 {
			m.tab = tabLive
		}
	case "/":
		if m.tab == tabEvents || m.tab == tabGraph {
			m.filtering = true
		}
	case "r":
		if m.tab == tabGraph {
			return m, m.graphLoadCmd(true)
		}
		return m, m.refreshCmd()
	case "esc":
		m.confirm = false
	case "c":
		if !m.selectedActive() {
			break
		}
		if !m.confirm {
			m.confirm = true
			break
		}
		runID := m.runs[m.selected].RunID
		return m, func() tea.Msg { return cancelMsg{err: m.core.CancelRun(runID)} }
	case "a":
		if !m.selectedWaiting() {
			break
		}
		runID := m.runs[m.selected].RunID
		return m, func() tea.Msg { return approveMsg{err: m.core.ApproveRun(runID, runner.DecisionGrant)} }
	case "d":
		if !m.selectedWaiting() {
			break
		}
		runID := m.runs[m.selected].RunID
		return m, func() tea.Msg { return approveMsg{err: m.core.ApproveRun(runID, runner.DecisionDeny)} }
	}
	return m, nil
}

func (m Model) handleIntentKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	switch key {
	case "ctrl+p":
		return m, m.intentCmd(true)
	case "ctrl+enter":
		return m, m.intentCmd(false)
	case "ctrl+b":
		return m, m.intentBlastCmd()
	case "ctrl+j":
		m.intentJSON = !m.intentJSON
		return m, nil
	case "esc":
		if m.intentConfirm {
			m.intentDraft = ""
			m.intentPreview = ""
			m.intentMethod = ""
			m.intentDirty = false
			m.intentConfirm = false
			m.err = nil
			return m, nil
		}
		if strings.TrimSpace(m.intentDraft) != "" || m.intentPreview != "" {
			m.intentConfirm = true
			return m, nil
		}
		return m, nil
	case "backspace":
		m.intentConfirm = false
		if m.intentDraft != "" {
			_, size := utf8.DecodeLastRuneInString(m.intentDraft)
			m.intentDraft = m.intentDraft[:len(m.intentDraft)-size]
			m.intentDirty = true
		}
		return m, nil
	case "enter":
		m.intentConfirm = false
		m.intentDraft += "\n"
		m.intentDirty = true
		return m, nil
	}
	if msg.Key().Text != "" {
		m.intentConfirm = false
		m.intentDraft += msg.Key().Text
		m.intentDirty = true
	}
	return m, nil
}

func (m Model) intentCmd(preview bool) tea.Cmd {
	text := m.intentDraft
	jsonMode := m.intentJSON
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if jsonMode {
			return m.submitJSONIntent(ctx, text, preview)
		}
		if preview {
			tk, method, err := m.core.CompileIntent(ctx, text, "tui", "intent")
			if err != nil {
				return intentResultMsg{preview: true, err: err}
			}
			pretty, _ := json.MarshalIndent(tk, "", "  ")
			hits := hitsFromGraph(m.core.QueryHits(ctx, graph.Query{Text: text}))
			return intentResultMsg{preview: true, pretty: string(pretty), method: method, summary: tk.Intent.Summary, hits: hits}
		}
		runID, method, err := m.core.SubmitIntent(ctx, text, "tui", "intent")
		if err != nil {
			return intentResultMsg{err: err, method: method}
		}
		tk, _, _ := m.core.CompileIntent(ctx, text, "tui", "intent")
		pretty, _ := json.MarshalIndent(tk, "", "  ")
		return intentResultMsg{pretty: string(pretty), method: method, runID: runID, summary: tk.Intent.Summary}
	}
}

func (m Model) submitJSONIntent(ctx context.Context, raw string, preview bool) tea.Msg {
	trimmed := strings.TrimSpace(raw)
	if err := task.ValidateDocument([]byte(trimmed)); err != nil {
		return intentResultMsg{preview: preview, err: err}
	}
	tk, err := task.Parse([]byte(trimmed))
	if err != nil {
		return intentResultMsg{preview: preview, err: err}
	}
	if tk.Source.EntryPoint == "" {
		tk.Source.EntryPoint = "tui"
	}
	pretty, _ := json.MarshalIndent(tk, "", "  ")
	if preview {
		hits := hitsFromGraph(m.core.QueryHits(ctx, graph.Query{Text: raw}))
		return intentResultMsg{preview: true, pretty: string(pretty), method: "json", summary: tk.Intent.Summary, hits: hits}
	}
	runID, err := m.core.SubmitTask(ctx, tk)
	if err != nil {
		return intentResultMsg{pretty: string(pretty), method: "json", err: err}
	}
	return intentResultMsg{pretty: string(pretty), method: "json", runID: runID, summary: tk.Intent.Summary}
}

func (m Model) intentBlastCmd() tea.Cmd {
	text := m.intentDraft
	jsonMode := m.intentJSON
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		var tk task.Task
		var err error
		if jsonMode {
			if err = task.ValidateDocument([]byte(strings.TrimSpace(text))); err != nil {
				return blastMsg{err: err}
			}
			tk, err = task.Parse([]byte(strings.TrimSpace(text)))
		} else {
			tk, _, err = m.core.CompileIntent(ctx, text, "tui", "intent")
		}
		if err != nil {
			return blastMsg{err: err}
		}
		rep, err := m.core.BlastTask(ctx, tk)
		return blastMsg{report: rep, err: err}
	}
}

func (m Model) liveBlastCmd() tea.Cmd {
	raw := append(json.RawMessage(nil), m.snapshot.Task...)
	return func() tea.Msg {
		if len(bytesTrim(raw)) == 0 {
			return blastMsg{err: fmt.Errorf("no task on selected run")}
		}
		tk, err := task.Parse(raw)
		if err != nil {
			return blastMsg{err: err}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		rep, err := m.core.BlastTask(ctx, tk)
		return blastMsg{report: rep, err: err}
	}
}

func bytesTrim(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

func (m *Model) resizeComponents() {
	bodyW := max(20, m.width-6)
	bodyH := max(6, m.height-10)
	m.textarea.SetWidth(max(20, bodyW/2-4))
	m.textarea.SetHeight(min(10, max(4, bodyH/3)))
	m.previewVP.SetWidth(max(20, bodyW/2-2))
	m.previewVP.SetHeight(max(6, bodyH-2))
	m.detailVP.SetWidth(max(20, bodyW/2-2))
	m.detailVP.SetHeight(max(6, bodyH/2))
	m.runsTable.SetWidth(bodyW)
	m.runsTable.SetHeight(max(6, bodyH-2))
	m.eventsTable.SetWidth(max(20, bodyW/2-2))
	m.eventsTable.SetHeight(max(6, bodyH-4))
	m.graphList.SetSize(max(20, bodyW/2-4), max(6, bodyH-4))
	m.help.SetWidth(max(20, m.width-4))
}

func (m Model) selectRunCmd(runID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		runs, err := m.core.ListRuns(ctx, 200)
		if err != nil {
			return refreshMsg{err: err}
		}
		snapshot, err := m.core.GetRun(ctx, runID)
		if err != nil {
			return refreshMsg{runs: runs, err: err}
		}
		events, err := m.core.ListRecentEvents(ctx, 300)
		return refreshMsg{
			runs: runs, snapshot: snapshot, events: events,
			config: m.core.ConfigSnapshot(), err: err,
		}
	}
}

func (m Model) refreshCmd() tea.Cmd {
	selectedID := ""
	if len(m.runs) > 0 && m.selected < len(m.runs) {
		selectedID = m.runs[m.selected].RunID
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		runs, err := m.core.ListRuns(ctx, 200)
		if err != nil {
			return refreshMsg{err: err}
		}
		if selectedID == "" && len(runs) > 0 {
			selectedID = runs[0].RunID
		}
		var snapshot api.RunSnapshot
		if selectedID != "" {
			snapshot, err = m.core.GetRun(ctx, selectedID)
			if err != nil {
				return refreshMsg{err: err}
			}
		}
		events, err := m.core.ListRecentEvents(ctx, 300)
		return refreshMsg{
			runs: runs, snapshot: snapshot, events: events,
			config: m.core.ConfigSnapshot(), err: err,
		}
	}
}

func waitEvent(ch <-chan event.Event) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		e, ok := <-ch
		if !ok {
			return streamClosedMsg{}
		}
		return streamEventMsg(e)
	}
}

func nextTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) selectedActive() bool {
	if len(m.runs) == 0 || m.selected >= len(m.runs) {
		return false
	}
	switch m.runs[m.selected].Status {
	case "accepted", "planned", "running", "waiting_approval":
		return true
	default:
		return false
	}
}

func (m Model) selectedWaiting() bool {
	if len(m.runs) == 0 || m.selected >= len(m.runs) {
		return false
	}
	return m.runs[m.selected].Status == "waiting_approval"
}

func (m Model) View() tea.View {
	header := m.renderHeader()
	tabs := m.renderTabs()
	body := m.renderBody()
	if m.helpOpen {
		body = m.renderHelpOverlay()
	}
	footer := m.renderFooter()
	content := lipgloss.JoinVertical(lipgloss.Left, header, tabs, body, footer)
	view := tea.NewView(content)
	view.AltScreen = true
	view.WindowTitle = "Runtgine — Constellation Mission Control"
	if !m.theme.NoColor {
		view.BackgroundColor = lipgloss.Color(Space)
		view.ForegroundColor = lipgloss.Color(Starlight)
	}
	return view
}

func (m Model) renderHelpOverlay() string {
	m.help.ShowAll = true
	m.help.SetWidth(max(20, m.width-8))
	body := "HELP  Constellation Mission Control\n\n" +
		m.theme.Muted().Render("q still quits") + "\n\n" +
		m.help.View(m.keys)
	return m.theme.Panel(true).Render(body)
}

func (m Model) renderHeader() string {
	running := 0
	waiting := 0
	for _, run := range m.runs {
		if run.Status == "running" {
			running++
		}
		if run.Status == "waiting_approval" {
			waiting++
		}
	}
	signal := "local connected"
	if waiting > 0 {
		signal = fmt.Sprintf("%s %d waiting", m.spinner.View(), waiting)
	} else if running > 0 {
		signal = fmt.Sprintf("%s %d active", m.spinner.View(), running)
	}
	title := fmt.Sprintf("%s RUNTGINE / CONSTELLATION MISSION CONTROL", m.theme.Star())
	meta := fmt.Sprintf("workspace %s · %s", shortPath(m.config.WorkspaceRoot, max(20, m.width-40)), signal)
	return m.theme.Header().Render(title) + "\n" + m.theme.Muted().Render(meta)
}

func (m Model) renderTabs() string {
	compact := m.width < 80
	parts := make([]string, 0, len(tabNames))
	for i, name := range tabNames {
		if compact {
			name = name[:1]
		}
		label := " " + name + " "
		if i == m.tab {
			label = m.theme.Selected().Render("[" + label + "]")
		} else {
			label = m.theme.Muted().Render(" " + label + " ")
		}
		parts = append(parts, label)
	}
	return strings.Join(parts, " ")
}

func (m Model) renderBody() string {
	var body string
	switch m.tab {
	case tabIntent:
		body = m.renderIntent()
	case tabRuns:
		body = m.renderRuns()
	case tabLive:
		body = m.renderLive()
	case tabBoard:
		body = m.renderBoard()
	case tabEvents:
		body = m.renderEvents()
	case tabGraph:
		body = m.renderGraph()
	case tabConfig:
		body = m.renderConfig()
	}
	height := max(8, m.height-7)
	return lipgloss.NewStyle().Width(max(20, m.width-2)).Height(height).Render(body)
}

func (m Model) renderIntent() string {
	mode := "NL"
	if m.intentJSON {
		mode = "JSON"
	}
	m.textarea.SetValue(m.intentDraft)
	m.textarea.SetWidth(max(20, m.width/2-8))
	m.textarea.SetHeight(min(10, max(4, m.height/4)))
	input := m.textarea.View()
	preview := m.intentPreview
	if preview == "" {
		preview = m.theme.Muted().Render("Ctrl+p preview · Ctrl+Enter submit · Ctrl+b blast")
	} else if m.intentMethod != "" {
		preview = "method " + m.intentMethod + "\n" + preview
	}
	hits := m.renderHits("HITS", m.intentHits)
	blast := m.renderBlast()
	if !m.blastOpen {
		blast = m.theme.Muted().Render("BLAST  Ctrl+b")
	}
	histLines := []string{"SESSION"}
	if len(m.intentHistory) == 0 {
		histLines = append(histLines, m.theme.Muted().Render("no submits this session"))
	} else {
		for i, h := range m.intentHistory {
			if i >= 5 {
				break
			}
			histLines = append(histLines, fmt.Sprintf("%s  %s", shortID(h.RunID), truncate(h.Summary, 36)))
		}
	}
	m.previewVP.SetContent(preview)
	title := fmt.Sprintf("INTENT  Mission Brief  mode %s", mode)
	left := title + "\n\n" + input
	right := "PREVIEW\n\n" + m.previewVP.View() + "\n\n" + hits + "\n\n" + blast + "\n\n" + strings.Join(histLines, "\n")
	if m.intentConfirm {
		left += "\n\n" + m.theme.Status("cancelled").Render("Clear draft? esc again confirms")
	}
	if m.width < 80 {
		return m.theme.Panel(true).Render(left + "\n\n" + right)
	}
	lw := max(28, m.width/2-4)
	rw := max(28, m.width-lw-6)
	return lipgloss.JoinHorizontal(lipgloss.Top,
		m.theme.Panel(true).Width(lw).Render(left),
		m.theme.Panel(false).Width(rw).Render(right),
	)
}

func (m Model) renderRuns() string {
	if len(m.runs) == 0 {
		return m.theme.Panel(true).Render("RUNS\n\nNo runs recorded.")
	}
	rows := make([]table.Row, 0, len(m.runs))
	for _, run := range m.runs {
		elapsed := run.UpdatedAt.Sub(run.CreatedAt)
		if run.Status == "running" || run.Status == "waiting_approval" {
			elapsed = time.Since(run.CreatedAt)
		}
		rows = append(rows, table.Row{
			m.theme.Symbol(run.Status),
			run.Status,
			shortID(run.RunID),
			truncate(run.Summary, 36),
			duration(elapsed),
		})
	}
	m.runsTable.SetRows(rows)
	m.runsTable.SetCursor(m.selected)
	m.runsTable.SetHeight(max(6, m.height-10))
	m.runsTable.SetWidth(max(40, m.width-6))
	return m.theme.Panel(true).Render("RUNS\n" + m.runsTable.View())
}

func (m Model) renderLive() string {
	if m.snapshot.RunID == "" {
		return m.theme.Panel(true).Render("LIVE\n\nSelect a run in RUNS.")
	}
	var t task.Task
	_ = json.Unmarshal(m.snapshot.Task, &t)
	state := stepStates(t, m.snapshot.Events)
	completed := 0
	for _, status := range state {
		if status == "succeeded" {
			completed++
		}
	}
	ratio := 0.0
	if len(t.Steps) > 0 {
		ratio = float64(completed) / float64(len(t.Steps))
	}
	title := fmt.Sprintf("LIVE  %s  %s  %s",
		shortID(m.snapshot.RunID), m.theme.Symbol(m.snapshot.Status), m.snapshot.Status)
	trajectory := m.renderTrajectory(t, state)
	progressLine := fmt.Sprintf("progress %s %d/%d", m.progress.ViewAs(ratio), completed, len(t.Steps))
	telemetry := latestTelemetry(m.snapshot.Events)
	currentStep, player, capability, contextBytes := currentExecution(t, m.snapshot.Events, state)
	if p := m.snapshot.PendingApproval; p != nil {
		currentStep, capability, player = p.StepID, p.Capability, p.Player
		contextBytes = "-"
	}
	detail := fmt.Sprintf("intent       %s\nsource       %s\ncurrent step %s\nplayer       %s\ncapability   %s\nContextPack  %s\ntelemetry    %s",
		truncate(t.Intent.Summary, max(20, m.width-24)),
		t.Source.EntryPoint, currentStep, player, capability, contextBytes, telemetry)
	hits := m.renderHits("HITS", hitsFromEvents(m.snapshot.Events))
	blast := m.renderBlast()
	if !m.blastOpen {
		blast = m.theme.Muted().Render("BLAST  press b")
	}
	m.detailVP.SetContent(detail)
	rightBody := "CURRENT RUN\n\n" + m.detailVP.View() + "\n\n" + hits + "\n\n" + blast

	if m.width >= 120 {
		left := m.theme.Panel(true).Width(m.width/2 - 4).Render(title + "\n\n" + trajectory + "\n\n" + progressLine)
		right := m.theme.Panel(false).Width(m.width/2 - 4).Render(rightBody)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	}
	return m.theme.Panel(true).Render(title + "\n\n" + trajectory + "\n\n" + progressLine + "\n\n" + rightBody)
}

func (m Model) renderTrajectory(t task.Task, states map[string]string) string {
	if len(t.Steps) == 0 {
		return "No steps."
	}
	parts := make([]string, 0, len(t.Steps))
	for _, step := range t.Steps {
		status := states[step.StepID]
		label := fmt.Sprintf("%s %s\n%s", m.theme.Symbol(status), step.StepID, step.Capability)
		if m.width >= 120 {
			label = fmt.Sprintf("%s %s · %s", m.theme.Symbol(status), step.StepID, step.Capability)
		}
		parts = append(parts, m.theme.Status(status).Render(label))
	}
	if m.width < 120 {
		return strings.Join(parts, "\n  |\n")
	}
	return strings.Join(parts, "  ---  ")
}

func (m Model) renderBoard() string {
	lanes := map[string][]api.RunSummary{"INTAKE": {}, "IN FLIGHT": {}, "LANDED": {}}
	for _, run := range m.runs {
		if run.Source != "board" {
			continue
		}
		lane := "LANDED"
		switch run.Status {
		case "accepted", "planned":
			lane = "INTAKE"
		case "running", "waiting_approval":
			lane = "IN FLIGHT"
		}
		lanes[lane] = append(lanes[lane], run)
	}
	names := []string{"INTAKE", "IN FLIGHT", "LANDED"}
	panels := make([]string, 0, 3)
	for _, name := range names {
		lines := []string{fmt.Sprintf("%s  %d", name, len(lanes[name]))}
		for _, run := range lanes[name] {
			lines = append(lines, fmt.Sprintf(
				"%s %s\n%s\nrun %s",
				m.theme.Symbol(run.Status), truncate(run.SourceRef, 24),
				truncate(run.Summary, 30), shortID(run.RunID),
			))
		}
		if len(lines) == 1 {
			lines = append(lines, "\nNo board runs.")
		}
		width := max(22, m.width/3-3)
		panels = append(panels, m.theme.Panel(name == "IN FLIGHT").Width(width).Render(strings.Join(lines, "\n\n")))
	}
	if m.width < 120 {
		return lipgloss.JoinVertical(lipgloss.Left, panels...)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, panels...)
}

func (m Model) renderEvents() string {
	filter := strings.ToLower(strings.TrimSpace(m.filter))
	rows := make([]table.Row, 0, len(m.events))
	var selected event.Event
	for _, e := range m.events {
		step := "-"
		if e.StepID != nil {
			step = *e.StepID
		}
		player := payloadString(e.Payload, "player")
		if player == "" {
			player = "-"
		}
		haystack := strings.ToLower(strings.Join([]string{e.Type, e.RunID, step, player}, " "))
		if filter != "" && !strings.Contains(haystack, filter) {
			continue
		}
		if selected.EventID == "" {
			selected = e
		}
		rows = append(rows, table.Row{
			e.TS.UTC().Format("15:04:05.000"),
			truncate(e.Type, 20),
			shortID(e.RunID),
			truncate(step, 16),
			truncate(player, 12),
		})
	}
	m.eventsTable.SetRows(rows)
	m.eventsTable.SetHeight(max(6, m.height-14))
	query := "filter: " + m.filter
	if m.filtering {
		query += "▌"
	}
	listPane := m.theme.Panel(true).Render("EVENTS\n" + m.eventsTable.View() + "\n\n" + query)
	if selected.EventID == "" {
		return listPane
	}
	payload, _ := json.MarshalIndent(selected.Payload, "", "  ")
	m.detailVP.SetContent(string(payload))
	detail := m.theme.Panel(false).Render("PAYLOAD\n" + m.detailVP.View())
	if m.width >= 120 {
		return lipgloss.JoinHorizontal(lipgloss.Top, listPane, detail)
	}
	return lipgloss.JoinVertical(lipgloss.Left, listPane, detail)
}

func (m Model) renderConfig() string {
	connected := func(v bool) string {
		if v {
			return "configured (secret masked)"
		}
		return "not configured"
	}
	content := fmt.Sprintf(`CONFIG  READ-ONLY

workspace             %s
SQLite               %s
log level            %s
max concurrent runs  %d
LLM backend          %s
LLM credentials      %s
GitHub credentials   %s

precedence
%s`,
		m.config.WorkspaceRoot, m.config.DBPath, m.config.LogLevel,
		m.config.MaxConcurrentRuns, m.config.LLMBackend,
		connected(m.config.LLMConnected), connected(m.config.GitHubConnected),
		m.config.Precedence)
	return m.theme.Panel(true).Render(content)
}

func (m Model) renderFooter() string {
	hint := "tab/shift+tab navigate · j/k select · enter inspect · c cancel · / filter · r refresh · ? help · q quit"
	if m.theme.ASCII {
		hint = "tab navigate | j/k select | enter inspect | c cancel | / filter | r refresh | ? help | q quit"
	}
	if m.helpOpen {
		hint = "?/esc close help · q quit"
		if m.theme.ASCII {
			hint = "?/esc close help | q quit"
		}
	} else if m.tab == tabIntent {
		hint = "tab/shift+tab navigate · Ctrl+p preview · Ctrl+Enter submit · Ctrl+b blast · Ctrl+j JSON · ? help · q quit"
		if m.theme.ASCII {
			hint = "tab navigate | Ctrl+p preview | Ctrl+Enter submit | Ctrl+b blast | Ctrl+j JSON | ? help | q quit"
		}
	}
	if m.tab == tabGraph && !m.helpOpen {
		hint = "tab/shift+tab navigate · j/k select · enter inspect · / filter · r refresh graph · ? help · q quit"
		if m.theme.ASCII {
			hint = "tab navigate | j/k select | enter inspect | / filter | r refresh graph | ? help | q quit"
		}
	}
	if m.tab == tabLive && !m.helpOpen && !m.selectedWaiting() {
		hint = "tab/shift+tab navigate · j/k select · b blast · c cancel · ? help · q quit"
		if m.theme.ASCII {
			hint = "tab navigate | j/k select | b blast | c cancel | ? help | q quit"
		}
	}
	if m.selectedWaiting() && m.tab != tabGraph && m.tab != tabIntent {
		hint = "tab/shift+tab navigate · j/k select · a approve · d deny · c cancel · q quit"
		if m.theme.ASCII {
			hint = "tab navigate | j/k select | a approve | d deny | c cancel | q quit"
		}
	}
	if m.confirm {
		hint = m.theme.Status("cancelled").Render("Confirm cancellation: press c again; esc aborts")
	}
	if m.err != nil {
		hint += "  |  " + m.theme.Status("failed").Render("error: "+truncate(m.err.Error(), 60))
	}
	if m.graphErr != nil && m.tab == tabGraph {
		hint += "  |  " + m.theme.Status("failed").Render("error: "+truncate(m.graphErr.Error(), 60))
	}
	return m.theme.Muted().Render(truncate(hint, max(20, m.width-1)))
}

func stepStates(t task.Task, events []event.Event) map[string]string {
	state := make(map[string]string, len(t.Steps))
	for _, step := range t.Steps {
		state[step.StepID] = "planned"
	}
	for _, e := range events {
		if e.StepID == nil {
			continue
		}
		switch e.Type {
		case event.TypeStepStarted:
			state[*e.StepID] = "running"
		case event.TypeStepSucceeded:
			state[*e.StepID] = "succeeded"
		case event.TypeStepFailed:
			state[*e.StepID] = "failed"
		case event.TypeRunWaitingApproval:
			state[*e.StepID] = "waiting_approval"
		}
	}
	return state
}

func latestTelemetry(events []event.Event) string {
	if len(events) == 0 {
		return "No events"
	}
	sorted := append([]event.Event(nil), events...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].TS.After(sorted[j].TS) })
	e := sorted[0]
	step := ""
	if e.StepID != nil {
		step = " · " + *e.StepID
	}
	return e.TS.UTC().Format("15:04:05") + " · " + e.Type + step
}

func currentExecution(
	t task.Task,
	events []event.Event,
	states map[string]string,
) (stepID, player, capability, contextSize string) {
	stepID, player, capability, contextSize = "-", "-", "-", "-"
	for _, step := range t.Steps {
		if states[step.StepID] != "running" && states[step.StepID] != "waiting_approval" {
			continue
		}
		stepID, capability = step.StepID, step.Capability
		for i := len(events) - 1; i >= 0; i-- {
			e := events[i]
			if e.StepID == nil || *e.StepID != step.StepID || e.Type != event.TypeStepStarted {
				continue
			}
			if value := payloadString(e.Payload, "player"); value != "" {
				player = value
			}
			if value := payloadString(e.Payload, "capability"); value != "" {
				capability = value
			}
			if value, ok := e.Payload["context_bytes"]; ok {
				contextSize = fmt.Sprintf("%v bytes", value)
			}
			break
		}
		break
	}
	return
}

func payloadString(payload map[string]any, key string) string {
	value, _ := payload[key].(string)
	return value
}

func shortID(id string) string {
	if len(id) <= 12 {
		return id
	}
	return id[:12]
}

func shortPath(path string, width int) string {
	if len(path) <= width {
		return path
	}
	return "…" + path[len(path)-width+1:]
}

func truncate(value string, width int) string {
	if width < 2 || len([]rune(value)) <= width {
		return value
	}
	runes := []rune(value)
	return string(runes[:width-1]) + "…"
}

func truncateLines(value string, limit int) string {
	lines := strings.Split(value, "\n")
	if len(lines) <= limit {
		return value
	}
	return strings.Join(lines[:limit], "\n") + "\n…"
}

func duration(value time.Duration) string {
	if value < 0 {
		value = 0
	}
	if value < time.Second {
		return "<1s"
	}
	return value.Round(time.Second).String()
}
