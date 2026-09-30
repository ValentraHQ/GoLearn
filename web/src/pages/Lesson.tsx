import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, ArrowRight, CheckCircle2, ChevronsLeft, Circle, Clock, ListTree, PanelRight } from "lucide-react";
import { useConfig, useLesson, useModules } from "@/api/queries";
import { send } from "@/api/client";
import type { ExerciseResult, LessonDetail, QuizQuestion, QuizResult } from "@/api/types";
import { Async } from "@/components/ui/async";
import { Badge, Button, Card, CardHeader, Kbd, LinkButton, ProgressBar } from "@/components/ui/ui";
import { Sheet } from "@/components/ui/sheet";
import { CodeBlock, Markdown } from "@/components/Markdown";
import { RunPanel, type Outcome } from "@/components/RunPanel";
import { DifficultyBadge } from "@/components/common";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/lib/auth";
import { useDocumentTitle, useDraft, useLearningTimer } from "@/lib/hooks";
import { cn, isTypingTarget, modKey } from "@/lib/utils";

export function LessonPage() {
  const { module = "", slug = "" } = useParams();
  const q = useLesson(module, slug);
  useDocumentTitle(q.data?.lesson.title);
  useLearningTimer();
  return (
    <Async query={q}>{(d) => <LessonView key={d.lesson.id} d={d} />}</Async>
  );
}

