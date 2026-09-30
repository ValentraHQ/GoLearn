import { beforeEach, describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RunPanel } from "./RunPanel";

// CodeMirror needs real layout; the panel logic is what we test here.
vi.mock("./CodeEditor", () => ({ default: () => <div data-testid="editor" /> }));

const auth = vi.hoisted(() => ({ user: null as null | { id: string } }));
vi.mock("@/lib/auth", () => ({ useAuth: () => ({ user: auth.user }) }));

const config = vi.hoisted(() => ({ data: { runner: { available: true } } as { runner: { available: boolean; reason?: string } } }));
vi.mock("@/api/queries", () => ({ useConfig: () => ({ data: config.data }) }));

function renderPanel() {
  const client = new QueryClient();
  return render(
    <QueryClientProvider client={client}>
      <MemoryRouter>
        <RunPanel label="test" initial="package main" onRun={async () => ({ run: undefined })} />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("RunPanel", () => {
  beforeEach(() => {
    auth.user = null;
    config.data = { runner: { available: true } };
  });

  it("asks anonymous visitors to sign in and disables Run", async () => {
    renderPanel();
    expect(await screen.findByRole("link", { name: /sign in/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /run/i })).toHaveAttribute("aria-disabled", "true");
  });

  it("explains when the code runner is unavailable instead of pretending to work", async () => {
    auth.user = { id: "u1" };
    config.data = { runner: { available: false, reason: "Docker is not reachable by the server." } };
    renderPanel();
    expect(await screen.findByText(/Code execution is unavailable: Docker is not reachable/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /run/i })).toHaveAttribute("aria-disabled", "true");
  });

  it("enables Run for signed-in users when the runner is available", async () => {
    auth.user = { id: "u1" };
    renderPanel();
    expect(await screen.findByTestId("editor")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /^run$/i })).not.toHaveAttribute("aria-disabled", "true");
    expect(screen.queryByText(/unavailable/i)).not.toBeInTheDocument();
  });
});
