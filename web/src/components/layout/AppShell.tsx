import { useEffect, useState, type ReactNode } from "react";
import { Link, NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";
import {
  Award,
  BookOpen,
  BookText,
  Code2,
  FolderKanban,
  LayoutDashboard,
  LogOut,
  Map,
  Menu,
  Moon,
  Search,
  Settings,
  Sun,
  SquareTerminal,
  Target,
  Trophy,
  User as UserIcon,
  Flame,
} from "lucide-react";
import { useAuth } from "@/lib/auth";
import { useTheme } from "@/lib/theme";
import { cn, modKey } from "@/lib/utils";
import { Button, Kbd, LinkButton } from "@/components/ui/ui";
import { Sheet } from "@/components/ui/sheet";
import { SearchPalette } from "@/components/SearchPalette";
import { useDashboard } from "@/api/queries";

const NAV = [
  { to: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
  { to: "/roadmap", label: "Roadmap", icon: Map },
  { to: "/learn", label: "Learn", icon: BookOpen },
  { to: "/challenge", label: "Challenges", icon: Code2 },
  { to: "/projects", label: "Projects", icon: FolderKanban },
  { to: "/playground", label: "Playground", icon: SquareTerminal },
  { to: "/skills", label: "Skills", icon: Target },
  { to: "/achievements", label: "Achievements", icon: Trophy },
  { to: "/glossary", label: "Glossary", icon: BookText },
  { to: "/settings", label: "Settings", icon: Settings },
] as const;

export function Logo({ className }: { className?: string }) {
  return (
    <Link to="/" className={cn("flex items-center gap-2 font-semibold tracking-tight", className)} aria-label="GoLearn home">
      <svg viewBox="0 0 32 32" className="h-7 w-7" aria-hidden>
        <rect width="32" height="32" rx="7" className="fill-fg" />
        <path d="M8 11.5 4.5 16 8 20.5M24 11.5 27.5 16 24 20.5" fill="none" stroke="var(--accent)" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" />
        <path d="M13.5 22.5 18.5 9.5" stroke="var(--accent-strong)" strokeWidth="2.4" strokeLinecap="round" />
      </svg>
      <span className="text-lg">
        Go<span className="text-accent">Learn</span>
      </span>
    </Link>
  );
}

function NavList({ onNavigate }: { onNavigate?: () => void }) {
  return (
    <nav aria-label="Primary" className="flex flex-col gap-0.5 p-3">
      {NAV.map(({ to, label, icon: Icon }) => (
        <NavLink
          key={to}
          to={to}
          onClick={onNavigate}
          className={({ isActive }) =>
            cn(
              "flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors",
              isActive ? "bg-accent-soft font-medium text-accent" : "text-muted hover:bg-surface-2 hover:text-fg",
            )
          }
        >
          <Icon className="h-4 w-4" aria-hidden />
          {label}
        </NavLink>
      ))}
    </nav>
  );
}

function ThemeToggle() {
  const { resolved, setPref } = useTheme();
  const dark = resolved === "dark";
  return (
    <Button variant="ghost" size="sm" aria-label={dark ? "Switch to light theme" : "Switch to dark theme"} onClick={() => setPref(dark ? "light" : "dark")}>
      {dark ? <Sun className="h-4 w-4" aria-hidden /> : <Moon className="h-4 w-4" aria-hidden />}
    </Button>
  );
}

function StreakChip() {
  const { user } = useAuth();
  const q = useDashboard();
  if (!user || !q.data) return null;
  const s = q.data.summary;
  return (
    <Link to="/dashboard" className="hidden items-center gap-1.5 rounded-lg px-2 py-1 text-sm text-muted hover:bg-surface-2 sm:flex" aria-label={`${s.currentStreak} day streak`}>
      <Flame className={cn("h-4 w-4", s.currentStreak > 0 ? "text-warn" : "text-faint")} aria-hidden />
      <span className="font-mono">{s.currentStreak}</span>
    </Link>
  );
}

function UserMenu() {
  const { user, logout } = useAuth();
  const nav = useNavigate();
  if (!user)
    return (
      <div className="flex items-center gap-2">
        <LinkButton to="/login" variant="ghost" size="sm">
          Sign in
        </LinkButton>
        <LinkButton to="/register" variant="primary" size="sm">
          Get started
        </LinkButton>
      </div>
    );
  return (
    <div className="flex items-center gap-1">
      <Link to="/profile" className="flex items-center gap-2 rounded-lg px-2 py-1.5 text-sm hover:bg-surface-2" aria-label="Your profile">
        <span className="flex h-6 w-6 items-center justify-center rounded-full bg-accent-soft text-xs font-semibold text-accent" aria-hidden>
          {user.profile.displayName.slice(0, 1).toUpperCase() || <UserIcon className="h-3 w-3" />}
        </span>
        <span className="hidden max-w-32 truncate md:inline">{user.profile.displayName}</span>
      </Link>
      <Button
        variant="ghost"
        size="sm"
        aria-label="Sign out"
        onClick={async () => {
          await logout();
          nav("/");
        }}
      >
        <LogOut className="h-4 w-4" aria-hidden />
      </Button>
    </div>
  );
}

export function AppShell({ children, wide }: { children?: ReactNode; wide?: boolean }) {
  const [drawer, setDrawer] = useState(false);
  const [search, setSearch] = useState(false);
  const loc = useLocation();

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === "k") {
        e.preventDefault();
        setSearch((s) => !s);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, []);

  // Move focus to the page heading region on navigation for screen-reader users.
  useEffect(() => {
    document.getElementById("main")?.focus({ preventScroll: true });
    window.scrollTo(0, 0);
  }, [loc.pathname]);

  return (
    <div className="min-h-screen">
      <a href="#main" className="sr-only focus:not-sr-only focus:fixed focus:left-3 focus:top-3 focus:z-[70] focus:rounded-lg focus:bg-accent focus:px-3 focus:py-2 focus:text-accent-fg">
        Skip to content
      </a>
      <aside className="fixed inset-y-0 left-0 z-30 hidden w-56 flex-col border-r border-border bg-surface lg:flex">
        <div className="flex h-14 items-center border-b border-border px-4">
          <Logo />
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto">
          <NavList />
        </div>
        <p className="border-t border-border px-4 py-3 text-xs text-faint">Learn Go. Build Real Software.</p>
      </aside>

      <Sheet open={drawer} onOpenChange={setDrawer} title="Navigation">
        <NavList onNavigate={() => setDrawer(false)} />
      </Sheet>

      <div className="lg:pl-56">
        <header className="sticky top-0 z-20 flex h-14 items-center gap-2 border-b border-border bg-bg/85 px-3 backdrop-blur sm:px-5">
          <Button variant="ghost" size="sm" className="lg:hidden" aria-label="Open navigation" onClick={() => setDrawer(true)}>
            <Menu className="h-4 w-4" aria-hidden />
          </Button>
          <Logo className="lg:hidden" />
          <button
            onClick={() => setSearch(true)}
            className="ml-auto flex h-9 w-full max-w-md items-center gap-2 rounded-lg border border-border bg-surface px-3 text-sm text-muted hover:border-border-strong lg:ml-0"
            aria-label="Search lessons, challenges, projects and glossary"
            aria-keyshortcuts="Control+K Meta+K"
          >
            <Search className="h-4 w-4" aria-hidden />
            <span className="hidden flex-1 text-left sm:inline">Search…</span>
            <span className="ml-auto hidden items-center gap-1 sm:flex">
              <Kbd>{modKey()}</Kbd>
              <Kbd>K</Kbd>
            </span>
          </button>
          <div className="ml-auto flex items-center gap-1 lg:ml-auto">
            <StreakChip />
            <ThemeToggle />
            <UserMenu />
          </div>
        </header>
        <main id="main" tabIndex={-1} className={cn("mx-auto w-full px-4 py-6 outline-none sm:px-6 lg:py-8", wide ? "max-w-[96rem]" : "max-w-6xl")}>
          {children ?? <Outlet />}
        </main>
      </div>
      <SearchPalette open={search} onOpenChange={setSearch} />
    </div>
  );
}

/** Icon shown for achievements/skills in a few places. */
export const AwardIcon = Award;