function LessonView({ d }: { d: LessonDetail }) {
  const { user } = useAuth();
  const nav = useNavigate();
  const [curriculumOpen, setCurriculumOpen] = useState(false);
  const [progressOpen, setProgressOpen] = useState(false);

  // Ctrl/Cmd + ← / → move between lessons (ignored while typing).
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (!(e.metaKey || e.ctrlKey) || isTypingTarget(e.target)) return;
      if (e.key === "ArrowRight" && d.next) {
        e.preventDefault();
        nav(`/learn/${d.next.module}/${d.next.slug}`);
      } else if (e.key === "ArrowLeft" && d.prev) {
        e.preventDefault();
        nav(`/learn/${d.prev.module}/${d.prev.slug}`);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [d.next, d.prev, nav]);

  return (
    <div className="xl:grid xl:grid-cols-[15rem_minmax(0,1fr)_16rem] xl:gap-6 lg:max-xl:grid lg:max-xl:grid-cols-[14rem_minmax(0,1fr)] lg:max-xl:gap-6">
      <aside className="hidden lg:block" aria-label="Curriculum">
        <div className="sticky top-20 max-h-[calc(100vh-6rem)] overflow-y-auto">
          <CurriculumNav d={d} />
        </div>
      </aside>

      <article className="min-w-0" aria-labelledby="lesson-title">
        <div className="mb-4 flex items-center gap-2 lg:hidden">
          <Button size="sm" onClick={() => setCurriculumOpen(true)}>
            <ListTree className="h-4 w-4" aria-hidden /> Curriculum
          </Button>
          <Button size="sm" onClick={() => setProgressOpen(true)}>
            <PanelRight className="h-4 w-4" aria-hidden /> Progress
          </Button>
        </div>
        <div className="mb-4 hidden items-center gap-2 lg:max-xl:flex xl:hidden">
          <Button size="sm" onClick={() => setProgressOpen(true)}>
            <PanelRight className="h-4 w-4" aria-hidden /> Progress
          </Button>
        </div>

        <nav aria-label="Breadcrumb" className="mb-2 text-sm text-muted">
          <Link to="/learn" className="hover:text-fg">Learn</Link> /{" "}
          <Link to={`/learn/${d.module.id}`} className="hover:text-fg">
            Module {d.module.number}
          </Link>
        </nav>
        <header className="mb-6">
          <h1 id="lesson-title" className="text-3xl font-semibold tracking-tight">{d.lesson.title}</h1>
          <p className="mt-2 flex flex-wrap items-center gap-3 text-sm text-muted">
            <span className="flex items-center gap-1"><Clock className="h-3.5 w-3.5" aria-hidden /> {d.lesson.minutes} min</span>
            <span>{d.module.title}</span>
            <span className="hidden text-faint sm:inline">
              <Kbd>{modKey()}</Kbd> <Kbd>←</Kbd> <Kbd>→</Kbd> to move between lessons
            </span>
          </p>
        </header>

        <div className="space-y-10">
          <Section id="objectives" title="Learning objectives">
            <ul className="list-disc space-y-1 pl-5 text-sm">
              {d.lesson.objectives.map((o) => <li key={o}>{o}</li>)}
            </ul>
          </Section>

          <Section id="concept" title="Concept">
            <Markdown>{d.lesson.concept}</Markdown>
          </Section>

          <Section id="example" title={d.lesson.examples.length > 1 ? "Code examples" : "Code example"}>
            <div className="space-y-5">
              {d.lesson.examples.map((ex, i) => (
                <Example key={i} ex={ex} label={`${d.lesson.title} example ${i + 1}`} />
              ))}
            </div>
          </Section>

          {d.lesson.exercise && <ExerciseSection d={d} />}
          {d.lesson.quiz && d.lesson.quiz.length > 0 && <QuizSection d={d} />}

          {d.challenges && d.challenges.length > 0 && (
            <Section id="challenge" title="Coding challenge">
              <ul className="space-y-2">
                {d.challenges.map((c) => (
                  <li key={c.id}>
                    <Link to={`/challenge/${c.id}`} className="flex items-center gap-3 rounded-lg border border-border bg-surface px-4 py-3 hover:border-accent">
                      {c.passed ? <CheckCircle2 className="h-4 w-4 text-success" aria-label="Passed" /> : <Circle className="h-4 w-4 text-muted" aria-label="Not passed" />}
                      <span className="flex-1 font-medium">{c.title}</span>
                      <DifficultyBadge level={c.difficulty} />
                    </Link>
                  </li>
                ))}
              </ul>
            </Section>
          )}

          {d.lesson.exercise && <SolutionSection d={d} />}

          <Section id="takeaways" title="Key takeaways">
            <ul className="space-y-2">
              {d.lesson.takeaways.map((t) => (
                <li key={t} className="flex gap-2 text-sm">
                  <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-accent" aria-hidden />
                  {t}
                </li>
              ))}
            </ul>
          </Section>

          <CompleteBar d={d} />

          <nav aria-label="Lesson navigation" className="flex items-stretch justify-between gap-3 border-t border-border pt-6">
            {d.prev ? (
              <Link to={`/learn/${d.prev.module}/${d.prev.slug}`} className="flex min-w-0 flex-1 items-center gap-2 rounded-lg border border-border p-3 text-sm hover:border-accent">
                <ArrowLeft className="h-4 w-4 shrink-0" aria-hidden />
                <span className="min-w-0"><span className="block text-xs text-faint">Previous</span><span className="block truncate">{d.prev.title}</span></span>
              </Link>
            ) : <span className="flex-1" />}
            {d.next ? (
              <Link to={`/learn/${d.next.module}/${d.next.slug}`} className="flex min-w-0 flex-1 items-center justify-end gap-2 rounded-lg border border-border p-3 text-right text-sm hover:border-accent">
                <span className="min-w-0"><span className="block text-xs text-faint">Next lesson</span><span className="block truncate">{d.next.title}</span></span>
                <ArrowRight className="h-4 w-4 shrink-0" aria-hidden />
              </Link>
            ) : (
              <LinkButton to={`/learn/${d.module.id}`} className="flex-1">Back to module</LinkButton>
            )}
          </nav>
        </div>
      </article>

      <aside className="hidden xl:block" aria-label="Lesson progress">
        <div className="sticky top-20 max-h-[calc(100vh-6rem)] space-y-4 overflow-y-auto">
          <ProgressPanel d={d} />
        </div>
      </aside>

      <Sheet open={curriculumOpen} onOpenChange={setCurriculumOpen} title="Curriculum">
        <div className="p-3">
          <CurriculumNav d={d} onNavigate={() => setCurriculumOpen(false)} />
        </div>
      </Sheet>
      <Sheet open={progressOpen} onOpenChange={setProgressOpen} title="Progress" side="right">
        <div className="space-y-4 p-4"><ProgressPanel d={d} /></div>
      </Sheet>
      {!user && (
        <p className="col-span-full mt-6 rounded-lg border border-border bg-surface px-4 py-3 text-sm text-muted">
          You're reading as a guest. <Link className="text-accent underline" to={`/register`}>Create a free account</Link> to run code, take knowledge checks and track progress.
        </p>
      )}
    </div>
  );
}

