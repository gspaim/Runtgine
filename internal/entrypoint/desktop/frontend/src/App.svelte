<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import {
    compileIntent,
    compileTaskJSON,
    submitIntent,
    submitTaskJSON,
    getRun,
    listRuns,
    listBoardRuns,
    listRecentEvents,
    configSnapshot,
    getGraphSnapshot,
    refreshGraph,
    listLessons,
    approveLesson,
    rejectLesson,
    cancelRun,
    approveRun,
    denyRun,
    onRuntimeEvent,
    errMessage,
    shortID,
    boardLane,
    type RunSnapshot,
    type RunSummary,
    type RuntimeEvent,
    type ConfigSnapshot,
    type GraphSnapshot,
    type Lesson,
  } from "./lib/core";

  const views = ["INTENT", "RUNS", "LIVE", "BOARD", "EVENTS", "GRAPH", "CONFIG"] as const;
  type View = (typeof views)[number];
  const laneNames = ["INTAKE", "IN FLIGHT", "LANDED"] as const;

  let view = $state<View>("INTENT");
  let draft = $state("git status");
  let jsonMode = $state(false);
  let preview = $state("");
  let method = $state("");
  let err = $state("");
  let busy = $state(false);
  let run = $state<RunSnapshot | null>(null);
  let liveEvents = $state<string[]>([]);
  let runs = $state<RunSummary[]>([]);
  let boardRuns = $state<RunSummary[]>([]);
  let events = $state<RuntimeEvent[]>([]);
  let eventFilter = $state("");
  let graph = $state<GraphSnapshot>({ nodes: [], edges: [] });
  let graphFilter = $state("");
  let config = $state<ConfigSnapshot | null>(null);
  let lessons = $state<Lesson[]>([]);
  let lessonFilter = $state("pending");
  let stopEvents: (() => void) | undefined;

  onMount(() => {
    stopEvents = onRuntimeEvent((ev) => {
      if (ev?.type) {
        events = [ev, ...events].slice(0, 300);
      }
      if (!run?.run_id || ev.run_id !== run.run_id) return;
      const line = ev.type ?? JSON.stringify(ev);
      liveEvents = [line, ...liveEvents].slice(0, 80);
      void refreshRun(run.run_id);
    });
  });
  onDestroy(() => stopEvents?.());

  async function loadView(name: View) {
    err = "";
    try {
      if (name === "RUNS") runs = (await listRuns()) ?? [];
      if (name === "BOARD") boardRuns = (await listBoardRuns()) ?? [];
      if (name === "EVENTS") events = (await listRecentEvents()) ?? [];
      if (name === "GRAPH") graph = (await getGraphSnapshot()) ?? { nodes: [], edges: [] };
      if (name === "CONFIG") {
        config = await configSnapshot();
        lessons = (await listLessons(lessonFilter === "all" ? "" : lessonFilter)) ?? [];
      }
    } catch (e) {
      err = errMessage(e);
    }
  }

  async function previewIntent() {
    busy = true;
    err = "";
    try {
      const out = jsonMode ? await compileTaskJSON(draft) : await compileIntent(draft);
      method = out.method;
      preview = JSON.stringify(out.task, null, 2);
    } catch (e) {
      err = errMessage(e);
    } finally {
      busy = false;
    }
  }

  async function submit() {
    busy = true;
    err = "";
    try {
      const out = jsonMode ? await submitTaskJSON(draft) : await submitIntent(draft);
      method = out.method;
      preview = JSON.stringify(out.task, null, 2);
      await refreshRun(out.run_id);
      view = "LIVE";
    } catch (e) {
      err = errMessage(e);
    } finally {
      busy = false;
    }
  }

  async function refreshRun(id: string) {
    run = await getRun(id);
    liveEvents = (run.events ?? []).map((e) => e.type ?? "").reverse();
  }

  async function openRun(id: string) {
    err = "";
    try {
      await refreshRun(id);
      view = "LIVE";
    } catch (e) {
      err = errMessage(e);
    }
  }

  async function onRefreshGraph() {
    busy = true;
    err = "";
    try {
      await refreshGraph();
      graph = (await getGraphSnapshot()) ?? { nodes: [], edges: [] };
    } catch (e) {
      err = errMessage(e);
    } finally {
      busy = false;
    }
  }

  async function onLesson(id: string, action: "approve" | "reject") {
    busy = true;
    err = "";
    try {
      if (action === "approve") await approveLesson(id);
      else await rejectLesson(id);
      lessons = (await listLessons(lessonFilter === "all" ? "" : lessonFilter)) ?? [];
    } catch (e) {
      err = errMessage(e);
    } finally {
      busy = false;
    }
  }

  async function reloadLessons() {
    lessons = (await listLessons(lessonFilter === "all" ? "" : lessonFilter)) ?? [];
  }

  function onKey(e: KeyboardEvent) {
    const mod = e.metaKey || e.ctrlKey;
    if (!mod) return;
    if (e.key.toLowerCase() === "p") {
      e.preventDefault();
      void previewIntent();
    } else if (e.key === "Enter") {
      e.preventDefault();
      void submit();
    } else if (e.key.toLowerCase() === "j") {
      e.preventDefault();
      jsonMode = !jsonMode;
    }
  }

  function selectView(name: View) {
    view = name;
    void loadView(name);
  }

  function eventLine(ev: RuntimeEvent): string {
    const hay = [ev.type, ev.run_id, ev.step_id, ev.payload?.player].filter(Boolean).join(" ").toLowerCase();
    return hay;
  }

  const shownEvents = $derived(
    events.filter((ev) => {
      const q = eventFilter.trim().toLowerCase();
      if (!q) return true;
      return eventLine(ev).includes(q);
    }),
  );

  const shownNodes = $derived(
    (graph.nodes ?? []).filter((n) => {
      const q = graphFilter.trim().toLowerCase();
      if (!q) return true;
      return `${n.kind} ${n.id}`.toLowerCase().includes(q);
    }),
  );

  const shownEdges = $derived(
    (graph.edges ?? []).filter((e) => {
      const q = graphFilter.trim().toLowerCase();
      if (!q) return true;
      return `${e.kind} ${e.from?.id} ${e.to?.id}`.toLowerCase().includes(q);
    }),
  );

  const lanes = $derived.by(() => {
    const out = {
      INTAKE: [] as RunSummary[],
      "IN FLIGHT": [] as RunSummary[],
      LANDED: [] as RunSummary[],
    };
    for (const row of boardRuns) {
      out[boardLane(row.status)].push(row);
    }
    return out;
  });
