import { createContext, useCallback, useContext, useEffect, useMemo, type ReactNode } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { get, send } from "@/api/client";
import type { Profile, User } from "@/api/types";
import { useTheme } from "@/lib/theme";

interface AuthCtx {
  user: User | null;
  loading: boolean;
  error: Error | null;
  retry: () => void;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, password: string, displayName: string) => Promise<void>;
  logout: () => Promise<void>;
  updateProfile: (p: Profile) => Promise<void>;
  refresh: () => Promise<void>;
}

const Ctx = createContext<AuthCtx | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const qc = useQueryClient();
  const { setPref } = useTheme();
  const q = useQuery({ queryKey: ["me"], queryFn: ({ signal }) => get<User | null>("/api/auth/me", signal), retry: 1, staleTime: 5 * 60_000 });
  const user = q.data ?? null;

  // A signed-in learner's saved theme preference wins over the local default.
  const savedTheme = user?.profile.theme;
  useEffect(() => {
    if (savedTheme) setPref(savedTheme);
  }, [savedTheme, setPref]);

  const startSession = useCallback(
    (u: User) => {
      // Keep the "me" query (and its observer); drop everything user-specific.
      qc.removeQueries({ predicate: (q) => q.queryKey[0] !== "me" });
      qc.setQueryData(["me"], u);
    },
    [qc],
  );

  const value = useMemo<AuthCtx>(
    () => ({
      user,
      loading: q.isLoading,
      error: q.error,
      retry: () => void q.refetch(),
      login: async (email, password) => startSession(await send<User>("POST", "/api/auth/login", { email, password })),
      register: async (email, password, displayName) =>
        startSession(
          await send<User>("POST", "/api/auth/register", {
            email,
            password,
            displayName,
            timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
          }),
        ),
      logout: async () => {
        await send("POST", "/api/auth/logout");
        qc.removeQueries({ predicate: (q) => q.queryKey[0] !== "me" });
        qc.setQueryData(["me"], null);
      },
      updateProfile: async (p) => {
        const profile = await send<Profile>("PUT", "/api/profile", p);
        qc.setQueryData<User | null>(["me"], (u) => (u ? { ...u, profile } : u));
        await qc.invalidateQueries({ predicate: (query) => query.queryKey[0] !== "me" });
      },
      refresh: async () => {
        await qc.invalidateQueries({ predicate: (query) => query.queryKey[0] !== "me" });
      },
    }),
    [user, q, qc, startSession],
  );
  return <Ctx.Provider value={value}>{children}</Ctx.Provider>;
}

export function useAuth(): AuthCtx {
  const c = useContext(Ctx);
  if (!c) throw new Error("useAuth outside AuthProvider");
  return c;
}
