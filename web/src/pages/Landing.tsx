import { Link } from "react-router-dom";
import { ArrowRight, BookOpenCheck, Boxes, Braces, Cloud, GitBranch, ShieldCheck, TerminalSquare } from "lucide-react";
import { useConfig } from "@/api/queries";
import { useAuth } from "@/lib/auth";
import { useTheme } from "@/lib/theme";
import { Logo } from "@/components/layout/AppShell";
import { Button, LinkButton, Card } from "@/components/ui/ui";
import { Moon, Sun } from "lucide-react";

const LOOP = ["Learn", "Understand", "Code", "Practice", "Build", "Review", "Master"];

const features = [
  { icon: TerminalSquare, title: "Write and run real Go", body: "Every lesson has an editor wired to an isolated sandbox. Compile, run, and get real compiler errors." },
  { icon: BookOpenCheck, title: "Short, focused lessons", body: "Objective, concept, example, exercise, knowledge check, takeaways. About 20% reading, 80% practice." },
  { icon: Braces, title: "Challenges with real tests", body: "Solve problems against Go’s own testing package, not string matching. Hints and solutions unlock as you try." },
  { icon: Cloud, title: "DevOps & cloud track", body: "CLIs, Docker, Kubernetes client-go, controllers, observability and production hardening." },
  { icon: GitBranch, title: "Projects that ship", body: "From a CLI calculator to a Kubernetes controller and a microservice capstone." },
  { icon: ShieldCheck, title: "Idiomatic and current", body: "Modern Go: generics, slog, errors.Is/As, context, table-driven tests. No cargo-cult patterns." },
];

const sample = `package main

import "fmt"

func largest(xs []int) int {
	m := xs[0]
	for _, x := range xs[1:] {
		m = max(m, x)
	}
	return m
}

func main() {
	fmt.Println(largest([]int{3, 9, 4})) // 9
}`;

