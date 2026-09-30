import { lazy, Suspense } from "react";
import { Route, Routes, useMatch } from "react-router-dom";
import { AppShell } from "@/components/layout/AppShell";
import { RequireAuth } from "@/components/common";
import { LoadingBlock } from "@/components/ui/async";
import { Landing } from "@/pages/Landing";
import { AuthPage } from "@/pages/Auth";
import { NotFound } from "@/pages/NotFound";

// Route-level code splitting: lessons, editor pages and dashboards load on demand.
const Dashboard = lazy(() => import("@/pages/Dashboard").then((m) => ({ default: m.Dashboard })));
const Roadmap = lazy(() => import("@/pages/Roadmap").then((m) => ({ default: m.Roadmap })));
const LearnIndex = lazy(() => import("@/pages/Learn").then((m) => ({ default: m.LearnIndex })));
const ModulePage = lazy(() => import("@/pages/Learn").then((m) => ({ default: m.ModulePage })));
const LessonPage = lazy(() => import("@/pages/Lesson").then((m) => ({ default: m.LessonPage })));
const ChallengeIndex = lazy(() => import("@/pages/Challenges").then((m) => ({ default: m.ChallengeIndex })));
const ChallengePage = lazy(() => import("@/pages/Challenges").then((m) => ({ default: m.ChallengePage })));
const ProjectIndex = lazy(() => import("@/pages/Projects").then((m) => ({ default: m.ProjectIndex })));
const ProjectPage = lazy(() => import("@/pages/Projects").then((m) => ({ default: m.ProjectPage })));
const Playground = lazy(() => import("@/pages/Misc").then((m) => ({ default: m.Playground })));
const Skills = lazy(() => import("@/pages/Misc").then((m) => ({ default: m.Skills })));
const Achievements = lazy(() => import("@/pages/Misc").then((m) => ({ default: m.Achievements })));
const Glossary = lazy(() => import("@/pages/Misc").then((m) => ({ default: m.Glossary })));
const ProfilePage = lazy(() => import("@/pages/Misc").then((m) => ({ default: m.ProfilePage })));
const SettingsPage = lazy(() => import("@/pages/Misc").then((m) => ({ default: m.SettingsPage })));

const guard = (el: React.ReactNode) => <RequireAuth>{el}</RequireAuth>;

function Shell() {
  // Lesson pages use a wider canvas for the three-column layout.
  const wide = !!useMatch("/learn/:module/:lesson");
  return (
    <AppShell wide={wide}>
      <Suspense fallback={<LoadingBlock label="Loading page" />}>
        <Routes>
          <Route path="/dashboard" element={guard(<Dashboard />)} />
          <Route path="/roadmap" element={<Roadmap />} />
          <Route path="/learn" element={<LearnIndex />} />
          <Route path="/learn/:module" element={<ModulePage />} />
          <Route path="/learn/:module/:slug" element={<LessonPage />} />
          <Route path="/challenge" element={<ChallengeIndex />} />
          <Route path="/challenge/:id" element={<ChallengePage />} />
          <Route path="/projects" element={<ProjectIndex />} />
          <Route path="/projects/:id" element={<ProjectPage />} />
          <Route path="/playground" element={<Playground />} />
          <Route path="/skills" element={guard(<Skills />)} />
          <Route path="/achievements" element={guard(<Achievements />)} />
          <Route path="/glossary" element={<Glossary />} />
          <Route path="/profile" element={guard(<ProfilePage />)} />
          <Route path="/settings" element={guard(<SettingsPage />)} />
          <Route path="*" element={<NotFound />} />
        </Routes>
      </Suspense>
    </AppShell>
  );
}

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/login" element={<AuthPage mode="login" />} />
      <Route path="/register" element={<AuthPage mode="register" />} />
      <Route path="/*" element={<Shell />} />
    </Routes>
  );
}
