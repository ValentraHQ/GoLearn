import { Children, isValidElement, type ReactNode } from "react";
import ReactMarkdown, { type Components } from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeHighlight from "rehype-highlight";
import bash from "highlight.js/lib/languages/bash";
import go from "highlight.js/lib/languages/go";
import json from "highlight.js/lib/languages/json";
import sql from "highlight.js/lib/languages/sql";
import yaml from "highlight.js/lib/languages/yaml";
import dockerfile from "highlight.js/lib/languages/dockerfile";
import { cn } from "@/lib/utils";

// Only these languages are registered, keeping the highlighter bundle small.
const languages = { go, bash, json, sql, yaml, dockerfile, sh: bash, shell: bash, yml: yaml };

const CALLOUT = /^\s*\[!(NOTE|TIP|WARNING|DANGER)\]\s*/i;

function textOf(node: ReactNode): string {
  if (typeof node === "string") return node;
  if (Array.isArray(node)) return node.map(textOf).join("");
  if (isValidElement<{ children?: ReactNode }>(node)) return textOf(node.props.children);
  return "";
}

const components: Components = {
  pre({ children, ...props }) {
    // ```text blocks are treated as diagrams: monospace, never wrapped.
    const code = Children.toArray(children)[0];
    const cls = isValidElement<{ className?: string }>(code) ? (code.props.className ?? "") : "";
    const isDiagram = /language-(text|diagram|plain)/.test(cls);
    return (
      <pre {...props} className={cn(isDiagram && "diagram")} tabIndex={0}>
        {children}
      </pre>
    );
  },
  blockquote({ children }) {
    const m = CALLOUT.exec(textOf(children));
    const kind = m?.[1]?.toLowerCase();
    const cls = kind === "warning" ? "callout-warning" : kind === "danger" ? "callout-danger" : kind === "tip" ? "callout-tip" : "";
    return (
      <blockquote className={cls} role={kind ? "note" : undefined}>
        {kind && <p className="text-xs font-semibold uppercase tracking-wide">{kind}</p>}
        {stripCallout(children)}
      </blockquote>
    );
  },
  a({ href, children }) {
    const external = href?.startsWith("http");
    return (
      <a href={href} {...(external ? { target: "_blank", rel: "noopener noreferrer" } : {})}>
        {children}
      </a>
    );
  },
};

function stripCallout(children: ReactNode): ReactNode {
  const arr = Children.toArray(children);
  return arr.map((c, i) => {
    if (i === 0 && isValidElement<{ children?: ReactNode }>(c)) {
      const inner = Children.toArray(c.props.children);
      if (typeof inner[0] === "string") {
        const first = (inner[0] as string).replace(CALLOUT, "");
        return { ...c, props: { ...c.props, children: [first, ...inner.slice(1)] } };
      }
    }
    return c;
  });
}

export function Markdown({ children, className }: { children: string; className?: string }) {
  return (
    <div className={cn("prose-go", className)}>
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[[rehypeHighlight, { languages, detect: false, plainText: ["text", "diagram", "plain"] }]]}
        components={components}
      >
        {children}
      </ReactMarkdown>
    </div>
  );
}

/** Static highlighted code block for read-only examples. */
export function CodeBlock({ code, lang }: { code: string; lang: string }) {
  return <Markdown>{"```" + lang + "\n" + code.replace(/\n$/, "") + "\n```"}</Markdown>;
}
