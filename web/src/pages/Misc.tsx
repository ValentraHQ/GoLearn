import { useEffect, useMemo, useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router-dom";
import { Trophy, Lock } from "lucide-react";
import { useAchievements, useDashboard, useGlossary, useSkills, useModules } from "@/api/queries";
import { send } from "@/api/client";
import type { Profile } from "@/api/types";
import { Async } from "@/components/ui/async";
import { Badge, Button, Card, CardHeader, PageHeader, ProgressBar } from "@/components/ui/ui";
import { CodeBlock } from "@/components/Markdown";
import { RunPanel } from "@/components/RunPanel";
import { SkillLevelBadge } from "@/components/common";
import { Field } from "@/pages/Auth";
import { useAuth } from "@/lib/auth";
import { useTheme } from "@/lib/theme";
import { useDocumentTitle, useDraft, useLearningTimer } from "@/lib/hooks";
import { cn, formatDuration, skillLevelLabel } from "@/lib/utils";
import type { SkillLevel } from "@/api/types";

/* ---------------- Playground ---------------- */

const PLAYGROUND_STARTER = `package main

import "fmt"

func main() {
	fmt.Println("Hello GoLearn")
}
`;

export function Playground() {
  const [draft, setDraft] = useDraft("golearn.playground");
  useDocumentTitle("Playground");
  useLearningTimer();
  return (
    <>
      <PageHeader title="Playground" description="A scratch pad for Go. Code runs in an isolated container with no network, limited CPU and memory, and a 10-second time limit." />
      <RunPanel
        label="Playground code"
        initial={draft ?? PLAYGROUND_STARTER}
        resetTo={PLAYGROUND_STARTER}
        minHeight="20rem"
        maxHeight="36rem"
        onCodeChange={(c) => setDraft(c === PLAYGROUND_STARTER ? null : c)}
        onRun={async (code) => ({ run: await send<import("@/api/types").RunResult>("POST", "/api/run", { code }) })}
      />
      <p className="mt-3 text-xs text-faint">Single-file <code className="font-mono">package main</code> programs using the standard library only. Programs can’t read stdin.</p>
    </>
  );
}

/* ---------------- Skills ---------------- */

const LEVELS: SkillLevel[] = ["not_started", "learning", "practicing", "proficient", "mastered"];

export function Skills() {
  const q = useSkills();
  const mods = useModules(true);
  return (
    <>
      <PageHeader title="Skills" description="Skill levels are earned by finishing lessons and passing challenges that teach each skill — separate from module completion." />
      <Async query={q}>
        {(skills) => (
          <>
            <ul className="mb-5 flex flex-wrap gap-2 text-xs text-muted" aria-label="Skill level scale">
              {LEVELS.map((l, i) => <li key={l} className="flex items-center gap-1.5"><span className="font-mono text-faint">{i}</span>{skillLevelLabel[l]}</li>)}
            </ul>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {skills.map((s) => (
                <Card key={s.skillId} className={cn("p-4", s.total === 0 && "opacity-70")}>
                  <div className="flex items-start justify-between gap-2">
                    <h2 className="font-semibold">{s.name}</h2>
                    <SkillLevelBadge level={s.level} />
                  </div>
                  <ol className="mt-3 flex gap-1" aria-hidden>
                    {LEVELS.slice(1).map((l) => (
                      <li key={l} className={cn("h-1.5 flex-1 rounded-full", LEVELS.indexOf(s.level) >= LEVELS.indexOf(l) ? "bg-accent" : "bg-surface-2 border border-border")} />
                    ))}
                  </ol>
                  <ProgressBar className="mt-3" value={s.percent} label={`${s.name} progress`} tone={s.percent === 100 ? "success" : "accent"} />
                  <p className="mt-2 text-xs text-muted">
                    {s.total === 0 ? "Content for this skill is still being written." : `${s.done} of ${s.total} lessons and challenges`}
                  </p>
                  {s.modules && s.modules.length > 0 && (
                    <p className="mt-1 text-xs text-faint">
                      Taught in {s.modules.map((id, i) => (
                        <span key={id}>{i > 0 && ", "}<Link className="underline" to={`/learn/${id}`}>{mods.data?.find((m) => m.id === id)?.title ?? id}</Link></span>
                      ))}
                    </p>
                  )}
                </Card>
              ))}
            </div>
          </>
        )}
      </Async>
    </>
  );
}

/* ---------------- Achievements ---------------- */

export function Achievements() {
  const q = useAchievements();
  return (
    <>
      <PageHeader title="Achievements" description="Milestones for real progress. XP and achievements are optional — turn them off in Settings." />
      <Async query={q}>
        {(d) =>
          !d.gamification ? (
            <Card className="p-6 text-sm text-muted">Gamification is turned off. <Link className="text-accent underline" to="/settings">Enable it in Settings</Link> to see XP and achievements.</Card>
          ) : (
            <>
              <Card className="mb-5 p-4">
                <div className="flex items-center justify-between text-sm"><span className="font-medium">Level {d.level}</span><span className="font-mono text-xs text-muted">{d.xp} XP</span></div>
                <ProgressBar className="mt-3" value={(d.xpIntoLevel / Math.max(d.xpForNextLevel, 1)) * 100} label="Progress to next level" />
                <p className="mt-2 text-xs text-faint">Lessons 20 XP · exercises and quizzes 10 XP · challenges 10–100 XP by difficulty · projects 150 XP</p>
              </Card>
              <ul className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {d.achievements.map((a) => (
                  <li key={a.id}>
                    <Card className={cn("flex h-full gap-3 p-4", !a.unlocked && "opacity-60")}>
                      <span className={cn("flex h-9 w-9 shrink-0 items-center justify-center rounded-lg", a.unlocked ? "bg-warn-soft text-warn" : "bg-surface-2 text-faint")}>
                        {a.unlocked ? <Trophy className="h-4 w-4" aria-hidden /> : <Lock className="h-4 w-4" aria-hidden />}
                      </span>
                      <div>
                        <p className="font-semibold">{a.title}</p>
                        <p className="text-sm text-muted">{a.description}</p>
                        <p className="mt-1 text-xs text-faint">{a.unlocked ? `Unlocked${a.unlockedAt ? " " + new Date(a.unlockedAt).toLocaleDateString() : ""}` : "Locked"}</p>
                      </div>
                    </Card>
                  </li>
                ))}
              </ul>
            </>
          )
        }
      </Async>
    </>
  );
}

/* ---------------- Glossary ---------------- */

export function Glossary() {
  const q = useGlossary();
  const [filter, setFilter] = useState("");
  const terms = useMemo(() => {
    const f = filter.trim().toLowerCase();
    return (q.data ?? []).filter((t) => !f || `${t.term} ${t.simple} ${t.technical}`.toLowerCase().includes(f));
  }, [q.data, filter]);

  // Scroll to a term when arriving with #anchor.
  useEffect(() => {
    if (!q.data) return;
    const id = location.hash.slice(1);
    if (id) document.getElementById(id)?.scrollIntoView({ block: "start" });
  }, [q.data]);

  return (
    <>
      <PageHeader title="Glossary" description="Go terms explained twice: once simply, once precisely." />
      <label htmlFor="g-f" className="sr-only">Filter terms</label>
      <input id="g-f" value={filter} onChange={(e) => setFilter(e.target.value)} placeholder="Filter terms…" className="mb-5 h-9 w-full max-w-xs rounded-lg border border-border bg-surface px-3 text-sm outline-none focus:border-accent" />
      <Async query={q} isEmpty={(d) => d.length === 0}>
        {() =>
          terms.length === 0 ? (
            <p className="text-sm text-muted">No terms match “{filter}”.</p>
          ) : (
            <div className="grid gap-4 lg:grid-cols-2">
              {terms.map((t) => (
                <Card key={t.id} className="scroll-mt-24 p-5">
                  <h2 id={t.id} className="scroll-mt-24 text-lg font-semibold">{t.term}</h2>
                  <p className="mt-2 text-sm"><span className="font-semibold">In plain words: </span>{t.simple}</p>
                  <p className="mt-2 text-sm text-muted"><span className="font-semibold text-fg">Technically: </span>{t.technical}</p>
                  {t.code && <div className="mt-3"><CodeBlock code={t.code} lang="go" /></div>}
                  {t.relatedLessons.length > 0 && (
                    <p className="mt-3 flex flex-wrap items-center gap-1.5 text-xs">
                      <span className="text-faint">Related:</span>
                      {t.relatedLessons.map((l) => <Link key={l.id} to={`/learn/${l.module}/${l.slug}`}><Badge tone="accent">{l.title}</Badge></Link>)}
                    </p>
                  )}
                </Card>
              ))}
            </div>
          )
        }
      </Async>
    </>
  );
}

/* ---------------- Profile ---------------- */

export function ProfilePage() {
  const { user, updateProfile } = useAuth();
  const q = useDashboard();
  const [displayName, setDisplayName] = useState(user?.profile.displayName ?? "");
  const [bio, setBio] = useState(user?.profile.bio ?? "");
  const [status, setStatus] = useState<{ ok: boolean; text: string } | null>(null);
  const [busy, setBusy] = useState(false);
  if (!user) return null;

  const save = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setStatus(null);
    try {
      await updateProfile({ ...user.profile, displayName, bio });
      setStatus({ ok: true, text: "Profile saved." });
    } catch (err) {
      setStatus({ ok: false, text: err instanceof Error ? err.message : "Couldn't save." });
    } finally {
      setBusy(false);
    }
  };

  return (
    <>
      <PageHeader title="Profile" description={user.email} />
      <div className="grid gap-5 lg:grid-cols-2">
        <Card>
          <CardHeader title="About you" />
          <form onSubmit={save} className="space-y-4 p-5">
            <Field id="dn" label="Display name" value={displayName} onChange={setDisplayName} />
            <div>
              <label htmlFor="bio" className="mb-1 block text-sm font-medium">Bio</label>
              <textarea id="bio" value={bio} onChange={(e) => setBio(e.target.value)} rows={3} maxLength={300} className="w-full rounded-lg border border-border bg-bg px-3 py-2 text-sm outline-none focus:border-accent" />
            </div>
            {status && <p role={status.ok ? "status" : "alert"} className={cn("text-sm", status.ok ? "text-success" : "text-danger")}>{status.text}</p>}
            <Button type="submit" variant="primary" disabled={busy}>{busy ? "Saving…" : "Save profile"}</Button>
          </form>
        </Card>
        <Card>
          <CardHeader title="Your stats" />
          <Async query={q}>
            {({ summary: s }) => (
              <dl className="grid grid-cols-2 gap-4 p-5 text-sm">
                {[
                  ["Lessons completed", `${s.lessonsCompleted}/${s.lessonsTotal}`],
                  ["Challenges passed", `${s.challengesCompleted}/${s.challengesTotal}`],
                  ["Projects completed", `${s.projectsCompleted}/${s.projectsTotal}`],
                  ["Modules completed", s.modulesCompleted],
                  ["Current streak", `${s.currentStreak} days`],
                  ["Longest streak", `${s.longestStreak} days`],
                  ["Learning time", formatDuration(s.learningSeconds)],
                  ["Quiz average", s.quizzesPassed ? `${s.quizAveragePercent}%` : "—"],
                ].map(([k, v]) => (
                  <div key={String(k)}><dt className="text-muted">{k}</dt><dd className="font-mono text-lg font-semibold">{v}</dd></div>
                ))}
              </dl>
            )}
          </Async>
        </Card>
      </div>
    </>
  );
}

