import { lazy, Suspense, useState, type ReactNode } from "react";
import { CheckCircle2, Loader2, Play, RotateCcw, XCircle, AlertTriangle } from "lucide-react";
import { Link } from "react-router-dom";
import { useConfig } from "@/api/queries";
import { ApiError } from "@/api/client";
import type { RunResult } from "@/api/types";
import { useAuth } from "@/lib/auth";
import { Badge, Button, Kbd, Skeleton } from "@/components/ui/ui";
import { modKey } from "@/lib/utils";

const CodeEditor = lazy(() => import("./CodeEditor"));

export interface Outcome {
  run?: RunResult;
  /** Optional verdict shown above the console (e.g. exercise passed/failed). */
  verdict?: { ok: boolean; text: string };
  tests?: { name: string; passed: boolean; message?: string }[];
  buildOutput?: string;
  expected?: string;
}

interface Props {
  label: string;
  initial: string;
  /** Called with the current code; should reject with ApiError on failure. */
  onRun: (code: string) => Promise<Outcome>;
  runLabel?: string;
  expected?: string;
  minHeight?: string;
  maxHeight?: string;
  extraActions?: ReactNode;
  onCodeChange?: (code: string) => void;
  resetTo?: string;
}

export function RunPanel({ label, initial, onRun, runLabel = "Run", expected, minHeight, maxHeight, extraActions, onCodeChange, resetTo }: Props) {
  const { user } = useAuth();
  const cfg = useConfig();
  const [code, setCode] = useState(initial);
  const [busy, setBusy] = useState(false);
  const [outcome, setOutcome] = useState<Outcome | null>(null);
  const [error, setError] = useState<ApiError | Error | null>(null);
  const runner = cfg.data?.runner;
  const disabledReason = !user ? "Sign in to run code." : runner && !runner.available ? (runner.reason ?? "Code execution is unavailable.") : null;

  const run = async () => {
    if (busy || disabledReason) return;
    setBusy(true);
    setError(null);
    try {
      setOutcome(await onRun(code));
    } catch (e) {
      setOutcome(null);
      setError(e instanceof Error ? e : new Error("Run failed."));
    } finally {
      setBusy(false);
    }
  };

  const reset = () => {
    const to = resetTo ?? initial;
    setCode(to);
    onCodeChange?.(to);
    setOutcome(null);
    setError(null);
  };

  return (
    <div className="overflow-hidden rounded-xl border border-border bg-surface shadow-card">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border px-3 py-2">
        <span className="flex items-center gap-2 font-mono text-xs text-muted">
          <span className="h-2 w-2 rounded-full bg-accent" aria-hidden /> main.go
        </span>
        <div className="flex items-center gap-2">
          {extraActions}
          <Button size="sm" variant="ghost" onClick={reset} aria-label="Reset code to starter">
            <RotateCcw className="h-3.5 w-3.5" aria-hidden /> Reset
          </Button>
          <Button size="sm" variant="primary" onClick={run} aria-disabled={busy || !!disabledReason} aria-describedby={disabledReason ? "run-disabled-reason" : undefined}>
            {busy ? <Loader2 className="h-3.5 w-3.5 animate-spin" aria-hidden /> : <Play className="h-3.5 w-3.5" aria-hidden />}
            {busy ? "Running…" : runLabel}
          </Button>
        </div>
      </div>

      {disabledReason && (
        <div id="run-disabled-reason" className="flex items-start gap-2 border-b border-border bg-warn-soft px-3 py-2 text-sm text-warn">
          <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" aria-hidden />
          <span>
            {!user ? (
              <>
                <Link className="underline" to="/login">
                  Sign in
                </Link>{" "}
                to run code.
              </>
            ) : (
              <>Code execution is unavailable: {disabledReason}</>
            )}
          </span>
        </div>
      )}

      <div className="p-2">
        <Suspense fallback={<Skeleton className="h-40 w-full" />}>
          <CodeEditor
            label={label}
            value={code}
            onChange={(v) => {
              setCode(v);
              onCodeChange?.(v);
            }}
            onRun={run}
            minHeight={minHeight}
            maxHeight={maxHeight}
          />
        </Suspense>
        <p id="editor-hint" className="mt-1.5 px-1 text-xs text-faint">
          <Kbd>{modKey()}</Kbd> + <Kbd>Enter</Kbd> to run · <Kbd>Esc</Kbd> then <Kbd>Tab</Kbd> to leave the editor
        </p>
      </div>

      <Console busy={busy} error={error} outcome={outcome} expected={expected} />
    </div>
  );
}

