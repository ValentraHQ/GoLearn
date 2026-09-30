import { useEffect, useMemo, useRef, useState } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { useNavigate } from "react-router-dom";
import { BookOpen, BookText, Code2, FileCode2, FolderKanban, Layers, Search } from "lucide-react";
import { useSearch } from "@/api/queries";
import type { SearchResult } from "@/api/types";
import { ErrorBlock } from "@/components/ui/async";
import { cn } from "@/lib/utils";

const icons: Record<SearchResult["kind"], typeof Search> = {
  lesson: BookOpen,
  module: Layers,
  challenge: Code2,
  project: FolderKanban,
  example: FileCode2,
  glossary: BookText,
};

function useDebounced<T>(value: T, ms: number): T {
  const [v, setV] = useState(value);
  useEffect(() => {
    const t = setTimeout(() => setV(value), ms);
    return () => clearTimeout(t);
  }, [value, ms]);
  return v;
}

export function SearchPalette({ open, onOpenChange }: { open: boolean; onOpenChange: (o: boolean) => void }) {
  const [input, setInput] = useState("");
  const q = useDebounced(input, 150);
  const [active, setActive] = useState(0);
  const nav = useNavigate();
  const query = useSearch(q);
  const listRef = useRef<HTMLDivElement>(null);
  const results = useMemo(() => (q.trim() ? (query.data?.results ?? []) : []), [q, query.data]);

  // Group by section while keeping a flat index for arrow-key navigation.
  const groups = useMemo(() => {
    const order: string[] = [];
    const by = new Map<string, { r: SearchResult; i: number }[]>();
    results.forEach((r, i) => {
      if (!by.has(r.group)) {
        by.set(r.group, []);
        order.push(r.group);
      }
      by.get(r.group)!.push({ r, i });
    });
    return order.map((g) => ({ name: g, items: by.get(g)! }));
  }, [results]);

  useEffect(() => setActive(0), [results]);
  useEffect(() => {
    if (!open) setInput("");
  }, [open]);
  useEffect(() => {
    listRef.current?.querySelector(`[data-idx="${active}"]`)?.scrollIntoView({ block: "nearest" });
  }, [active]);

  const go = (r: SearchResult) => {
    onOpenChange(false);
    nav(r.path);
  };

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "ArrowDown") {
      e.preventDefault();
      setActive((a) => Math.min(a + 1, results.length - 1));
    } else if (e.key === "ArrowUp") {
      e.preventDefault();
      setActive((a) => Math.max(a - 1, 0));
    } else if (e.key === "Enter" && results[active]) {
      e.preventDefault();
      go(results[active]);
    }
  };

  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-50 bg-black/50" />
        <Dialog.Content
          aria-describedby={undefined}
          className="fixed left-1/2 top-[10vh] z-50 w-[calc(100vw-2rem)] max-w-xl -translate-x-1/2 overflow-hidden rounded-xl border border-border bg-surface shadow-card"
        >
          <Dialog.Title className="sr-only">Search</Dialog.Title>
          <div className="flex items-center gap-2 border-b border-border px-3">
            <Search className="h-4 w-4 text-muted" aria-hidden />
            <input
              autoFocus
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={onKeyDown}
              role="combobox"
              aria-expanded={results.length > 0}
              aria-controls="search-results"
              aria-activedescendant={results.length ? `search-opt-${active}` : undefined}
              aria-label="Search lessons, challenges, projects and glossary"
              placeholder='Search e.g. "goroutine"'
              className="h-12 flex-1 bg-transparent text-sm outline-none placeholder:text-faint"
            />
          </div>
          <div ref={listRef} id="search-results" role="listbox" aria-label="Search results" className="max-h-[60vh] overflow-y-auto p-2">
            {!q.trim() && <p className="px-3 py-6 text-center text-sm text-muted">Type to search across lessons, challenges, projects, code examples and the glossary.</p>}
            {q.trim() && query.isFetching && results.length === 0 && <p className="px-3 py-6 text-center text-sm text-muted">Searching…</p>}
            {q.trim() && query.isError && <ErrorBlock error={query.error} onRetry={() => void query.refetch()} />}
            {q.trim() && !query.isFetching && !query.isError && results.length === 0 && (
              <p className="px-3 py-6 text-center text-sm text-muted">No results for “{q}”.</p>
            )}
            {groups.map((g) => (
              <div key={g.name} role="group" aria-label={g.name} className="mb-2">
                <p className="px-3 pb-1 pt-2 text-xs font-semibold uppercase tracking-wide text-faint">{g.name}</p>
                {g.items.map(({ r, i }) => {
                  const Icon = icons[r.kind];
                  return (
                    <div
                      key={`${r.kind}-${r.id}-${i}`}
                      id={`search-opt-${i}`}
                      data-idx={i}
                      role="option"
                      aria-selected={i === active}
                      onMouseMove={() => setActive(i)}
                      onClick={() => go(r)}
                      className={cn("flex cursor-pointer items-start gap-3 rounded-lg px-3 py-2", i === active ? "bg-accent-soft" : "hover:bg-surface-2")}
                    >
                      <Icon className="mt-0.5 h-4 w-4 shrink-0 text-muted" aria-hidden />
                      <div className="min-w-0">
                        <p className="truncate text-sm font-medium">{r.title}</p>
                        {r.snippet ? (
                          <pre className="mt-1 max-h-14 overflow-hidden font-mono text-xs text-muted">{r.snippet}</pre>
                        ) : (
                          r.subtitle && <p className="line-clamp-1 text-xs text-muted">{r.subtitle}</p>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>
            ))}
          </div>
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
