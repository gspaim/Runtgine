<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import {
    compileIntent,
    compileTaskJSON,
    submitIntent,
    submitTaskJSON,
    getRun,
    cancelRun,
    approveRun,
    denyRun,
    onRuntimeEvent,
    errMessage,
    type RunSnapshot,
  } from "./lib/core";

  const views = ["INTENT", "RUNS", "LIVE", "BOARD", "EVENTS", "GRAPH", "CONFIG"] as const;
  type View = (typeof views)[number];

  let view = $state<View>("INTENT");
  let draft = $state("git status");
  let jsonMode = $state(false);
  let preview = $state("");
  let method = $state("");
  let err = $state("");
  let busy = $state(false);
  let run = $state<RunSnapshot | null>(null);
  let liveEvents = $state<string[]>([]);
  let stopEvents: (() => void) | undefined;

  onMount(() => {
    stopEvents = onRuntimeEvent((ev) => {
      if (!run?.run_id || ev.run_id !== run.run_id) return;
      const line = ev.type ?? JSON.stringify(ev);
      liveEvents = [line, ...liveEvents].slice(0, 80);
      void refreshRun(run.run_id);
    });
  });
  onDestroy(() => stopEvents?.());

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
    liveEvents = (run.events ?? []).map((e) => e.type).reverse();
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
</script>

<svelte:window onkeydown={onKey} />

<div class="app">
  <header class="header">
    <span class="brand">✦ RUNTGINE / CONSTELLATION MISSION CONTROL</span>
    <span class="workspace">desktop · source wails</span>
  </header>

  <nav class="sidebar" aria-label="views">
    {#each views as name}
      <button class="nav" class:active={view === name} onclick={() => (view = name)}>{name}</button>
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
    {:else if view === "LIVE"}
      <section class="card">
        <h1>LIVE</h1>
        {#if !run}
          <p class="muted">Submit from INTENT to follow a run.</p>
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
    {:else}
      <section class="card">
        <h1>{view}</h1>
        <p class="muted">Slice 28 — remaining Constellation views.</p>
      </section>
    {/if}
  </main>

  <footer class="footer">
    Ctrl/Cmd+P preview · Ctrl/Cmd+Enter submit · Ctrl/Cmd+J JSON · one window
  </footer>
</div>
