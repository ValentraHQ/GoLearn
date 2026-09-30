import { useState } from "react";
import { Link, useParams } from "react-router-dom";
import { useQueryClient } from "@tanstack/react-query";
import { CheckCircle2, Clock } from "lucide-react";
import { useProject, useProjects } from "@/api/queries";
import { send } from "@/api/client";
import { Async } from "@/components/ui/async";
import { Badge, Card, CardHeader, PageHeader, ProgressBar } from "@/components/ui/ui";
import { Markdown } from "@/components/Markdown";
import { useToast } from "@/components/ui/toast";
import { useAuth } from "@/lib/auth";
import { useDocumentTitle } from "@/lib/hooks";
import type { ProjectSummary } from "@/api/types";

const levels: { id: ProjectSummary["level"]; label: string }[] = [
  { id: "beginner", label: "Beginner" },
  { id: "intermediate", label: "Intermediate" },
  { id: "advanced", label: "Advanced" },
  { id: "capstone", label: "Final capstone" },
];

export function ProjectIndex() {
  const q = useProjects();
  return (
    <>
      <PageHeader title="Projects" description="Build software that grows in complexity. Projects are self-paced: work through the tasks locally and tick them off as you go." />
      <Async query={q} isEmpty={(d) => d.length === 0} empty={<p className="text-sm text-muted">No projects yet.</p>}>
        {(ps) => (
          <div className="space-y-8">
            {levels.map((l) => {
              const items = ps.filter((p) => p.level === l.id);
              if (!items.length) return null;
              return (
                <section key={l.id} aria-labelledby={`lv-${l.id}`}>
                  <h2 id={`lv-${l.id}`} className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted">{l.label}</h2>
                  <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                    {items.map((p) => (
                      <Link key={p.id} to={`/projects/${p.id}`} className="flex flex-col rounded-xl border border-border bg-surface p-4 shadow-card hover:border-accent">
                        <h3 className="font-semibold">{p.title}</h3>
                        <p className="mt-1 line-clamp-3 flex-1 text-sm text-muted">{p.summary}</p>
                        <div className="mt-3 flex items-center gap-3 text-xs text-faint">
                          <span className="flex items-center gap-1"><Clock className="h-3 w-3" aria-hidden />~{p.hours}h</span>
                          <span>{p.taskCount} tasks</span>
                        </div>
                        {p.tasksDone > 0 && <ProgressBar className="mt-2" value={(p.tasksDone / p.taskCount) * 100} label={`${p.title} progress`} tone={p.tasksDone === p.taskCount ? "success" : "accent"} />}
                      </Link>
                    ))}
                  </div>
                </section>
              );
            })}
          </div>
        )}
      </Async>
    </>
  );
}

export function ProjectPage() {
  const { id = "" } = useParams();
  const q = useProject(id);
  const { user } = useAuth();
  const qc = useQueryClient();
  const toast = useToast();
  const [pending, setPending] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  useDocumentTitle(q.data?.project.title);

  return (
    <Async query={q}>
      {({ project: p, done, module }) => {
        const count = Object.keys(done).length;
        const toggle = async (taskId: string, value: boolean) => {
          setPending(taskId);
          setError(null);
          try {
            const r = await send<{ achievements: { title: string }[] }>("PUT", `/api/projects/${p.id}/tasks/${taskId}`, { done: value });
            toast.achievement(r.achievements.map((a) => a.title));
            await qc.invalidateQueries({ predicate: (query) => query.queryKey[0] !== "me" });
          } catch (e) {
            setError(e instanceof Error ? e.message : "Couldn't save.");
          } finally {
            setPending(null);
          }
        };
        return (
          <>
            <nav aria-label="Breadcrumb" className="mb-3 text-sm text-muted"><Link to="/projects" className="hover:text-fg">Projects</Link> / <span className="text-fg">{p.title}</span></nav>
            <PageHeader title={p.title} description={p.summary} />
            <div className="mb-5 flex flex-wrap items-center gap-2 text-sm">
              <Badge tone="accent">{levels.find((l) => l.id === p.level)?.label}</Badge>
              <span className="text-muted">~{p.hours} hours</span>
              {module && <Link to={`/learn/${module.module}`} className="text-accent underline">{module.title}</Link>}
            </div>
            <div className="grid gap-5 lg:grid-cols-3">
              <div className="space-y-5 lg:col-span-2">
                <Card><CardHeader title="Overview" /><div className="p-5"><Markdown>{p.description}</Markdown></div></Card>
                {p.requirements && p.requirements.length > 0 && (
                  <Card><CardHeader title="Requirements" /><ul className="list-disc space-y-1.5 p-5 pl-9 text-sm">{p.requirements.map((r) => <li key={r}>{r}</li>)}</ul></Card>
                )}
                {p.architecture && (
                  <Card><CardHeader title="Architecture" /><div className="p-4"><Markdown>{"```text\n" + p.architecture.trim() + "\n```"}</Markdown></div></Card>
                )}
                {p.stretch && p.stretch.length > 0 && (
                  <Card><CardHeader title="Stretch goals" /><ul className="list-disc space-y-1.5 p-5 pl-9 text-sm">{p.stretch.map((r) => <li key={r}>{r}</li>)}</ul></Card>
                )}
              </div>
              <Card className="h-fit lg:sticky lg:top-20">
                <CardHeader title="Tasks" action={<span className="font-mono text-sm text-muted">{count}/{p.tasks.length}</span>} />
                <div className="p-4">
                  <ProgressBar value={(count / p.tasks.length) * 100} label="Project progress" tone={count === p.tasks.length ? "success" : "accent"} />
                  {count === p.tasks.length && <p className="mt-2 flex items-center gap-1.5 text-sm font-medium text-success"><CheckCircle2 className="h-4 w-4" aria-hidden />Project complete</p>}
                  <ol className="mt-4 space-y-3">
                    {p.tasks.map((t, i) => (
                      <li key={t.id}>
                        <label className="flex cursor-pointer items-start gap-2.5 text-sm">
                          <input type="checkbox" className="mt-1 accent-[var(--accent)]" checked={!!done[t.id]} disabled={!user || pending === t.id} onChange={(e) => toggle(t.id, e.target.checked)} />
                          <span>
                            <span className={done[t.id] ? "text-muted line-through" : "font-medium"}>{i + 1}. {t.title}</span>
                            {t.notes && <span className="mt-0.5 block text-xs text-muted">{t.notes}</span>}
                          </span>
                        </label>
                      </li>
                    ))}
                  </ol>
                  {!user && <p className="mt-4 text-xs text-muted"><Link className="text-accent underline" to="/login">Sign in</Link> to track tasks.</p>}
                  {error && <p role="alert" className="mt-3 text-sm text-danger">{error}</p>}
                </div>
              </Card>
            </div>
          </>
        );
      }}
    </Async>
  );
}
