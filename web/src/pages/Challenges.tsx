import { useState } from "react";
import { Link, useParams, useSearchParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { CheckCircle2, Circle, Lightbulb, Lock } from "lucide-react";
import { useChallenge, useChallenges, useModules } from "@/api/queries";
import { get, send } from "@/api/client";
import type { ChallengeResult } from "@/api/types";
import { Async } from "@/components/ui/async";
import { Badge, Button, Card, CardHeader, PageHeader } from "@/components/ui/ui";
import { CodeBlock, Markdown } from "@/components/Markdown";
import { RunPanel, type Outcome } from "@/components/RunPanel";
import { DifficultyBadge } from "@/components/common";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/lib/auth";
import { useDocumentTitle, useDraft, useLearningTimer } from "@/lib/hooks";
import { cn, difficultyLabel } from "@/lib/utils";

export function ChallengeIndex() {
  const { user } = useAuth();
  const [sp, setSp] = useSearchParams();
  const filters = {
    difficulty: sp.get("difficulty") ?? undefined,
    module: sp.get("module") ?? undefined,
    status: sp.get("status") ?? undefined,
    q: sp.get("q") ?? undefined,
    page: Number(sp.get("page") ?? 1),
  };
  const q = useChallenges(filters);
  const mods = useModules(!!user);
  const set = (k: string, v: string) => {
    const n = new URLSearchParams(sp);
    if (v) n.set(k, v);
    else n.delete(k);
    n.delete("page");
    setSp(n, { replace: true });
  };
  const sel = "h-9 rounded-lg border border-border bg-surface px-2 text-sm";
  return (
    <>
      <PageHeader title="Challenges" description="Solve problems against real Go tests. Each challenge has hints, and the solution unlocks after three attempts or a pass." />
      <div className="mb-5 flex flex-wrap items-center gap-2">
        <label className="sr-only" htmlFor="c-q">Search challenges</label>
        <input id="c-q" defaultValue={filters.q ?? ""} placeholder="Filter by title…" onChange={(e) => set("q", e.target.value)} className="h-9 w-48 rounded-lg border border-border bg-surface px-3 text-sm outline-none focus:border-accent" />
        <label className="sr-only" htmlFor="c-d">Difficulty</label>
        <select id="c-d" className={sel} value={filters.difficulty ?? ""} onChange={(e) => set("difficulty", e.target.value)}>
          <option value="">All difficulties</option>
          {Object.entries(difficultyLabel).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
        </select>
        <label className="sr-only" htmlFor="c-m">Module</label>
        <select id="c-m" className={sel} value={filters.module ?? ""} onChange={(e) => set("module", e.target.value)}>
          <option value="">All modules</option>
          {mods.data?.map((m) => <option key={m.id} value={m.id}>{m.number} {m.title}</option>)}
        </select>
        {user && (
          <>
            <label className="sr-only" htmlFor="c-s">Status</label>
            <select id="c-s" className={sel} value={filters.status ?? ""} onChange={(e) => set("status", e.target.value)}>
              <option value="">Any status</option>
              <option value="todo">To do</option>
              <option value="passed">Passed</option>
            </select>
          </>
        )}
      </div>
      <Async query={q} isEmpty={(d) => d.data.length === 0} empty={<p className="rounded-xl border border-dashed border-border p-8 text-center text-sm text-muted">No challenges match these filters.</p>}>
        {(res) => (
          <>
            <ul className="grid gap-2">
              {res.data.map((c) => (
                <li key={c.id}>
                  <Link to={`/challenge/${c.id}`} className="flex items-center gap-3 rounded-xl border border-border bg-surface px-4 py-3 hover:border-accent">
                    {c.passed ? <CheckCircle2 className="h-4 w-4 shrink-0 text-success" aria-label="Passed" /> : <Circle className="h-4 w-4 shrink-0 text-faint" aria-label="Not passed" />}
                    <span className="min-w-0 flex-1">
                      <span className="block truncate font-medium">{c.title}</span>
                      <span className="block truncate text-xs text-muted">{mods.data?.find((m) => m.id === c.moduleId)?.title ?? c.moduleId}</span>
                    </span>
                    {c.attempts > 0 && !c.passed && <span className="hidden text-xs text-faint sm:inline">{c.attempts} attempts</span>}
                    <DifficultyBadge level={c.difficulty} />
                  </Link>
                </li>
              ))}
            </ul>
            <Pager page={res.meta.page} size={res.meta.pageSize} total={res.meta.total} onPage={(p) => { const n = new URLSearchParams(sp); n.set("page", String(p)); setSp(n); }} />
          </>
        )}
      </Async>
    </>
  );
}

function Pager({ page, size, total, onPage }: { page: number; size: number; total: number; onPage: (p: number) => void }) {
  const pages = Math.max(1, Math.ceil(total / size));
  if (pages <= 1) return <p className="mt-4 text-xs text-faint">{total} challenges</p>;
  return (
    <nav aria-label="Pagination" className="mt-4 flex items-center justify-between text-sm">
      <span className="text-muted">{total} challenges · page {page} of {pages}</span>
      <div className="flex gap-2">
        <Button size="sm" disabled={page <= 1} onClick={() => onPage(page - 1)}>Previous</Button>
        <Button size="sm" disabled={page >= pages} onClick={() => onPage(page + 1)}>Next</Button>
      </div>
    </nav>
  );
}

export function ChallengePage() {
  const { id = "" } = useParams();
  const q = useChallenge(id);
  useDocumentTitle(q.data?.challenge.title);
  useLearningTimer();
  return <Async query={q}>{(d) => <ChallengeView key={d.challenge.id} d={d} />}</Async>;
}

function ChallengeView({ d }: { d: NonNullable<ReturnType<typeof useChallenge>["data"]> }) {
  const { user } = useAuth();
  const c = d.challenge;
  const qc = useQueryClient();
  const toast = useToast();
  const [draft, setDraft] = useDraft(`golearn.challenge.${c.id}`);
  const [hints, setHints] = useState(0);
  const [solution, setSolution] = useState<{ solution: string; explanation: string } | null>(null);
  const [solErr, setSolErr] = useState<string | null>(null);
  const initial = draft ?? (d.lastCode || c.starter);

  const loadSolution = async () => {
    setSolErr(null);
    try {
      setSolution(await get<{ solution: string; explanation: string }>(`/api/challenges/${c.id}/solution`));
    } catch (e) {
      setSolErr(e instanceof Error ? e.message : "Couldn't load the solution.");
    }
  };
  const unlocked = d.solutionUnlocked || d.status.attempts >= d.solutionAfter;
  const hintList = c.hints ?? [];

  return (
    <div className="grid gap-6 lg:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
      <div className="space-y-5">
        <nav aria-label="Breadcrumb" className="text-sm text-muted"><Link to="/challenge" className="hover:text-fg">Challenges</Link> / <span className="text-fg">{c.title}</span></nav>
        <header>
          <h1 className="text-2xl font-semibold tracking-tight">{c.title}</h1>
          <div className="mt-2 flex flex-wrap items-center gap-2 text-sm">
            <DifficultyBadge level={c.difficulty} />
            {d.status.passed && <Badge tone="success"><CheckCircle2 className="h-3 w-3" aria-hidden /> Passed</Badge>}
            {d.status.attempts > 0 && <span className="text-muted">{d.status.attempts} attempt{d.status.attempts === 1 ? "" : "s"}</span>}
            {d.lesson && <Link to={`/learn/${d.lesson.module}/${d.lesson.slug}`} className="text-accent underline">Related lesson: {d.lesson.title}</Link>}
          </div>
        </header>
        <Section title="Problem"><Markdown>{c.problem}</Markdown></Section>
        {c.input && <Section title="Input"><Markdown>{c.input}</Markdown></Section>}
        {c.expectedOutput && <Section title="Expected output"><Markdown>{c.expectedOutput}</Markdown></Section>}
        {c.constraints && c.constraints.length > 0 && (
          <Section title="Constraints"><Markdown>{c.constraints.map((x) => `- ${x}`).join("\n")}</Markdown></Section>
        )}
        <Section title={`Test cases (${c.testNames.length})`}>
          <ul className="space-y-1 font-mono text-xs text-muted">{c.testNames.map((t) => <li key={t}>{t}</li>)}</ul>
          <p className="mt-2 text-xs text-faint">Tests use Go's <code className="font-mono">testing</code> package. Run <code className="font-mono">go test ./...</code> locally to reproduce.</p>
        </Section>
        <Card>
          <CardHeader title="Hints" action={<span className="text-xs text-faint">{Math.min(hints, hintList.length)}/{hintList.length}</span>} />
          <div className="space-y-2 p-4 text-sm">
            {hintList.slice(0, hints).map((h, i) => (
              <p key={i} className="flex gap-2"><Lightbulb className="mt-0.5 h-4 w-4 shrink-0 text-warn" aria-hidden />{h}</p>
            ))}
            {hints < hintList.length ? (
              <Button size="sm" onClick={() => setHints((h) => h + 1)}>Reveal {hints === 0 ? "a" : "another"} hint</Button>
            ) : hintList.length === 0 ? <p className="text-muted">No hints for this one.</p> : null}
          </div>
        </Card>
        <Card>
          <CardHeader title="Solution" action={!unlocked && <Lock className="h-4 w-4 text-faint" aria-label="Locked" />} />
          <div className="space-y-3 p-4 text-sm">
            {solution ? (
              <>
                <CodeBlock code={solution.solution} lang="go" />
                <Markdown>{solution.explanation}</Markdown>
              </>
            ) : unlocked ? (
              <Button size="sm" onClick={loadSolution}>Show solution and explanation</Button>
            ) : (
              <p className="text-muted">Unlocks after {d.solutionAfter} attempts or when you pass all tests. {user ? `${d.status.attempts}/${d.solutionAfter} attempts so far.` : "Sign in to attempt."}</p>
            )}
            {solErr && <p role="alert" className="text-danger">{solErr}</p>}
          </div>
        </Card>
      </div>

      <div className="min-w-0 lg:sticky lg:top-20 lg:self-start">
        <RunPanel
          label={`Solution for ${c.title}`}
          initial={initial}
          resetTo={c.starter}
          minHeight="18rem"
          maxHeight="32rem"
          runLabel="Submit & run tests"
          onCodeChange={(v) => setDraft(v === c.starter ? null : v)}
          onRun={async (code): Promise<Outcome> => {
            const r = await send<ChallengeResult>("POST", `/api/challenges/${c.id}/submit`, { code });
            toast.achievement(r.achievements.map((a) => a.title));
            void qc.invalidateQueries({ predicate: (query) => query.queryKey[0] !== "me" });
            return {
              tests: r.tests,
              buildOutput: r.buildOutput,
              run: { stdout: "", stderr: "", exitCode: r.passed ? 0 : 1, timedOut: r.timedOut, truncated: r.truncated, buildError: !!r.buildOutput, durationMs: r.durationMs },
              verdict: { ok: r.passed, text: r.passed ? `All ${r.testsTotal} tests passed.` : `${r.testsPassed}/${r.testsTotal} tests passed.` },
            };
          }}
        />
      </div>
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section>
      <h2 className={cn("mb-2 text-xs font-semibold uppercase tracking-wider text-muted")}>{title}</h2>
      {children}
    </section>
  );
}
