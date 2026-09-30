export type Difficulty = "beginner" | "intermediate" | "advanced" | "expert";
export type ModuleState = "completed" | "in_progress" | "available" | "locked" | "planned";
export type SkillLevel = "not_started" | "learning" | "practicing" | "proficient" | "mastered";

export interface Profile {
  displayName: string;
  bio: string;
  learningPath: "beginner" | "pro";
  theme: "system" | "light" | "dark";
  dailyLessons: number;
  dailyChallenges: number;
  dailyQuizzes: number;
  gamification: boolean;
  timezone: string;
}

export interface User {
  id: string;
  email: string;
  role: string;
  profile: Profile;
}

export interface RunnerStatus {
  available: boolean;
  backend: string;
  reason?: string;
}

export interface AppConfig {
  auth: { local: boolean; oidc: boolean };
  runner: RunnerStatus;
  limits: { maxCodeBytes: number };
  content: { modules: number; lessons: number; challenges: number; projects: number; glossary: number };
}

export interface LessonSummary {
  id: string;
  slug: string;
  title: string;
  status: "published" | "planned" | "draft" | "review" | "archived";
  minutes: number;
  skill: string;
  hasExercise: boolean;
  hasQuiz: boolean;
  completed: boolean;
}

export interface ModuleProgress {
  moduleId: string;
  state: ModuleState;
  totalLessons: number;
  completedLessons: number;
  percent: number;
  blockedBy?: string[];
  inPath: boolean;
}

export interface ModuleSummary {
  id: string;
  number: string;
  title: string;
  summary: string;
  track: string;
  paths: string[];
  skill: string;
  requires: string[];
  project?: string;
  lessons: LessonSummary[];
  progress?: ModuleProgress;
}

export interface Example {
  lang: string;
  code: string;
  runnable: boolean;
  caption?: string;
}

export type QuestionType = "mcq" | "tf" | "output" | "debug" | "multi" | "short";

export interface QuizQuestion {
  id: string;
  type: QuestionType;
  prompt: string;
  code?: string;
  options?: string[];
}

export interface Lesson {
  id: string;
  slug: string;
  moduleId: string;
  title: string;
  minutes: number;
  skill: string;
  objectives: string[];
  concept: string;
  examples: Example[];
  exercise: { prompt: string; starter: string; expected: string } | null;
  quiz: QuizQuestion[] | null;
  takeaways: string[];
}

export interface LessonNav {
  id: string;
  title: string;
  module: string;
  slug: string;
}

export interface ChallengeSummary {
  id: string;
  title: string;
  difficulty: Difficulty;
  moduleId: string;
  lesson?: string;
  skill: string;
  attempts: number;
  passed: boolean;
}

export interface LessonDetail {
  lesson: Lesson;
  module: { id: string; number: string; title: string };
  curriculum: LessonSummary[];
  prev: LessonNav | null;
  next: LessonNav | null;
  progress: { completed: boolean; exercisePassed: boolean; quizPassed: boolean } | null;
  challenges: ChallengeSummary[] | null;
  glossary: { id: string; term: string }[] | null;
  requirements: { quiz: boolean; exercise: boolean };
}

export interface ChallengeDetail {
  challenge: {
    id: string;
    title: string;
    difficulty: Difficulty;
    moduleId: string;
    skill: string;
    problem: string;
    input: string;
    expectedOutput: string;
    constraints: string[] | null;
    starter: string;
    hints: string[] | null;
    testNames: string[];
    explanation: string;
  };
  status: ChallengeSummary;
  lastCode: string;
  lesson: LessonNav | null;
  solutionUnlocked: boolean;
  solutionAfter: number;
}

export interface ProjectSummary {
  id: string;
  title: string;
  level: "beginner" | "intermediate" | "advanced" | "capstone";
  moduleId?: string;
  skills: string[];
  hours: number;
  summary: string;
  taskCount: number;
  tasksDone: number;
}

export interface ProjectDetail {
  project: {
    id: string;
    title: string;
    level: ProjectSummary["level"];
    moduleId?: string;
    skills: string[];
    hours: number;
    summary: string;
    description: string;
    requirements: string[] | null;
    architecture?: string;
    tasks: { id: string; title: string; notes?: string }[];
    stretch?: string[] | null;
  };
  done: Record<string, boolean>;
  module: LessonNav | null;
}

export interface GlossaryTerm {
  id: string;
  term: string;
  simple: string;
  technical: string;
  code?: string;
  relatedLessons: { id: string; title: string; module: string; slug: string }[];
}

export interface SearchResult {
  kind: "lesson" | "module" | "challenge" | "project" | "example" | "glossary";
  id: string;
  title: string;
  group: string;
  subtitle?: string;
  snippet?: string;
  path: string;
}

export interface Achievement {
  id: string;
  title: string;
  description: string;
  category: string;
  unlocked: boolean;
  unlockedAt?: string;
}

export interface SkillProgress {
  skillId: string;
  name: string;
  level: SkillLevel;
  done: number;
  total: number;
  percent: number;
  modules?: string[];
}

export interface Summary {
  lessonsCompleted: number;
  lessonsTotal: number;
  modulesCompleted: number;
  challengesCompleted: number;
  challengesTotal: number;
  projectsCompleted: number;
  projectsTotal: number;
  quizzesPassed: number;
  quizAveragePercent: number;
  currentStreak: number;
  longestStreak: number;
  learningSeconds: number;
  overallPercent: number;
  xp: number;
  level: number;
  xpIntoLevel: number;
  xpForNextLevel: number;
  modules: ModuleProgress[];
  skills: SkillProgress[];
  daily: { lessons: Count; challenges: Count; quizzes: Count };
  achievements: Achievement[];
  continue?: { moduleId: string; moduleTitle: string; lessonId: string; lessonTitle: string; started: boolean };
}

export interface Count {
  done: number;
  goal: number;
}

export interface RunResult {
  stdout: string;
  stderr: string;
  exitCode: number;
  timedOut: boolean;
  truncated: boolean;
  buildError: boolean;
  durationMs: number;
}

export interface Unlocked {
  id: string;
  title: string;
}

export interface ExerciseResult {
  run: RunResult;
  passed: boolean;
  expected: string;
  achievements: Unlocked[];
}

export interface QuizResult {
  score: number;
  total: number;
  passed: boolean;
  results: { id: string; correct: boolean; explanation: string }[];
  achievements: Unlocked[];
}

export interface ChallengeResult {
  submissionId: string;
  passed: boolean;
  testsPassed: number;
  testsTotal: number;
  tests: { name: string; passed: boolean; message?: string }[];
  buildOutput: string;
  timedOut: boolean;
  truncated: boolean;
  durationMs: number;
  achievements: Unlocked[];
}

export interface Paged<T> {
  data: T[];
  meta: { page: number; pageSize: number; total: number };
}
