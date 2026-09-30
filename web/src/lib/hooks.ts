import { useCallback, useEffect, useState } from "react";
import { send } from "@/api/client";
import { useAuth } from "@/lib/auth";

/** Records learning time while the tab is visible (30 s heartbeats). */
export function useLearningTimer() {
  const { user } = useAuth();
  const signedIn = !!user;
  useEffect(() => {
    if (!signedIn) return;
    const id = window.setInterval(() => {
      if (document.visibilityState === "visible" && document.hasFocus()) {
        send("POST", "/api/sessions/ping", { seconds: 30 }).catch(() => undefined);
      }
    }, 30_000);
    return () => window.clearInterval(id);
  }, [signedIn]);
}

/** Unsent-draft persistence in localStorage; silently a no-op if storage is blocked. */
export function useDraft(key: string): [string | null, (v: string | null) => void] {
  const read = useCallback(() => {
    try {
      return localStorage.getItem(key);
    } catch {
      return null;
    }
  }, [key]);
  const [value, setValue] = useState<string | null>(read);
  useEffect(() => setValue(read()), [read]);
  const write = useCallback(
    (v: string | null) => {
      try {
        if (v === null) localStorage.removeItem(key);
        else localStorage.setItem(key, v);
      } catch {
        /* ignore */
      }
    },
    [key],
  );
  return [value, write];
}

export function useDocumentTitle(title: string | undefined) {
  useEffect(() => {
    if (title) document.title = `${title} · GoLearn`;
    return () => {
      document.title = "GoLearn — Learn Go. Build Real Software.";
    };
  }, [title]);
}
