import { useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { ArrowDown } from "lucide-react";
import { useModules } from "@/api/queries";
import type { ModuleSummary } from "@/api/types";
import { Async } from "@/components/ui/async";
import { Badge, PageHeader, ProgressBar } from "@/components/ui/ui";
import { StateBadge, StateIcon, trackNames } from "@/components/common";
import { useAuth } from "@/lib/auth";
import { cn } from "@/lib/utils";

type PathFilter = "all" | "beginner" | "pro";

export function Roadmap() {
  const { user } = useAuth();
  const q = useModules(!!user);
  const [path, setPath] = useState<PathFilter>("all");

  return (
    <>
      <PageHeader
        title="GoLearn roadmap"
        description="The full curriculum from installing Go to Kubernetes controllers. Locked modules are recommendations, not barriers — you can open any module."
        actions={
          <div role="group" aria-label="Filter by learning path" className="flex rounded-lg border border-border bg-surface p-0.5 text-sm">
            {(["all", "beginner", "pro"] as const).map((p) => (
              <button
                key={p}
                aria-pressed={path === p}
                onClick={() => setPath(p)}
                className={cn("rounded-md px-3 py-1.5", path === p ? "bg-accent-soft font-medium text-accent" : "text-muted hover:text-fg")}
              >
                {p === "all" ? "All" : p === "beginner" ? "Beginner" : "Professional / DevOps"}
              </button>
            ))}
          </div>
        }
      />
      <Async query={q} isEmpty={(d) => d.length === 0} empty={<p className="text-sm text-muted">No modules published yet.</p>}>
        {(mods) => (
          <Flow
            modules={mods.filter((m) => path === "all" || m.paths.includes(path))}
            authed={!!user}
            titles={Object.fromEntries(mods.map((m) => [m.id, m.title]))}
          />
        )}
      </Async>
    </>
  );
}

function Flow({ modules, authed, titles }: { modules: ModuleSummary[]; authed: boolean; titles: Record<string, string> }) {
  const groups = useMemo(() => {
    const order: string[] = [];
    const by = new Map<string, ModuleSummary[]>();
    for (const m of modules) {
      if (!by.has(m.track)) {
        by.set(m.track, []);
        order.push(m.track);
      }
      by.get(m.track)!.push(m);
    }
    return order.map((t) => ({ track: t, items: by.get(t)! }));
  }, [modules]);

  return (
    <div className="mx-auto max-w-2xl">
      {!authed && (
        <p className="mb-6 rounded-lg border border-border bg-surface px-4 py-3 text-sm text-muted">
          <Link to="/register" className="text-accent underline">
            Create an account
          </Link>{" "}
          to see your completion state on each module.
        </p>
      )}
      {groups.map((g) => (
        <section key={g.track} aria-labelledby={`track-${g.track}`} className="mb-8">
          <h2 id={`track-${g.track}`} className="mb-3 text-xs font-semibold uppercase tracking-wider text-muted">
            {trackNames[g.track] ?? g.track}
          </h2>
          <ol className="flex flex-col items-stretch">
            {g.items.map((m, i) => {
              const p = m.progress;
              const state = p?.state ?? "available";
              const planned = m.lessons.filter((l) => l.status === "planned").length;
              return (
                <li key={m.id} className="flex flex-col items-stretch">
                  <Link
                    to={`/learn/${m.id}`}
                    className={cn(
                      "group rounded-xl border bg-surface p-4 shadow-card transition-colors hover:border-accent",
                      state === "in_progress" ? "border-accent" : "border-border",
                      state === "locked" && "opacity-80",
                    )}
                  >
                    <div className="flex items-start gap-3">
                      <StateIcon state={state} className="mt-1 h-5 w-5 shrink-0" />
                      <div className="min-w-0 flex-1">
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="font-mono text-xs text-muted">{m.number}</span>
                          <h3 className="font-semibold">{m.title}</h3>
                          {authed && <StateBadge state={state} />}
                        </div>
                        <p className="mt-1 text-sm text-muted">{m.summary}</p>
                        <div className="mt-3 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-faint">
                          <span>{m.lessons.length - planned} lessons</span>
                          {planned > 0 && <Badge tone="neutral">{planned} more planned</Badge>}
                          {m.project && <span>Project</span>}
                        </div>
                        {authed && p && p.totalLessons > 0 && <ProgressBar className="mt-3" value={p.percent} label={`${m.title} progress`} tone={p.percent === 100 ? "success" : "accent"} />}
                        {authed && state === "locked" && p?.blockedBy && (
                          <p className="mt-2 text-xs text-faint">Recommended first: {p.blockedBy.map((id) => titles[id] ?? id).join(", ")}</p>
                        )}
                      </div>
                    </div>
                  </Link>
                  {i < g.items.length - 1 && <ArrowDown className="mx-auto my-1.5 h-4 w-4 text-faint" aria-hidden />}
                </li>
              );
            })}
          </ol>
        </section>
      ))}
    </div>
  );
}