function Section({ id, title, children }: { id: string; title: string; children: React.ReactNode }) {
  return (
    <section aria-labelledby={`${id}-h`} id={id} className="scroll-mt-24">
      <h2 id={`${id}-h`} className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted">{title}</h2>
      {children}
    </section>
  );
}

function CurriculumNav({ d, onNavigate }: { d: LessonDetail; onNavigate?: () => void }) {
  return (
    <nav aria-label={`Module ${d.module.number} lessons`}>
      <Link to={`/learn/${d.module.id}`} onClick={onNavigate} className="mb-2 flex items-center gap-1.5 px-2 text-xs font-semibold uppercase tracking-wider text-muted hover:text-fg">
        <ChevronsLeft className="h-3.5 w-3.5" aria-hidden /> Module {d.module.number}
      </Link>
      <p className="mb-2 px-2 text-sm font-semibold">{d.module.title}</p>
      <ol className="space-y-0.5">
        {d.curriculum.map((l) => {
          const current = l.id === d.lesson.id;
          const planned = l.status !== "published";
          const cls = cn("flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm", current ? "bg-accent-soft font-medium text-accent" : planned ? "text-faint" : "text-muted hover:bg-surface-2 hover:text-fg");
          const icon = current ? <span className="h-2.5 w-2.5 rounded-full bg-accent" aria-hidden /> : l.completed ? <CheckCircle2 className="h-4 w-4 text-success" aria-hidden /> : planned ? <Clock className="h-4 w-4" aria-hidden /> : <Circle className="h-4 w-4" aria-hidden />;
          return (
            <li key={l.id}>
              {planned ? (
                <span className={cls} aria-label={`${l.title} (coming soon)`}>{icon}<span className="truncate">{l.title}</span></span>
              ) : (
                <Link to={`/learn/${l.id}`} onClick={onNavigate} aria-current={current ? "page" : undefined} className={cls}>
                  <span className="flex w-4 justify-center">{icon}</span>
                  <span className="truncate">{l.title}</span>
                  {l.completed && <span className="sr-only">(completed)</span>}
                </Link>
              )}
            </li>
          );
        })}
      </ol>
    </nav>
  );
}

function ProgressPanel({ d }: { d: LessonDetail }) {
  const { user } = useAuth();
  const cfg = useConfig();
  const mods = useModules(!!user);
  const runnerUp = cfg.data?.runner.available ?? true;
  const p = d.progress;
  const mp = mods.data?.find((m) => m.id === d.module.id)?.progress;
  const items = [
    { label: "Read the lesson", done: !!p?.completed || false, note: "" },
    ...(d.requirements.exercise ? [{ label: "Pass the exercise", done: !!p?.exercisePassed, note: runnerUp ? "" : "Runner unavailable — optional" }] : []),
    ...(d.requirements.quiz ? [{ label: "Pass the knowledge check", done: !!p?.quizPassed, note: "" }] : []),
    { label: "Mark complete", done: !!p?.completed, note: "" },
  ];
  const doneCount = items.filter((i) => i.done).length;
  const pct = Math.round((doneCount / items.length) * 100);
  return (
    <>
      <Card>
        <CardHeader title="This lesson" action={<span className="font-mono text-sm text-muted">{user ? `${pct}%` : "—"}</span>} />
        <div className="space-y-3 p-4">
          {user && <ProgressBar value={pct} label="Lesson progress" tone={pct === 100 ? "success" : "accent"} />}
          <ul className="space-y-2 text-sm">
            {items.map((i) => (
              <li key={i.label} className="flex items-start gap-2">
                {i.done ? <CheckCircle2 className="mt-0.5 h-4 w-4 text-success" aria-label="done" /> : <Circle className="mt-0.5 h-4 w-4 text-faint" aria-label="to do" />}
                <span className={i.done ? "text-muted" : ""}>{i.label}{i.note && <span className="block text-xs text-faint">{i.note}</span>}</span>
              </li>
            ))}
          </ul>
        </div>
      </Card>
      {mp && mp.totalLessons > 0 && (
        <Card>
          <CardHeader title={d.module.title} action={<span className="font-mono text-sm text-muted">{mp.percent}%</span>} />
          <div className="p-4">
            <ProgressBar value={mp.percent} label="Module progress" tone={mp.percent === 100 ? "success" : "accent"} />
            <p className="mt-2 text-xs text-muted">{mp.completedLessons} of {mp.totalLessons} lessons</p>
          </div>
        </Card>
      )}
      {d.glossary && d.glossary.length > 0 && (
        <Card>
          <CardHeader title="Glossary" />
          <ul className="flex flex-wrap gap-1.5 p-4">
            {d.glossary.map((t) => (
              <li key={t.id}><Link to={`/glossary#${t.id}`}><Badge tone="accent">{t.term}</Badge></Link></li>
            ))}
          </ul>
        </Card>
      )}
    </>
  );
}