/* ---------------- Settings ---------------- */

export function SettingsPage() {
  const { user, updateProfile, logout } = useAuth();
  const { setPref } = useTheme();
  const nav = useNavigate();
  const [p, setP] = useState<Profile | null>(user?.profile ?? null);
  const [status, setStatus] = useState<{ ok: boolean; text: string } | null>(null);
  const [busy, setBusy] = useState(false);
  const [pw, setPw] = useState({ current: "", next: "" });
  const [pwStatus, setPwStatus] = useState<{ ok: boolean; text: string } | null>(null);
  const [del, setDel] = useState("");
  const [delErr, setDelErr] = useState<string | null>(null);
  const zones = useMemo(() => {
    try {
      return (Intl as unknown as { supportedValuesOf?: (k: string) => string[] }).supportedValuesOf?.("timeZone") ?? [];
    } catch {
      return [];
    }
  }, []);
  if (!user || !p) return null;
  const set = <K extends keyof Profile>(k: K, v: Profile[K]) => setP({ ...p, [k]: v });
  const num = (v: string) => Math.max(0, Math.min(20, Number(v) || 0));

  const save = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setStatus(null);
    try {
      await updateProfile(p);
      setPref(p.theme);
      setStatus({ ok: true, text: "Settings saved." });
    } catch (err) {
      setStatus({ ok: false, text: err instanceof Error ? err.message : "Couldn't save." });
    } finally {
      setBusy(false);
    }
  };
  const inp = "h-10 rounded-lg border border-border bg-bg px-3 text-sm outline-none focus:border-accent";

  return (
    <>
      <PageHeader title="Settings" />
      <div className="grid gap-5 lg:grid-cols-2">
        <Card>
          <CardHeader title="Learning preferences" />
          <form onSubmit={save} className="space-y-4 p-5">
            <div>
              <label htmlFor="path" className="mb-1 block text-sm font-medium">Learning path</label>
              <select id="path" className={cn(inp, "w-full")} value={p.learningPath} onChange={(e) => set("learningPath", e.target.value as Profile["learningPath"])}>
                <option value="beginner">Beginner — start from installing Go</option>
                <option value="pro">Professional / DevOps — skip the basics</option>
              </select>
            </div>
            <fieldset>
              <legend className="mb-1 text-sm font-medium">Daily goal</legend>
              <div className="grid grid-cols-3 gap-3">
                {([["dailyLessons", "Lessons"], ["dailyChallenges", "Challenges"], ["dailyQuizzes", "Quizzes"]] as const).map(([k, l]) => (
                  <div key={k}>
                    <label htmlFor={k} className="mb-1 block text-xs text-muted">{l}</label>
                    <input id={k} type="number" min={0} max={20} className={cn(inp, "w-full")} value={p[k]} onChange={(e) => set(k, num(e.target.value))} />
                  </div>
                ))}
              </div>
            </fieldset>
            <div>
              <label htmlFor="theme" className="mb-1 block text-sm font-medium">Theme</label>
              <select id="theme" className={cn(inp, "w-full")} value={p.theme} onChange={(e) => set("theme", e.target.value as Profile["theme"])}>
                <option value="system">Match system</option>
                <option value="light">Light</option>
                <option value="dark">Dark</option>
              </select>
            </div>
            <div>
              <label htmlFor="tz" className="mb-1 block text-sm font-medium">Timezone</label>
              <select id="tz" className={cn(inp, "w-full")} value={p.timezone} onChange={(e) => set("timezone", e.target.value)}>
                {(zones.length ? zones : [p.timezone]).map((z) => <option key={z}>{z}</option>)}
              </select>
              <p className="mt-1 text-xs text-faint">Used for streaks and daily goals.</p>
            </div>
            <label className="flex items-start gap-2.5 text-sm">
              <input type="checkbox" className="mt-1 accent-[var(--accent)]" checked={p.gamification} onChange={(e) => set("gamification", e.target.checked)} />
              <span><span className="font-medium">Show XP and achievements</span><span className="block text-xs text-muted">Optional. Progress tracking works either way.</span></span>
            </label>
            {status && <p role={status.ok ? "status" : "alert"} className={cn("text-sm", status.ok ? "text-success" : "text-danger")}>{status.text}</p>}
            <Button type="submit" variant="primary" disabled={busy}>{busy ? "Saving…" : "Save settings"}</Button>
          </form>
        </Card>

        <div className="space-y-5">
          <Card>
            <CardHeader title="Keyboard shortcuts" />
            <dl className="divide-y divide-border text-sm">
              {[["Ctrl/⌘ + K", "Global search"], ["Ctrl/⌘ + Enter", "Run code (in an editor)"], ["Ctrl/⌘ + →", "Next lesson"], ["Ctrl/⌘ + ←", "Previous lesson"], ["Esc", "Close a panel, or leave the editor"]].map(([k, v]) => (
                <div key={k} className="flex items-center justify-between px-5 py-2.5"><dt className="text-muted">{v}</dt><dd><kbd className="rounded border border-border bg-surface-2 px-1.5 py-0.5 font-mono text-xs">{k}</kbd></dd></div>
              ))}
            </dl>
          </Card>

          <Card>
            <CardHeader title="Change password" />
            <form className="space-y-3 p-5" onSubmit={async (e) => {
              e.preventDefault();
              setPwStatus(null);
              try {
                await send("POST", "/api/auth/password", { current: pw.current, new: pw.next });
                setPw({ current: "", next: "" });
                setPwStatus({ ok: true, text: "Password changed. Other devices were signed out." });
              } catch (err) {
                setPwStatus({ ok: false, text: err instanceof Error ? err.message : "Couldn't change password." });
              }
            }}>
              <Field id="pw-c" type="password" label="Current password" autoComplete="current-password" value={pw.current} onChange={(v) => setPw({ ...pw, current: v })} />
              <Field id="pw-n" type="password" label="New password" autoComplete="new-password" hint="At least 10 characters." value={pw.next} onChange={(v) => setPw({ ...pw, next: v })} />
              {pwStatus && <p role={pwStatus.ok ? "status" : "alert"} className={cn("text-sm", pwStatus.ok ? "text-success" : "text-danger")}>{pwStatus.text}</p>}
              <Button type="submit" disabled={!pw.current || !pw.next}>Change password</Button>
            </form>
          </Card>

          <Card className="border-danger/40">
            <CardHeader title="Delete account" />
            <form className="space-y-3 p-5" onSubmit={async (e) => {
              e.preventDefault();
              setDelErr(null);
              try {
                await send("DELETE", "/api/account", { password: del });
                await logout().catch(() => undefined);
                nav("/");
              } catch (err) {
                setDelErr(err instanceof Error ? err.message : "Couldn't delete.");
              }
            }}>
              <p className="text-sm text-muted">Permanently deletes your account and all progress. This can’t be undone.</p>
              <Field id="del" type="password" label="Confirm with your password" autoComplete="current-password" value={del} onChange={setDel} />
              {delErr && <p role="alert" className="text-sm text-danger">{delErr}</p>}
              <Button type="submit" variant="danger" disabled={!del}>Delete my account</Button>
            </form>
          </Card>
        </div>
      </div>
    </>
  );
}
