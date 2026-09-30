import { useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { CheckCircle2, Clock, Circle } from "lucide-react";
import { useModule, useModules } from "@/api/queries";
import { Async } from "@/components/ui/async";
import { Badge, Card, CardHeader, LinkButton, PageHeader, ProgressBar } from "@/components/ui/ui";
import { DifficultyBadge, StateBadge, StateIcon, trackNames } from "@/components/common";
import { useAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";

export function LearnIndex() {
  const { user } = useAuth();
  const q = useModules(!!user);
  const [path, setPath] = useState<"all" | "beginner" | "pro">(user?.profile.learningPath ?? "all");
  const groups = useMemo(() => {
    const mods = (q.data ?? []).filter((m) => path === "all" || m.paths.includes(path));
    const order: string[] = [];
    const by = new Map<string, typeof mods>();
    for (const m of mods) {
      if (!by.has(m.track)) {
        by.set(m.track, []);
        order.push(m.track);
      }
      by.get(m.track)!.push(m);
    }
    return order.map((t) => ({ track: t, items: by.get(t)! }));
  }, [q.data, path]);

  return (
    <>
      <PageHeader
        title="Learn"
        description="Pick a module. Each lesson is short, ends with a knowledge check, and most include a coding exercise."
        actions={
          <div role="group" aria-label="Filter by learning path" className="flex rounded-lg border border-border bg-surface p-0.5 text-sm">
            {(["all", "beginner", "pro"] as const).map((p) => (
              <button key={p} aria-pressed={path === p} onClick={() => setPath(p)} className={cn("rounded-md px-3 py-1.5", path === p ? "bg-accent-soft font-medium text-accent" : "text-muted hover:text-fg")}>
                {p === "all" ? "All modules" : p === "beginner" ? "Beginner" : "Professional / DevOps"}
              </button>
            ))}
          </div>
        }
      />
      <Async query={q} isEmpty={(d) => d.length === 0} empty={<p className="text-sm text-muted">No modules yet.</p>}>
        {() =>
          groups.length === 0 ? (
            <p className="text-sm text-muted">No modules for this path.</p>
          ) : (
            <div className="space-y-8">
              {groups.map((g) => (
                <section key={g.track} aria-labelledby={`t-${g.track}`}>
                  <h2 id={`t-${g.track}`} className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted">
                    {trackNames[g.track] ?? g.track}
                  </h2>
                  <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                    {g.items.map((m) => {
                      const published = m.lessons.filter((l) => l.status === "published").length;
                      const p = m.progress;
                      return (
                        <Link key={m.id} to={`/learn/${m.id}`} className="group flex flex-col rounded-xl border border-border bg-surface p-4 shadow-card transition-colors hover:border-accent">
                          <div className="flex items-center justify-between">
                            <span className="font-mono text-xs text-muted">Module {m.number}</span>
                            {p && <StateBadge state={p.state} />}
                          </div>
                          <h3 className="mt-2 font-semibold group-hover:text-accent">{m.title}</h3>
                          <p className="mt-1 line-clamp-2 flex-1 text-sm text-muted">{m.summary}</p>
                          <p className="mt-3 text-xs text-faint">
                            {published} lessons{m.lessons.length > published ? ` · ${m.lessons.length - published} planned` : ""}
                          </p>
                          {p && p.totalLessons > 0 && <ProgressBar className="mt-2" value={p.percent} label={`${m.title} progress`} tone={p.percent === 100 ? "success" : "accent"} />}
                        </Link>
                      );
                    })}
                  </div>
                </section>
              ))}
            </div>
          )
        }
      </Async>
    </>
  );
}

export function ModulePage() {
  const { module = "" } = useParams();
  const q = useModule(module);
  const all = useModules(!!useAuth().user);

  return (
    <Async query={q}>
      {({ module: m, challenges, project }) => {
        const p = m.progress;
        const blockers = (p?.blockedBy ?? []).map((id) => all.data?.find((x) => x.id === id)).filter(Boolean);
        const firstOpen = m.lessons.find((l) => l.status === "published" && !l.completed) ?? m.lessons.find((l) => l.status === "published");
        return (
          <>
            <nav aria-label="Breadcrumb" className="mb-3 text-sm text-muted">
              <Link to="/learn" className="hover:text-fg">
                Learn
              </Link>{" "}
              / <span className="text-fg">Module {m.number}</span>
            </nav>
            <PageHeader
              title={m.title}
              description={m.summary}
              actions={
                firstOpen && (
                  <LinkButton to={`/learn/${firstOpen.id}`} variant="primary">
                    {p && p.completedLessons > 0 ? "Continue module" : "Start module"}
                  </LinkButton>
                )
              }
            />
            {p?.state === "locked" && blockers.length > 0 && (
              <p role="note" className="mb-5 rounded-lg border border-warn/40 bg-warn-soft px-4 py-3 text-sm text-warn">
                This module builds on{" "}
                {blockers.map((b, i) => (
                  <span key={b!.id}>
                    {i > 0 && ", "}
                    <Link className="underline" to={`/learn/${b!.id}`}>
                      {b!.title}
                    </Link>
                  </span>
                ))}
                . You can still start here, but you may find it easier after finishing those.
              </p>
            )}
            {p && p.totalLessons > 0 && (
              <div className="mb-6 flex items-center gap-4">
                <ProgressBar value={p.percent} label="Module progress" className="max-w-sm" tone={p.percent === 100 ? "success" : "accent"} />
                <span className="text-sm text-muted">
                  {p.completedLessons}/{p.totalLessons} lessons
                </span>
              </div>
            )}
            <div className="grid gap-5 lg:grid-cols-3">
              <Card className="lg:col-span-2">
                <CardHeader title="Lessons" />
                <ol>
                  {m.lessons.map((l, i) => {
                    const planned = l.status !== "published";
                    const inner = (
                      <>
                        <span className="w-6 shrink-0 text-right font-mono text-xs text-faint">{i + 1}</span>
                        {planned ? (
                          <Clock className="h-4 w-4 shrink-0 text-faint" aria-hidden />
                        ) : l.completed ? (
                          <CheckCircle2 className="h-4 w-4 shrink-0 text-success" aria-label="Completed" />
                        ) : (
                          <Circle className="h-4 w-4 shrink-0 text-muted" aria-label="Not completed" />
                        )}
                        <span className="min-w-0 flex-1 truncate">{l.title}</span>
                        {planned ? <Badge>Coming soon</Badge> : <span className="text-xs text-faint">{l.minutes} min</span>}
                      </>
                    );
                    return (
                      <li key={l.id} className="border-b border-border last:border-b-0">
                        {planned ? (
                          <div className="flex items-center gap-3 px-4 py-2.5 text-sm text-muted">{inner}</div>
                        ) : (
                          <Link to={`/learn/${l.id}`} className="flex items-center gap-3 px-4 py-2.5 text-sm hover:bg-surface-2">
                            {inner}
                          </Link>
                        )}
                      </li>
                    );
                  })}
                </ol>
              </Card>
              <div className="space-y-5">
                {project && (
                  <Card>
                    <CardHeader title="Module project" />
                    <div className="p-4">
                      <p className="font-medium">{project.title}</p>
                      <p className="mt-1 text-sm text-muted">{project.summary}</p>
                      <LinkButton to={`/projects/${project.id}`} className="mt-3" size="sm">
                        View project
                      </LinkButton>
                    </div>
                  </Card>
                )}
                {challenges && challenges.length > 0 && (
                  <Card>
                    <CardHeader title="Challenges" />
                    <ul>
                      {challenges.map((c) => (
                        <li key={c.id} className="border-b border-border last:border-b-0">
                          <Link to={`/challenge/${c.id}`} className="flex items-center gap-2 px-4 py-2.5 text-sm hover:bg-surface-2">
                            {c.passed ? <CheckCircle2 className="h-4 w-4 text-success" aria-label="Passed" /> : <StateIcon state="available" />}
                            <span className="flex-1 truncate">{c.title}</span>
                            <DifficultyBadge level={c.difficulty} />
                          </Link>
                        </li>
                      ))}
                    </ul>
                  </Card>
                )}
              </div>
            </div>
          </>
        );
      }}
    </Async>
  );
}
