import { useQuery } from "@tanstack/react-query";
import { get, getEnvelope, qs } from "./client";
import type {
  AppConfig,
  ChallengeDetail,
  ChallengeSummary,
  GlossaryTerm,
  LessonDetail,
  ModuleSummary,
  Paged,
  ProjectDetail,
  ProjectSummary,
  SearchResult,
  SkillProgress,
  Summary,
  Achievement,
  Profile,
} from "./types";

const STATIC = 5 * 60_000;

export const useConfig = () =>
  useQuery({ queryKey: ["config"], queryFn: ({ signal }) => get<AppConfig>("/api/config", signal), staleTime: 30_000 });

export const useModules = (authed: boolean) =>
  useQuery({ queryKey: ["modules", authed], queryFn: ({ signal }) => get<ModuleSummary[]>("/api/modules", signal), staleTime: 30_000 });

export const useModule = (id: string) =>
  useQuery({
    queryKey: ["module", id],
    queryFn: ({ signal }) =>
      get<{ module: ModuleSummary; challenges: ChallengeSummary[] | null; project: ProjectSummary | null }>(`/api/modules/${id}`, signal),
  });

export const useLesson = (module: string, slug: string) =>
  useQuery({ queryKey: ["lesson", module, slug], queryFn: ({ signal }) => get<LessonDetail>(`/api/lessons/${module}/${slug}`, signal) });

export interface ChallengeFilters {
  difficulty?: string;
  module?: string;
  skill?: string;
  status?: string;
  q?: string;
  page?: number;
}

export const useChallenges = (f: ChallengeFilters) =>
  useQuery({
    queryKey: ["challenges", f],
    queryFn: ({ signal }) => getEnvelope<Paged<ChallengeSummary>>(`/api/challenges${qs({ ...f, pageSize: 20 })}`, signal),
    placeholderData: (prev) => prev,
  });

export const useChallenge = (id: string) =>
  useQuery({ queryKey: ["challenge", id], queryFn: ({ signal }) => get<ChallengeDetail>(`/api/challenges/${id}`, signal) });

export const useProjects = () =>
  useQuery({ queryKey: ["projects"], queryFn: ({ signal }) => get<ProjectSummary[]>("/api/projects", signal) });

export const useProject = (id: string) =>
  useQuery({ queryKey: ["project", id], queryFn: ({ signal }) => get<ProjectDetail>(`/api/projects/${id}`, signal) });

export const useGlossary = () =>
  useQuery({ queryKey: ["glossary"], queryFn: ({ signal }) => get<GlossaryTerm[]>("/api/glossary", signal), staleTime: STATIC });

export const useSearch = (q: string) =>
  useQuery({
    queryKey: ["search", q],
    enabled: q.trim().length > 0,
    queryFn: ({ signal }) => get<{ query: string; results: SearchResult[] }>(`/api/search${qs({ q, limit: 30 })}`, signal),
    staleTime: STATIC,
  });

export const useDashboard = (enabled = true) =>
  useQuery({
    queryKey: ["dashboard"],
    enabled,
    queryFn: ({ signal }) => get<{ summary: Summary; profile: Profile }>("/api/dashboard", signal),
  });

export const useSkills = () =>
  useQuery({ queryKey: ["skills"], queryFn: ({ signal }) => get<SkillProgress[]>("/api/skills", signal) });

export const useAchievements = () =>
  useQuery({
    queryKey: ["achievements"],
    queryFn: ({ signal }) =>
      get<{ achievements: Achievement[]; xp: number; level: number; xpIntoLevel: number; xpForNextLevel: number; gamification: boolean }>(
        "/api/achievements",
        signal,
      ),
  });