function Console({ busy, error, outcome, expected }: { busy: boolean; error: Error | null; outcome: Outcome | null; expected?: string }) {
  return (
    <div className="border-t border-border bg-bg-subtle px-3 py-3" aria-live="polite">
      <div className="mb-2 flex items-center justify-between">
        <h3 className="text-xs font-semibold uppercase tracking-wide text-muted">Output</h3>
        {outcome?.run && (
          <span className="font-mono text-xs text-faint">
            exit {outcome.run.exitCode} · {outcome.run.durationMs} ms
          </span>
        )}
      </div>
      {busy && <p className="text-sm text-muted">Compiling and running in an isolated sandbox…</p>}
      {!busy && error && (
        <div role="alert" className="rounded-lg border border-danger/40 bg-danger-soft p-3 text-sm text-danger">
          {error.message}
        </div>
      )}
      {!busy && !error && !outcome && <p className="text-sm text-faint">Run your code to see output here.</p>}
      {!busy && outcome && (
        <div className="space-y-3">
          {outcome.verdict && (
            <p className={`flex items-center gap-2 text-sm font-medium ${outcome.verdict.ok ? "text-success" : "text-danger"}`}>
              {outcome.verdict.ok ? <CheckCircle2 className="h-4 w-4" aria-hidden /> : <XCircle className="h-4 w-4" aria-hidden />}
              {outcome.verdict.text}
            </p>
          )}
          {outcome.run?.timedOut && <Badge tone="danger">Timed out — check for infinite loops</Badge>}
          {outcome.run?.truncated && <Badge tone="warn">Output truncated</Badge>}
          {outcome.run && outcome.run.stdout && <Pre label="stdout">{outcome.run.stdout}</Pre>}
          {outcome.run && outcome.run.stderr && <Pre label={outcome.run.buildError ? "build errors" : "stderr"} danger>{outcome.run.stderr}</Pre>}
          {outcome.buildOutput && <Pre label="build errors" danger>{outcome.buildOutput}</Pre>}
          {outcome.run && !outcome.run.stdout && !outcome.run.stderr && !outcome.buildOutput && !outcome.tests && (
            <p className="text-sm text-faint">(no output)</p>
          )}
          {outcome.tests && (
            <ul className="space-y-1.5" aria-label="Test results">
              {outcome.tests.map((t) => (
                <li key={t.name} className="text-sm">
                  <span className={`flex items-center gap-2 font-mono ${t.passed ? "text-success" : "text-danger"}`}>
                    {t.passed ? <CheckCircle2 className="h-4 w-4" aria-hidden /> : <XCircle className="h-4 w-4" aria-hidden />}
                    {t.name}
                  </span>
                  {!t.passed && t.message && <Pre label="failure" danger>{t.message}</Pre>}
                </li>
              ))}
            </ul>
          )}
          {expected && outcome.verdict && !outcome.verdict.ok && <Pre label="expected output">{expected}</Pre>}
        </div>
      )}
    </div>
  );
}

function Pre({ children, label, danger }: { children: string; label: string; danger?: boolean }) {
  return (
    <div>
      <p className="mb-1 text-[0.7rem] uppercase tracking-wide text-faint">{label}</p>
      <pre
        tabIndex={0}
        className={`max-h-64 overflow-auto whitespace-pre-wrap break-words rounded-lg border p-2.5 font-mono text-xs leading-relaxed ${
          danger ? "border-danger/30 bg-danger-soft text-danger" : "border-border bg-code-bg text-fg"
        }`}
      >
        {children}
      </pre>
    </div>
  );
}
