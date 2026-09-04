import { Call, Events } from "@wailsio/runtime";

const svc = (method: string, ...args: unknown[]) =>
  Call.ByName(
    `github.com/gspaim/Runtgine/internal/entrypoint/desktop.Service.${method}`,
    ...args,
  );

export type IntentPreview = { task: unknown; method: string };
export type IntentSubmit = { run_id: string; method: string; task: unknown };
export type RuntimeEvent = {
  type?: string;
  ts?: string;
  run_id?: string;
  task_id?: string;
  step_id?: string | null;
  payload?: Record<string, unknown>;
};
export type RunSnapshot = {
  run_id: string;
  task_id?: string;
  status: string;
  error?: string;
  events?: RuntimeEvent[];
  pending_approval?: { step_id: string; capability: string; player: string } | null;
};
export type RunSummary = {
  run_id: string;
  task_id?: string;
  status: string;
  summary?: string;
  source?: string;
  source_ref?: string;
  created_at?: string;
  updated_at?: string;
};
export type ConfigSnapshot = {
  workspace_root: string;
  db_path: string;
  log_level: string;
  max_concurrent_runs: number;
  llm_backend: string;
  llm_connected: boolean;
  github_connected: boolean;
  precedence: string;
};
export type GraphNode = { kind: string; id: string; attrs?: Record<string, unknown> };
export type GraphEdge = { kind: string; from: { kind: string; id: string }; to: { kind: string; id: string } };
export type GraphSnapshot = { nodes?: GraphNode[]; edges?: GraphEdge[] };
export type Lesson = {
  id: string;
  run_id?: string;
  task_id?: string;
  kind?: string;
  title: string;
  body?: string;
  status: string;
  created_at?: string;
};

export const compileIntent = (text: string) => svc("CompileIntent", text) as Promise<IntentPreview>;
export const submitIntent = (text: string) => svc("SubmitIntent", text) as Promise<IntentSubmit>;
export const compileTaskJSON = (raw: string) => svc("CompileTaskJSON", raw) as Promise<IntentPreview>;
export const submitTaskJSON = (raw: string) => svc("SubmitTaskJSON", raw) as Promise<IntentSubmit>;
export const getRun = (id: string) => svc("GetRun", id) as Promise<RunSnapshot>;
export const listRuns = (limit = 200) => svc("ListRuns", limit) as Promise<RunSummary[]>;
export const listBoardRuns = (limit = 200) => svc("ListBoardRuns", limit) as Promise<RunSummary[]>;
export const listRecentEvents = (limit = 300) => svc("ListRecentEvents", limit) as Promise<RuntimeEvent[]>;
export const configSnapshot = () => svc("ConfigSnapshot") as Promise<ConfigSnapshot>;
export const getGraphSnapshot = () => svc("GetGraphSnapshot") as Promise<GraphSnapshot>;
export const refreshGraph = () => svc("RefreshGraph") as Promise<void>;
export const listLessons = (status = "") => svc("ListLessons", status) as Promise<Lesson[]>;
export const approveLesson = (id: string) => svc("ApproveLesson", id) as Promise<void>;
export const rejectLesson = (id: string) => svc("RejectLesson", id) as Promise<void>;
export const cancelRun = (id: string) => svc("CancelRun", id) as Promise<void>;
export const approveRun = (id: string) => svc("ApproveRun", id) as Promise<void>;
export const denyRun = (id: string) => svc("DenyRun", id) as Promise<void>;

export function onRuntimeEvent(cb: (ev: RuntimeEvent) => void) {
  return Events.On("runtgine:event", (ev: { data?: unknown }) => {
    const data = ev?.data as RuntimeEvent;
    cb(data ?? {});
  });
}

export function errMessage(err: unknown): string {
  if (!err) return "unknown error";
  if (typeof err === "string") return err;
  const e = err as { message?: string; cause?: { code?: string; message?: string } };
  if (e.cause?.code) return `${e.cause.code}: ${e.cause.message ?? e.message ?? ""}`;
  return e.message ?? String(err);
}

export function shortID(id: string | undefined): string {
  if (!id) return "-";
  return id.length > 12 ? id.slice(0, 8) : id;
}

export function boardLane(status: string): "INTAKE" | "IN FLIGHT" | "LANDED" {
  switch (status) {
    case "accepted":
    case "planned":
      return "INTAKE";
    case "running":
    case "waiting_approval":
      return "IN FLIGHT";
    default:
      return "LANDED";
  }
}
