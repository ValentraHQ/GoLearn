import { useEffect, useRef } from "react";
import { EditorState, Compartment } from "@codemirror/state";
import { EditorView, keymap, lineNumbers, highlightActiveLine, highlightActiveLineGutter, drawSelection } from "@codemirror/view";
import { defaultKeymap, history, historyKeymap, indentWithTab } from "@codemirror/commands";
import { bracketMatching, indentOnInput, syntaxHighlighting, HighlightStyle } from "@codemirror/language";
import { closeBrackets, closeBracketsKeymap } from "@codemirror/autocomplete";
import { go } from "@codemirror/lang-go";
import { json } from "@codemirror/lang-json";
import { tags as t } from "@lezer/highlight";

const highlight = HighlightStyle.define([
  { tag: [t.keyword, t.controlKeyword, t.definitionKeyword, t.moduleKeyword, t.operatorKeyword], color: "var(--hl-keyword)" },
  { tag: [t.string, t.special(t.string)], color: "var(--hl-string)" },
  { tag: [t.number, t.bool, t.null], color: "var(--hl-number)" },
  { tag: [t.comment, t.lineComment, t.blockComment], color: "var(--hl-comment)", fontStyle: "italic" },
  { tag: [t.function(t.variableName), t.function(t.propertyName)], color: "var(--hl-func)" },
  { tag: [t.typeName, t.className, t.standard(t.typeName)], color: "var(--hl-type)" },
]);

export interface CodeEditorProps {
  value: string;
  onChange?: (v: string) => void;
  onRun?: () => void;
  readOnly?: boolean;
  label: string;
  language?: "go" | "json";
  minHeight?: string;
  maxHeight?: string;
}

export default function CodeEditor({ value, onChange, onRun, readOnly, label, language = "go", minHeight = "10rem", maxHeight = "28rem" }: CodeEditorProps) {
  const host = useRef<HTMLDivElement>(null);
  const view = useRef<EditorView | null>(null);
  const onChangeRef = useRef(onChange);
  const onRunRef = useRef(onRun);
  const ro = useRef(new Compartment());
  onChangeRef.current = onChange;
  onRunRef.current = onRun;

  useEffect(() => {
    if (!host.current) return;
    const v = new EditorView({
      parent: host.current,
      state: EditorState.create({
        doc: value,
        extensions: [
          lineNumbers(),
          highlightActiveLineGutter(),
          highlightActiveLine(),
          drawSelection(),
          history(),
          indentOnInput(),
          bracketMatching(),
          closeBrackets(),
          language === "json" ? json() : go(),
          syntaxHighlighting(highlight),
          ro.current.of(EditorState.readOnly.of(!!readOnly)),
          EditorView.contentAttributes.of({
            "aria-label": label,
            "aria-describedby": "editor-hint",
            "aria-multiline": "true",
          }),
          EditorView.theme({
            "&": { maxHeight, minHeight },
            ".cm-scroller": { overflow: "auto", fontFamily: "var(--font-mono)" },
            ".cm-content": { padding: "10px 0" },
          }),
          keymap.of([
            { key: "Mod-Enter", preventDefault: true, run: () => (onRunRef.current?.(), true) },
            { key: "Escape", run: () => (v.contentDOM.blur(), false) },
            ...closeBracketsKeymap,
            ...defaultKeymap,
            ...historyKeymap,
            // Tab indents; Escape then Tab leaves the editor (keyboard-trap escape hatch).
            indentWithTab,
          ]),
          EditorView.updateListener.of((u) => {
            if (u.docChanged) onChangeRef.current?.(u.state.doc.toString());
          }),
        ],
      }),
    });
    view.current = v;
    return () => {
      v.destroy();
      view.current = null;
    };
    // The editor is created once; value/readOnly are synced by the effects below.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [label, language]);

  useEffect(() => {
    const v = view.current;
    if (v && v.state.doc.toString() !== value) v.dispatch({ changes: { from: 0, to: v.state.doc.length, insert: value } });
  }, [value]);

  useEffect(() => {
    view.current?.dispatch({ effects: ro.current.reconfigure(EditorState.readOnly.of(!!readOnly)) });
  }, [readOnly]);

  return <div ref={host} data-testid="code-editor" />;
}