function Example({ ex, label }: { ex: LessonDetail["lesson"]["examples"][number]; label: string }) {
  const [editing, setEditing] = useState(false);
  return (
    <div className="space-y-2">
      {ex.caption && <Markdown>{ex.caption}</Markdown>}
      {editing ? (
        <RunPanel
          label={label}
          initial={ex.code}
          onRun={async (code) => ({ run: (await send<Outcome["run"]>("POST", "/api/run", { code }))! })}
        />
      ) : (
        <div className="relative">
          <CodeBlock code={ex.code} lang={ex.lang} />
          {ex.runnable && (
            <Button size="sm" className="absolute right-2 top-2" onClick={() => setEditing(true)} aria-label="Open this example in an editor to run it">
              Try it
            </Button>
          )}
        </div>
      )}
      {!ex.runnable && ex.lang === "go" && <p className="text-xs text-faint">This snippet is for reading — it depends on packages or infrastructure the sandbox doesn't provide.</p>}
    </div>
  );
}

function ExerciseSection({ d }: { d: LessonDetail }) {
  const ex = d.lesson.exercise!;
  const qc = useQueryClient();
  const toast = useToast();
  const [draft, setDraft] = useDraft(`golearn.draft.${d.lesson.id}`);
  const passed = d.progress?.exercisePassed;
  return (
    <Section id="exercise" title="Interactive exercise">
      <div className="space-y-3">
        <div className="flex items-start justify-between gap-3">
          <Markdown className="flex-1">{ex.prompt}</Markdown>
          {passed && <Badge tone="success"><CheckCircle2 className="h-3 w-3" aria-hidden /> Passed</Badge>}
        </div>
        <details className="text-sm">
          <summary className="cursor-pointer text-muted">Expected output</summary>
          <pre tabIndex={0} className="mt-2 overflow-x-auto rounded-lg border border-border bg-code-bg p-3 font-mono text-xs">{ex.expected}</pre>
        </details>
        <RunPanel
          label={`Exercise: ${d.lesson.title}`}
          initial={draft ?? ex.starter}
          resetTo={ex.starter}
          onCodeChange={(c) => setDraft(c === ex.starter ? null : c)}
          runLabel="Run code"
          expected={ex.expected}
          onRun={async (code): Promise<Outcome> => {
            const r = await send<ExerciseResult>("POST", `/api/lessons/${d.lesson.id}/exercise`, { code });
            if (r.passed) {
              toast.achievement(r.achievements.map((a) => a.title));
              void qc.invalidateQueries({ queryKey: ["lesson", d.lesson.moduleId, d.lesson.slug] });
            }
            return { run: r.run, expected: r.expected, verdict: r.passed ? { ok: true, text: "Output matches. Exercise passed." } : { ok: false, text: "Output doesn't match the expected output yet." } };
          }}
        />
      </div>
    </Section>
  );
}

function SolutionSection({ d }: { d: LessonDetail }) {
  const [open, setOpen] = useState(false);
  const [solution, setSolution] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);
  useEffect(() => {
    if (!open || solution !== null) return;
    fetch(`/api/lessons/${d.lesson.id}/solution`)
      .then((r) => r.json())
      .then((j: { data?: { solution: string }; error?: { message: string } }) => {
        if (j.data) setSolution(j.data.solution);
        else setErr(j.error?.message ?? "Couldn't load the solution.");
      })
      .catch(() => setErr("Couldn't load the solution."));
  }, [open, solution, d.lesson.id]);
  return (
    <Section id="solution" title="Solution">
      {!open ? (
        <div className="flex flex-wrap items-center gap-3">
          <Button onClick={() => setOpen(true)}>Show solution</Button>
          <span className="text-sm text-muted">Try the exercise first — struggling is part of learning.</span>
        </div>
      ) : err ? (
        <p role="alert" className="text-sm text-danger">{err}</p>
      ) : solution === null ? (
        <p className="text-sm text-muted">Loading…</p>
      ) : (
        <CodeBlock code={solution} lang="go" />
      )}
    </Section>
  );
}