export function Landing() {
  const { user } = useAuth();
  const cfg = useConfig();
  const { resolved, setPref } = useTheme();
  const c = cfg.data?.content;
  return (
    <div className="min-h-screen">
      <header className="mx-auto flex h-16 max-w-6xl items-center justify-between px-4 sm:px-6">
        <Logo />
        <nav aria-label="Site" className="flex items-center gap-1 sm:gap-2">
          <Link to="/learn" className="hidden px-3 text-sm text-muted hover:text-fg sm:block">
            Curriculum
          </Link>
          <Link to="/roadmap" className="hidden px-3 text-sm text-muted hover:text-fg sm:block">
            Roadmap
          </Link>
          <Button variant="ghost" size="sm" aria-label={resolved === "dark" ? "Switch to light theme" : "Switch to dark theme"} onClick={() => setPref(resolved === "dark" ? "light" : "dark")}>
            {resolved === "dark" ? <Sun className="h-4 w-4" aria-hidden /> : <Moon className="h-4 w-4" aria-hidden />}
          </Button>
          {user ? (
            <LinkButton to="/dashboard" variant="primary" size="sm">
              Dashboard
            </LinkButton>
          ) : (
            <>
              <LinkButton to="/login" variant="ghost" size="sm">
                Sign in
              </LinkButton>
              <LinkButton to="/register" variant="primary" size="sm">
                Get started
              </LinkButton>
            </>
          )}
        </nav>
      </header>

      <main>
        <section className="mx-auto grid max-w-6xl items-center gap-10 px-4 pb-16 pt-10 sm:px-6 lg:grid-cols-2 lg:pt-16">
          <div>
            <p className="mb-4 inline-flex items-center gap-2 rounded-full border border-border bg-surface px-3 py-1 font-mono text-xs text-muted">
              <span className="h-1.5 w-1.5 rounded-full bg-accent" aria-hidden /> go 1.24+ · beginner → production
            </p>
            <h1 className="text-4xl font-semibold leading-tight tracking-tight sm:text-5xl">
              Learn Go.
              <br />
              <span className="text-accent">Build Real Software.</span>
            </h1>
            <p className="mt-5 max-w-lg text-lg text-muted">
              A hands-on path from your first <code className="font-mono text-fg">fmt.Println</code> to concurrent services, REST APIs, and Kubernetes controllers.
            </p>
            <div className="mt-7 flex flex-wrap gap-3">
              <LinkButton to={user ? "/dashboard" : "/register"} variant="primary" size="lg">
                {user ? "Continue learning" : "Start learning free"} <ArrowRight className="h-4 w-4" aria-hidden />
              </LinkButton>
              <LinkButton to="/roadmap" size="lg">
                View the roadmap
              </LinkButton>
            </div>
            {c && (
              <dl className="mt-8 grid max-w-md grid-cols-4 gap-4 text-sm">
                {[
                  [c.modules, "modules"],
                  [c.lessons, "lessons"],
                  [c.challenges, "challenges"],
                  [c.projects, "projects"],
                ].map(([n, l]) => (
                  <div key={String(l)}>
                    <dt className="text-muted">{l}</dt>
                    <dd className="font-mono text-xl font-semibold">{n}</dd>
                  </div>
                ))}
              </dl>
            )}
          </div>

          <Card className="overflow-hidden" aria-label="Example of the GoLearn editor">
            <div className="flex items-center gap-2 border-b border-border px-3 py-2">
              <span className="flex gap-1.5" aria-hidden>
                <i className="h-2.5 w-2.5 rounded-full bg-danger/70" />
                <i className="h-2.5 w-2.5 rounded-full bg-warn/70" />
                <i className="h-2.5 w-2.5 rounded-full bg-success/70" />
              </span>
              <span className="font-mono text-xs text-muted">main.go</span>
            </div>
            <pre className="overflow-x-auto bg-code-bg p-4 font-mono text-[0.8rem] leading-relaxed" tabIndex={0}>
              {sample}
            </pre>
            <div className="border-t border-border bg-bg-subtle p-3 font-mono text-xs">
              <p className="text-faint">$ go run .</p>
              <p className="text-success">9</p>
            </div>
          </Card>
        </section>

        <section aria-labelledby="loop" className="border-y border-border bg-surface">
          <div className="mx-auto max-w-6xl px-4 py-10 sm:px-6">
            <h2 id="loop" className="text-sm font-semibold uppercase tracking-wide text-muted">
              How you learn
            </h2>
            <ol className="mt-4 flex flex-wrap items-center gap-2 font-mono text-sm">
              {LOOP.map((s, i) => (
                <li key={s} className="flex items-center gap-2">
                  <span className="rounded-lg border border-border bg-bg px-3 py-1.5">{s}</span>
                  {i < LOOP.length - 1 && <ArrowRight className="h-3.5 w-3.5 text-faint" aria-hidden />}
                </li>
              ))}
            </ol>
          </div>
        </section>

        <section className="mx-auto max-w-6xl px-4 py-14 sm:px-6" aria-labelledby="features">
          <h2 id="features" className="sr-only">
            Features
          </h2>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {features.map(({ icon: Icon, title, body }) => (
              <Card key={title} className="p-5">
                <Icon className="h-5 w-5 text-accent" aria-hidden />
                <h3 className="mt-3 font-semibold">{title}</h3>
                <p className="mt-1 text-sm text-muted">{body}</p>
              </Card>
            ))}
          </div>
        </section>

        <section className="mx-auto max-w-6xl px-4 pb-20 sm:px-6">
          <Card className="flex flex-col items-start justify-between gap-4 p-6 sm:flex-row sm:items-center">
            <div className="flex items-center gap-3">
              <Boxes className="h-6 w-6 text-accent" aria-hidden />
              <div>
                <p className="font-semibold">Two paths, one curriculum</p>
                <p className="text-sm text-muted">Beginner path starts at installing Go. The Professional / DevOps path skips the basics and goes straight to services and infrastructure.</p>
              </div>
            </div>
            <LinkButton to="/learn" variant="secondary">
              Browse modules
            </LinkButton>
          </Card>
        </section>
      </main>
      <footer className="border-t border-border py-6 text-center text-xs text-faint">GoLearn — Learn Go. Build Real Software.</footer>
    </div>
  );
}
