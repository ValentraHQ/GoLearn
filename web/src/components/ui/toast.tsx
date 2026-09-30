import { createContext, useCallback, useContext, useMemo, useState, type ReactNode } from "react";
import { Trophy, X } from "lucide-react";

interface Toast {
  id: number;
  title: string;
  body?: string;
}
interface ToastCtx {
  push: (title: string, body?: string) => void;
  achievement: (titles: string[]) => void;
}
const Ctx = createContext<ToastCtx | null>(null);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const remove = useCallback((id: number) => setToasts((t) => t.filter((x) => x.id !== id)), []);
  const push = useCallback(
    (title: string, body?: string) => {
      const id = Date.now() + Math.random();
      setToasts((t) => [...t.slice(-3), { id, title, body }]);
      setTimeout(() => remove(id), 6000);
    },
    [remove],
  );
  const value = useMemo<ToastCtx>(
    () => ({ push, achievement: (titles) => titles.forEach((t) => push("Achievement unlocked", t)) }),
    [push],
  );
  return (
    <Ctx.Provider value={value}>
      {children}
      <div aria-live="polite" role="status" className="pointer-events-none fixed bottom-4 right-4 z-[60] flex w-80 max-w-[calc(100vw-2rem)] flex-col gap-2">
        {toasts.map((t) => (
          <div key={t.id} className="pointer-events-auto flex items-start gap-3 rounded-xl border border-border bg-surface p-3 shadow-card">
            <Trophy className="mt-0.5 h-4 w-4 shrink-0 text-warn" aria-hidden />
            <div className="min-w-0 flex-1">
              <p className="text-sm font-semibold">{t.title}</p>
              {t.body && <p className="text-sm text-muted">{t.body}</p>}
            </div>
            <button aria-label="Dismiss notification" onClick={() => remove(t.id)} className="text-faint hover:text-fg">
              <X className="h-4 w-4" />
            </button>
          </div>
        ))}
      </div>
    </Ctx.Provider>
  );
}

export function useToast(): ToastCtx {
  const c = useContext(Ctx);
  if (!c) throw new Error("useToast outside ToastProvider");
  return c;
}
