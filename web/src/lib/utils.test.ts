import { describe, expect, it } from "vitest";
import { formatDuration, isTypingTarget, cn } from "./utils";

describe("formatDuration", () => {
  it("formats seconds, minutes and hours", () => {
    expect(formatDuration(0)).toBe("0s");
    expect(formatDuration(45)).toBe("45s");
    expect(formatDuration(60)).toBe("1m");
    expect(formatDuration(3600)).toBe("1h 0m");
    expect(formatDuration(5400)).toBe("1h 30m");
  });
});

describe("isTypingTarget", () => {
  it("treats inputs and the code editor as typing targets", () => {
    const input = document.createElement("input");
    const div = document.createElement("div");
    const editor = document.createElement("div");
    editor.className = "cm-editor";
    const inner = document.createElement("span");
    editor.appendChild(inner);
    expect(isTypingTarget(input)).toBe(true);
    expect(isTypingTarget(inner)).toBe(true);
    expect(isTypingTarget(div)).toBe(false);
    expect(isTypingTarget(null)).toBe(false);
  });
});

describe("cn", () => {
  it("joins truthy class names", () => {
    expect(cn("a", false && "b", "c")).toBe("a c");
  });
});
