import type { ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { CheckCircle2, Circle, CircleDot, Lock, Clock } from "lucide-react";
import type { Difficulty, ModuleState, SkillLevel } from "@/api/types";
import { useAuth } from "@/lib/auth";
import { Badge } from "@/components/ui/ui";
import { LoadingBlock, ErrorBlock } from "@/components/ui/async";
import { difficultyLabel, skillLevelLabel, stateLabel } from "@/lib/utils";

export function RequireAuth({ children }: { children: ReactNode }) {
  const { user, loading, error, retry } = useAuth();
  const loc = useLocation();
  if (loading) return <LoadingBlock label="Checking your session" />;
  if (error && !user) return <ErrorBlock error={error} onRetry={retry} />;
  if (!user) return <Navigate to={`/login?next=${encodeURIComponent(loc.pathname + loc.search)}`} replace />;
  return <>{children}</>;
}

export function StateIcon({ state, className = "h-4 w-4" }: { state: ModuleState; className?: string }) {
  switch (state) {
    case "completed":
      return <CheckCircle2 className={`${className} text-success`} aria-hidden />;
    case "in_progress":
      return <CircleDot className={`${className} text-accent`} aria-hidden />;
    case "locked":
      return <Lock className={`${className} text-faint`} aria-hidden />;
    case "planned":
      return <Clock className={`${className} text-faint`} aria-hidden />;
    default:
      return <Circle className={`${className} text-muted`} aria-hidden />;
  }
}

export function StateBadge({ state }: { state: ModuleState }) {
  const tone = state === "completed" ? "success" : state === "in_progress" ? "accent" : "neutral";
  return (
    <Badge tone={tone}>
      <StateIcon state={state} className="h-3 w-3" />
      {stateLabel[state]}
    </Badge>
  );
}

export function DifficultyBadge({ level }: { level: Difficulty }) {
  const tone = level === "beginner" ? "success" : level === "intermediate" ? "accent" : level === "advanced" ? "warn" : "danger";
  return <Badge tone={tone}>{difficultyLabel[level]}</Badge>;
}

export function SkillLevelBadge({ level }: { level: SkillLevel }) {
  const tone = level === "mastered" ? "success" : level === "proficient" ? "accent" : level === "not_started" ? "neutral" : "warn";
  return <Badge tone={tone}>{skillLevelLabel[level]}</Badge>;
}

export const trackNames: Record<string, string> = {
  core: "Core Go",
  advanced: "Advanced Learning Track",
  devops: "DevOps & Cloud",
  "advanced-go": "Advanced Go",
  architecture: "Software Architecture",
  security: "Security",
  microservices: "Microservices",
};
