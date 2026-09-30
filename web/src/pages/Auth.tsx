import { useState, type FormEvent } from "react";
import { Link, Navigate, useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "@/lib/auth";
import { Button, Card } from "@/components/ui/ui";
import { Logo } from "@/components/layout/AppShell";
import { useConfig } from "@/api/queries";

function safeNext(n: string | null): string {
  return n && n.startsWith("/") && !n.startsWith("//") ? n : "/dashboard";
}

export function AuthPage({ mode }: { mode: "login" | "register" }) {
  const { user, login, register } = useAuth();
  const nav = useNavigate();
  const [params] = useSearchParams();
  const next = safeNext(params.get("next"));
  const cfg = useConfig();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const isRegister = mode === "register";

  if (user) return <Navigate to={next} replace />;

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      if (isRegister) await register(email, password, name);
      else await login(email, password);
      nav(next, { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong.");
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex min-h-screen flex-col items-center justify-center px-4 py-10">
      <Logo className="mb-6" />
      <Card className="w-full max-w-sm p-6">
        <h1 className="text-xl font-semibold tracking-tight">{isRegister ? "Create your account" : "Welcome back"}</h1>
        <p className="mt-1 text-sm text-muted">{isRegister ? "Start learning Go in minutes." : "Sign in to continue learning."}</p>
        <form onSubmit={submit} className="mt-5 space-y-4" noValidate>
          {isRegister && (
            <Field id="name" label="Display name" autoComplete="name" value={name} onChange={setName} placeholder="Ada Lovelace" />
          )}
          <Field id="email" label="Email" type="email" autoComplete="email" value={email} onChange={setEmail} required />
          <Field
            id="password"
            label="Password"
            type="password"
            autoComplete={isRegister ? "new-password" : "current-password"}
            value={password}
            onChange={setPassword}
            required
            hint={isRegister ? "At least 10 characters." : undefined}
          />
          {error && (
            <p role="alert" className="rounded-lg border border-danger/40 bg-danger-soft px-3 py-2 text-sm text-danger">
              {error}
            </p>
          )}
          <Button type="submit" variant="primary" className="w-full" disabled={busy}>
            {busy ? "Please wait…" : isRegister ? "Create account" : "Sign in"}
          </Button>
        </form>
        {cfg.data && !cfg.data.auth.oidc && (
          <p className="mt-4 text-xs text-faint">Single sign-on (OIDC) isn’t configured on this server.</p>
        )}
        <p className="mt-4 text-sm text-muted">
          {isRegister ? "Already have an account? " : "New to GoLearn? "}
          <Link className="text-accent underline underline-offset-2" to={`${isRegister ? "/login" : "/register"}${next !== "/dashboard" ? `?next=${encodeURIComponent(next)}` : ""}`}>
            {isRegister ? "Sign in" : "Create an account"}
          </Link>
        </p>
      </Card>
    </div>
  );
}

export function Field({
  id,
  label,
  value,
  onChange,
  type = "text",
  hint,
  ...rest
}: {
  id: string;
  label: string;
  value: string;
  onChange: (v: string) => void;
  type?: string;
  hint?: string;
  required?: boolean;
  autoComplete?: string;
  placeholder?: string;
}) {
  return (
    <div>
      <label htmlFor={id} className="mb-1 block text-sm font-medium">
        {label}
      </label>
      <input
        id={id}
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        aria-describedby={hint ? `${id}-hint` : undefined}
        className="h-10 w-full rounded-lg border border-border bg-bg px-3 text-sm outline-none focus:border-accent focus:ring-2 focus:ring-accent/30"
        {...rest}
      />
      {hint && (
        <p id={`${id}-hint`} className="mt-1 text-xs text-faint">
          {hint}
        </p>
      )}
    </div>
  );
}
