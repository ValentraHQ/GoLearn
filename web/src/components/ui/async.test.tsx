import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { UseQueryResult } from "@tanstack/react-query";
import { Async } from "./async";
import { ApiError } from "@/api/client";

function q<T>(over: Partial<UseQueryResult<T>>): UseQueryResult<T> {
  return { isPending: false, isError: false, data: undefined, error: null, refetch: vi.fn(), ...over } as UseQueryResult<T>;
}

describe("Async", () => {
  it("shows a loading status while pending", () => {
    render(<Async query={q<string[]>({ isPending: true })}>{() => <p>data</p>}</Async>);
    expect(screen.getByRole("status")).toHaveTextContent(/loading/i);
  });

  it("shows the error and a retry button that refetches", async () => {
    const refetch = vi.fn();
    render(
      <Async query={q<string[]>({ isError: true, error: new ApiError(500, "internal", "Something broke"), refetch })}>
        {() => <p>data</p>}
      </Async>,
    );
    expect(screen.getByRole("alert")).toHaveTextContent("Something broke");
    await userEvent.click(screen.getByRole("button", { name: /try again/i }));
    expect(refetch).toHaveBeenCalledOnce();
  });

  it("shows the empty state when the predicate matches", () => {
    render(
      <Async query={q<string[]>({ data: [] })} isEmpty={(d) => d.length === 0} empty={<p>Nothing yet</p>}>
        {() => <p>data</p>}
      </Async>,
    );
    expect(screen.getByText("Nothing yet")).toBeInTheDocument();
  });

  it("renders children with the data on success", () => {
    render(<Async query={q<string[]>({ data: ["a", "b"] })}>{(d) => <p>{d.join("+")}</p>}</Async>);
    expect(screen.getByText("a+b")).toBeInTheDocument();
  });
});