type Answers = Record<string, number | number[] | string>;

function QuizSection({ d }: { d: LessonDetail }) {
  const { user } = useAuth();
  const qc = useQueryClient();
  const toast = useToast();
  const questions = d.lesson.quiz!;
  const [answers, setAnswers] = useState<Answers>({});
  const [result, setResult] = useState<QuizResult | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const byId = useMemo(() => new Map(result?.results.map((r) => [r.id, r])), [result]);
  const answered = questions.every((q) => answers[q.id] !== undefined && answers[q.id] !== "" && !(Array.isArray(answers[q.id]) && (answers[q.id] as number[]).length === 0));

  const submit = async () => {
    setBusy(true);
    setError(null);
    try {
      const r = await send<QuizResult>("POST", `/api/lessons/${d.lesson.id}/quiz`, { answers });
      setResult(r);
      toast.achievement(r.achievements.map((a) => a.title));
      void qc.invalidateQueries({ queryKey: ["lesson", d.lesson.moduleId, d.lesson.slug] });
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't submit.");
    } finally {
      setBusy(false);
    }
  };

  return (
    <Section id="check" title="Knowledge check">
      <div className="space-y-5">
        {d.progress?.quizPassed && !result && <Badge tone="success"><CheckCircle2 className="h-3 w-3" aria-hidden /> Passed</Badge>}
        {questions.map((q, i) => (
          <Question key={q.id} q={q} n={i + 1} value={answers[q.id]} onChange={(v) => setAnswers((a) => ({ ...a, [q.id]: v }))} result={byId.get(q.id)} disabled={busy || !!result} />
        ))}
        {error && <p role="alert" className="text-sm text-danger">{error}</p>}
        <div aria-live="polite" className="space-y-3">
          {result && (
            <p className={cn("flex items-center gap-2 text-sm font-medium", result.passed ? "text-success" : "text-danger")}>
              {result.passed ? <CheckCircle2 className="h-4 w-4" aria-hidden /> : <Circle className="h-4 w-4" aria-hidden />}
              {result.score}/{result.total} correct — {result.passed ? "knowledge check passed." : "not quite. Review the explanations and try again."}
            </p>
          )}
        </div>
        <div className="flex gap-2">
          {!result ? (
            <Button variant="primary" onClick={submit} disabled={!user || busy || !answered}>
              {busy ? "Checking…" : "Check answers"}
            </Button>
          ) : (
            <Button onClick={() => { setResult(null); setAnswers({}); }}>{result.passed ? "Retake" : "Try again"}</Button>
          )}
          {!user && <span className="self-center text-sm text-muted"><Link className="text-accent underline" to="/login">Sign in</Link> to submit.</span>}
        </div>
      </div>
    </Section>
  );
}

const typeLabel: Record<string, string> = { mcq: "Multiple choice", tf: "True / false", output: "Predict the output", debug: "Debug", multi: "Select all that apply", short: "Short answer" };

