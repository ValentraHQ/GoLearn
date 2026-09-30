import { Link } from "react-router-dom";
import { CheckCircle2, Circle, Clock, Code2, Flame, FolderKanban, GraduationCap, Trophy } from "lucide-react";
import { useDashboard } from "@/api/queries";
import type { Summary } from "@/api/types";
import { Async } from "@/components/ui/async";
import { Card, CardHeader, LinkButton, PageHeader, ProgressBar } from "@/components/ui/ui";
import { useAuth } from "@/lib/auth";
import { formatDuration } from "@/lib/utils";

export function Dashboard() {
  const q = useDashboard();
  const { user } = useAuth();
  return (
    <>
      <PageHeader title={`Welcome back${user ? `, ${user.profile.displayName}` : ""}`} description="Pick up where you left off." />
      <Async query={q}>{({ summary, profile }) => <DashboardBody s={summary} gamification={profile.gamification} />}</Async>
    </>
  );
}

function Stat({ icon: Icon, label, value, sub }: { icon: typeof Flame; label: string; value: string | number; sub?: string }) {
  return (
    <Card className="p-4">
      <div className="flex items-center gap-2 text-sm text-muted">
        <Icon className="h-4 w-4" aria-hidden /> {label}
      </div>
      <p className="mt-2 font-mono text-2xl font-semibold">{value}</p>
      {sub && <p className="text-xs text-faint">{sub}</p>}
    </Card>
  );
}

function DashboardBody({ s, gamification }: { s: Summary; gamification: boolean }) {
  const plural = (n: number, one: string, many: string) => (n === 1 ? one : many);
  const goals = [
    { label: plural(s.daily.lessons.goal, "lesson", "lessons"), ...s.daily.lessons },
    { label: plural(s.daily.challenges.goal, "challenge", "challenges"), ...s.daily.challenges },
    { label: plural(s.daily.quizzes.goal, "quiz", "quizzes"), ...s.daily.quizzes },
  ].filter((g) => g.goal > 0);
  const skills = s.skills.filter((k) => k.total > 0).slice(0, 8);
  const c = s.continue;

  return (
    <div className="grid gap-5 lg:grid-cols-3">
      <div className="space-y-5 lg:col-span-2">
        <Card>
          <CardHeader title="Continue learning" />
          <div className="p-5">
            {c ? (
              <>
                <p className="text-xs uppercase tracking-wide text-muted">{c.moduleTitle}</p>
                <p className="mt-1 text-xl font-semibold">{c.lessonTitle}</p>
                <p className="mt-1 text-sm text-muted">{c.started ? "You've started this lesson." : "Your next lesson."}</p>
                <LinkButton to={`/learn/${c.lessonId}`} variant="primary" className="mt-4">
                  {c.started ? "Continue" : "Start lesson"}
                </LinkButton>
              </>
            ) : (
              <p className="text-sm text-muted">You have completed every published lesson. Try a project or challenge next.</p>
            )}
          </div>
        </Card>

        <Card>
          <CardHeader title="Your progress" action={<span className="font-mono text-sm text-muted">{s.overallPercent}%</span>} />
          <div className="space-y-4 p-5">
            <ProgressBar value={s.overallPercent} label="Overall progress" />
            <p className="text-sm text-muted">
              {s.lessonsCompleted} of {s.lessonsTotal} lessons · {s.modulesCompleted} modules complete
            </p>
            {skills.length > 0 && (
              <ul className="space-y-3 pt-1">
                {skills.map((k) => (
                  <li key={k.skillId} className="grid grid-cols-[8rem_1fr_3rem] items-center gap-3 text-sm sm:grid-cols-[10rem_1fr_3rem]">
                    <span className="truncate">{k.name}</span>
                    <ProgressBar value={k.percent} label={`${k.name} progress`} tone={k.percent === 100 ? "success" : "accent"} />
                    <span className="text-right font-mono text-xs text-muted">{k.percent}%</span>
                  </li>
                ))}
              </ul>
            )}
            <Link to="/skills" className="inline-block text-sm text-accent underline underline-offset-2">
              View the full skill matrix
            </Link>
          </div>
        </Card>
      </div>

      <div className="space-y-5">
        <Card>
          <CardHeader title="Today's goal" />
          <div className="p-5">
            {goals.length === 0 ? (
              <p className="text-sm text-muted">
                No daily goal set. <Link className="text-accent underline" to="/settings">Set one in Settings</Link>.
              </p>
            ) : (
              <ul className="space-y-2.5">
                {goals.map((g) => {
                  const done = g.done >= g.goal;
                  return (
                    <li key={g.label} className="flex items-center gap-2.5 text-sm">
                      {done ? <CheckCircle2 className="h-4 w-4 text-success" aria-label="done" /> : <Circle className="h-4 w-4 text-faint" aria-label="not done" />}
                      <span className={done ? "text-muted line-through" : ""}>
                        {g.goal} {g.label}
                      </span>
                      <span className="ml-auto font-mono text-xs text-faint">
                        {Math.min(g.done, g.goal)}/{g.goal}
                      </span>
                    </li>
                  );
                })}
              </ul>
            )}
          </div>
        </Card>

        <div className="grid grid-cols-2 gap-3">
          <Stat icon={GraduationCap} label="Lessons" value={s.lessonsCompleted} sub={`of ${s.lessonsTotal}`} />
          <Stat icon={Code2} label="Challenges" value={s.challengesCompleted} sub={`of ${s.challengesTotal}`} />
          <Stat icon={FolderKanban} label="Projects" value={s.projectsCompleted} sub={`of ${s.projectsTotal}`} />
          <Stat icon={Flame} label="Streak" value={`${s.currentStreak}d`} sub={`best ${s.longestStreak}d`} />
          <Stat icon={Clock} label="Learning time" value={formatDuration(s.learningSeconds)} />
          <Stat icon={CheckCircle2} label="Quiz average" value={s.quizzesPassed ? `${s.quizAveragePercent}%` : "—"} sub={`${s.quizzesPassed} passed`} />
        </div>

        {gamification && (
          <Card className="p-4">
            <div className="flex items-center justify-between text-sm">
              <span className="flex items-center gap-2 text-muted">
                <Trophy className="h-4 w-4" aria-hidden /> Level {s.level}
              </span>
              <span className="font-mono text-xs text-faint">{s.xp} XP</span>
            </div>
            <ProgressBar className="mt-3" value={(s.xpIntoLevel / Math.max(s.xpForNextLevel, 1)) * 100} label="Progress to next level" />
            <p className="mt-2 text-xs text-faint">{s.xpForNextLevel - s.xpIntoLevel} XP to level {s.level + 1}</p>
          </Card>
        )}
        <p className="px-1 text-xs text-faint">Days are counted in your profile timezone.</p>
      </div>
    </div>
  );
}

