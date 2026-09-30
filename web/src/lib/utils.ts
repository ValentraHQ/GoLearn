import { clsx, type ClassValue } from "clsx";

export const cn = (...v: ClassValue[]) => clsx(v);

export function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  const m = Math.floor(seconds / 60);
  if (m < 60) return `${m}m`;
  const h = Math.floor(m / 60);
  return `${h}h ${m % 60}m`;
}

export const skillLevelLabel: Record<string, string> = {
  not_started: "Not started",
  learning: "Learning",
  practicing: "Practicing",
  proficient: "Proficient",
  mastered: "Mastered",
};

export const difficultyLabel: Record<string, string> = {
  beginner: "Beginner",
  intermediate: "Intermediate",
  advanced: "Advanced",
  expert: "Expert",
};

export const stateLabel: Record<string, string> = {
  completed: "Completed",
  in_progress: "In progress",
  available: "Available",
  locked: "Locked",
  planned: "Coming soon",
};

export const isMac = () => typeof navigator !== "undefined" && /Mac|iPhone|iPad/.test(navigator.platform);
export const modKey = () => (isMac() ? "⌘" : "Ctrl");

/** True when keyboard focus is somewhere typing should win over global shortcuts. */
export function isTypingTarget(el: EventTarget | null): boolean {
  if (!(el instanceof HTMLElement)) return false;
  return el.isContentEditable || ["INPUT", "TEXTAREA", "SELECT"].includes(el.tagName) || !!el.closest(".cm-editor");
}
