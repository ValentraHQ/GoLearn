import { LinkButton } from "@/components/ui/ui";

export function NotFound() {
  return (
    <div className="flex flex-col items-center gap-3 py-24 text-center">
      <p className="font-mono text-5xl font-semibold text-accent">404</p>
      <h1 className="text-xl font-semibold">Page not found</h1>
      <p className="text-sm text-muted">That page doesn't exist or has moved.</p>
      <LinkButton to="/dashboard" variant="primary">Go to dashboard</LinkButton>
    </div>
  );
}