</script>

<svelte:window onkeydown={onKey} />

<div class="app">
  <header class="header">
    <span class="brand">✦ RUNTGINE / CONSTELLATION MISSION CONTROL</span>
    <span class="workspace">desktop · source wails</span>
  </header>

  <nav class="sidebar" aria-label="views">
    {#each views as name}
      <button class="nav" class:active={view === name} onclick={() => selectView(name)}>{name}</button>
    {/each}
  </nav>

  <main class="main">
    {#if view === "INTENT"}
      <section class="card">
        <h1>INTENT</h1>
        <p class="muted">Mission Brief. Compiles NL or Task IR — not a chatbot.</p>
        <label class="muted">
          <input type="checkbox" bind:checked={jsonMode} /> JSON Task IR
        </label>
        <textarea class="draft" bind:value={draft} placeholder="git status"></textarea>
        <div class="row">
          <button class="btn" onclick={previewIntent} disabled={busy}>Preview</button>
          <button class="btn primary" onclick={submit} disabled={busy}>Submit</button>
          {#if method}<span class="badge">{method}</span>{/if}
        </div>
        {#if err}<p class="err">{err}</p>{/if}
        {#if preview}
          <h2>Task IR</h2>
          <pre class="preview">{preview}</pre>
        {/if}
      </section>
    {:else if view === "RUNS"}
      <section class="card">
        <h1>RUNS</h1>
        <p class="muted">Select a run to follow it on LIVE.</p>
        <div class="row">
          <button class="btn" onclick={() => loadView("RUNS")} disabled={busy}>Refresh</button>
        </div>
        {#if err}<p class="err">{err}</p>{/if}
        {#if runs.length === 0}
          <p class="muted">No runs recorded.</p>
        {:else}
          <table>
            <thead>
              <tr><th>Status</th><th>Run</th><th>Source</th><th>Intent</th></tr>
            </thead>
            <tbody>
              {#each runs as row}
                <tr class="click" onclick={() => openRun(row.run_id)}>
                  <td><span class="badge">{row.status}</span></td>
                  <td><code>{shortID(row.run_id)}</code></td>
                  <td>{row.source ?? "-"}</td>
                  <td>{row.summary ?? ""}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      </section>
    {:else if view === "LIVE"}
      <section class="card">
        <h1>LIVE</h1>
        {#if !run}
          <p class="muted">Submit from INTENT or select a run in RUNS.</p>
        {:else}
          <p>
            <span class="badge">{run.status}</span>
            <code>{run.run_id}</code>
          </p>
          {#if run.error}<p class="err">{run.error}</p>{/if}
          <div class="row">
            <button class="btn warn" onclick={() => run && cancelRun(run.run_id)}>Cancel</button>
            {#if run.pending_approval}
              <button class="btn primary" onclick={() => run && approveRun(run.run_id)}>Approve</button>
              <button class="btn danger" onclick={() => run && denyRun(run.run_id)}>Deny</button>
            {/if}
          </div>
          <h2>Events</h2>
          <pre class="events">{liveEvents.join("\n") || "(waiting)"}</pre>
        {/if}
      </section>
    {:else if view === "BOARD"}
      <section class="card">
        <h1>BOARD</h1>
        <p class="muted">Display-only of board-origin runs. Poll with <code>runtgine board poll</code>.</p>
        <div class="row">
          <button class="btn" onclick={() => loadView("BOARD")} disabled={busy}>Refresh</button>
        </div>
        {#if err}<p class="err">{err}</p>{/if}
        <div class="lanes">
          {#each laneNames as name}
            <div class="lane">
              <h2>{name} · {lanes[name].length}</h2>
              {#if lanes[name].length === 0}
                <p class="muted">No board runs.</p>
              {:else}
                {#each lanes[name] as row}
                  <button class="cardlet" onclick={() => openRun(row.run_id)}>
                    <span class="badge">{row.status}</span>
                    <strong>{row.source_ref || shortID(row.run_id)}</strong>
                    <span class="muted">{row.summary ?? ""}</span>
                  </button>
                {/each}
              {/if}
            </div>
          {/each}
        </div>
      </section>
    {:else if view === "EVENTS"}
      <section class="card">
        <h1>EVENTS</h1>
        <p class="muted">Recent Core events plus the live <code>runtgine:event</code> stream.</p>
        <div class="row">
          <input class="filter" bind:value={eventFilter} placeholder="filter type / run / step" />
          <button class="btn" onclick={() => loadView("EVENTS")} disabled={busy}>Refresh</button>
        </div>
        {#if err}<p class="err">{err}</p>{/if}
        {#if shownEvents.length === 0}
          <p class="muted">No events.</p>
        {:else}
          <table>
            <thead>
              <tr><th>Type</th><th>Run</th><th>Step</th></tr>
            </thead>
            <tbody>
              {#each shownEvents.slice(0, 120) as ev}
                <tr>
                  <td>{ev.type ?? "-"}</td>
                  <td><code>{shortID(ev.run_id)}</code></td>
                  <td>{ev.step_id ?? "-"}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      </section>
    {:else if view === "GRAPH"}
      <section class="card">
        <h1>GRAPH</h1>
        <p class="muted">Read-only Runtime Graph snapshot.</p>
        <div class="row">
          <input class="filter" bind:value={graphFilter} placeholder="filter kind / id" />
          <button class="btn" onclick={onRefreshGraph} disabled={busy}>Refresh graph</button>
        </div>
        {#if err}<p class="err">{err}</p>{/if}
        <h2>Nodes · {shownNodes.length}</h2>
        {#if shownNodes.length === 0}
          <p class="muted">No nodes.</p>
        {:else}
          <table>
            <thead><tr><th>Kind</th><th>ID</th></tr></thead>
            <tbody>
              {#each shownNodes.slice(0, 80) as n}
                <tr><td>{n.kind}</td><td><code>{n.id}</code></td></tr>
              {/each}
            </tbody>
          </table>
        {/if}
        <h2>Edges · {shownEdges.length}</h2>
        {#if shownEdges.length === 0}
          <p class="muted">No edges.</p>
        {:else}
          <table>
            <thead><tr><th>Kind</th><th>From</th><th>To</th></tr></thead>
            <tbody>
              {#each shownEdges.slice(0, 80) as e}
                <tr>
                  <td>{e.kind}</td>
                  <td><code>{e.from?.kind}:{e.from?.id}</code></td>
                  <td><code>{e.to?.kind}:{e.to?.id}</code></td>
                </tr>
              {/each}
            </tbody>
          </table>
        {/if}
      </section>
    {:else if view === "CONFIG"}
      <section class="card">
        <h1>CONFIG</h1>
        <p class="muted">Read-only snapshot. Tokens and API keys are never shown.</p>
        {#if err}<p class="err">{err}</p>{/if}
        {#if config}
          <dl class="kv">
            <dt>workspace</dt><dd><code>{config.workspace_root}</code></dd>
            <dt>SQLite</dt><dd><code>{config.db_path}</code></dd>
            <dt>log level</dt><dd>{config.log_level}</dd>
            <dt>max concurrent runs</dt><dd>{config.max_concurrent_runs}</dd>
            <dt>LLM backend</dt><dd>{config.llm_backend || "none"}</dd>
            <dt>LLM credentials</dt><dd>{config.llm_connected ? "connected" : "not configured"}</dd>
            <dt>GitHub credentials</dt><dd>{config.github_connected ? "connected" : "not configured"}</dd>
            <dt>precedence</dt><dd>{config.precedence}</dd>
          </dl>
        {/if}
        <h2>Lessons HITL</h2>
        <div class="row">
          <select class="filter" bind:value={lessonFilter} onchange={() => reloadLessons()}>
            <option value="pending">pending</option>
            <option value="approved">approved</option>
            <option value="rejected">rejected</option>
            <option value="all">all</option>
          </select>
          <button class="btn" onclick={() => loadView("CONFIG")} disabled={busy}>Refresh</button>
        </div>
        {#if lessons.length === 0}
          <p class="muted">No lessons.</p>
        {:else}
          {#each lessons as les}
            <article class="lesson">
              <div class="row">
                <span class="badge">{les.status}</span>
                <strong>{les.title}</strong>
              </div>
              <p class="muted">{les.body ?? ""}</p>
              {#if les.status === "pending"}
                <div class="row">
                  <button class="btn primary" disabled={busy} onclick={() => onLesson(les.id, "approve")}>Approve</button>
                  <button class="btn danger" disabled={busy} onclick={() => onLesson(les.id, "reject")}>Reject</button>
                </div>
              {/if}
            </article>
          {/each}
        {/if}
      </section>
    {/if}
  </main>

  <footer class="footer">
    Ctrl/Cmd+P preview · Ctrl/Cmd+Enter submit · Ctrl/Cmd+J JSON · one window
  </footer>
</div>
