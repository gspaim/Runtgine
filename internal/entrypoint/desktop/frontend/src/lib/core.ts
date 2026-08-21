import { Call, Events } from "@wailsio/runtime";

const svc = (method: string, ...args: unknown[]) =>
  Call.ByName(
    `github.com/gspaim/Runtgine/internal/entrypoint/desktop.Service.${method}`,
    ...args,
  );

export type IntentPreview = { task: unknown; method: string };
export type IntentSubmit = { run_id: string; method: string; task: unknown };
export type RunSnapshot = {
  run_id: string;
  task_id?: string;
  status: string;
  error?: string;
  events?: Array<{ type: string; ts?: string; run_id?: string; payload?: unknown }>;
  pending_approval?: { step_id: string; capability: string; player: string } | null;
};

export const compileIntent = (text: string) => svc("CompileIntent", text) as Promise<IntentPreview>;
export const submitIntent = (text: string) => svc("SubmitIntent", text) as Promise<IntentSubmit>;
export const compileTaskJSON = (raw: string) => svc("CompileTaskJSON", raw) as Promise<IntentPreview>;
export const submitTaskJSON = (raw: string) => svc("SubmitTaskJSON", raw) as Promise<IntentSubmit>;
export const getRun = (id: string) => svc("GetRun", id) as Promise<RunSnapshot>;
export const cancelRun = (id: string) => svc("CancelRun", id) as Promise<void>;
export const approveRun = (id: string) => svc("ApproveRun", id) as Promise<void>;
export const denyRun = (id: string) => svc("DenyRun", id) as Promise<void>;

export function onRuntimeEvent(cb: (ev: { type?: string; run_id?: string; data?: unknown }) => void) {
  return Events.On("runtgine:event", (ev: { data?: unknown }) => {
    const data = ev?.data as { type?: string; run_id?: string };
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