function Question({ q, n, value, onChange, result, disabled }: { q: QuizQuestion; n: number; value: Answers[string] | undefined; onChange: (v: Answers[string]) => void; result?: { correct: boolean; explanation: string }; disabled: boolean }) {
  return (
    <fieldset className="rounded-xl border border-border bg-surface p-4" disabled={disabled}>
      <legend className="px-1 text-xs text-muted">Question {n} · {typeLabel[q.type] ?? q.type}</legend>
      <p className="mb-2 font-medium">{q.prompt}</p>
      {q.code && <div className="mb-3"><CodeBlock code={q.code} lang="go" /></div>}
      {q.type === "short" ? (
        <input
          aria-label={q.prompt}
          value={(value as string) ?? ""}
          onChange={(e) => onChange(e.target.value)}
          className="h-10 w-full max-w-sm rounded-lg border border-border bg-bg px-3 font-mono text-sm outline-none focus:border-accent"
          autoComplete="off"
          spellCheck={false}
        />
      ) : (
        <div className="space-y-1.5">
          {q.options?.map((opt, oi) => {
            const multi = q.type === "multi";
            const checked = multi ? ((value as number[] | undefined) ?? []).includes(oi) : value === oi;
            return (
              <label key={oi} className={cn("flex cursor-pointer items-start gap-2.5 rounded-lg border px-3 py-2 text-sm", checked ? "border-accent bg-accent-soft" : "border-border hover:border-border-strong")}>
                <input
                  type={multi ? "checkbox" : "radio"}
                  name={`q-${q.id}`}
                  checked={checked}
                  onChange={() => {
                    if (multi) {
                      const cur = (value as number[] | undefined) ?? [];
                      onChange(cur.includes(oi) ? cur.filter((x) => x !== oi) : [...cur, oi]);
                    } else onChange(oi);
                  }}
                  className="mt-0.5 accent-[var(--accent)]"
                />
                <span className={q.type === "output" || q.type === "debug" ? "font-mono text-[0.85rem]" : ""}>{opt}</span>
              </label>
            );
          })}
        </div>
      )}
      {result && (
        <div className={cn("mt-3 rounded-lg px-3 py-2 text-sm", result.correct ? "bg-success-soft text-success" : "bg-danger-soft text-danger")} role="status">
          <p className="font-semibold">{result.correct ? "Correct" : "Incorrect"}</p>
          <p className="text-fg">{result.explanation}</p>
        </div>
      )}
    </fieldset>
  );
}

function CompleteBar({ d }: { d: LessonDetail }) {
  const { user } = useAuth();
  const qc = useQueryClient();
  const toast = useToast();
  const cfg = useConfig();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const p = d.progress;
  const runnerUp = cfg.data?.runner.available ?? true;
  const missing: string[] = [];
  if (d.requirements.exercise && !p?.exercisePassed && runnerUp) missing.push("pass the exercise");
  if (d.requirements.quiz && !p?.quizPassed) missing.push("pass the knowledge check");

  const toggle = async (completed: boolean) => {
    setBusy(true);
    setError(null);
    try {
      const r = await send<{ achievements: { title: string }[] }>("POST", `/api/progress/lessons/${d.lesson.id}`, { completed });
      if (completed) toast.achievement(r.achievements.map((a) => a.title));
      await qc.invalidateQueries({ predicate: (q) => q.queryKey[0] !== "me" });
    } catch (e) {
      setError(e instanceof Error ? e.message : "Couldn't save progress.");
    } finally {
      setBusy(false);
    }
  };

  if (!user) return null;
  return (
    <Card className="flex flex-wrap items-center justify-between gap-3 p-4">
      <div>
        <p className="font-medium">{p?.completed ? "Lesson complete" : "Finished this lesson?"}</p>
        <p className="text-sm text-muted">
          {p?.completed ? "Progress saved." : missing.length ? `To complete: ${missing.join(" and ")}.` : "Mark it complete to track your progress."}
        </p>
        {error && <p role="alert" className="mt-1 text-sm text-danger">{error}</p>}
      </div>
      <div className="flex gap-2">
        {p?.completed ? (
          <>
            <Button variant="ghost" onClick={() => toggle(false)} disabled={busy}>Undo</Button>
            {d.next && <LinkButton to={`/learn/${d.next.module}/${d.next.slug}`} variant="primary">Next lesson <ArrowRight className="h-4 w-4" aria-hidden /></LinkButton>}
          </>
        ) : (
          <Button variant="primary" onClick={() => toggle(true)} disabled={busy || missing.length > 0}>
            {busy ? "Saving…" : "Mark lesson complete"}
          </Button>
        )}
      </div>
    </Card>
  );
}
