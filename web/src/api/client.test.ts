import { afterEach, describe, expect, it, vi } from "vitest";
import { ApiError, get, qs, send } from "./client";

const okJson = (body: unknown, status = 200) =>
  Promise.resolve(new Response(JSON.stringify(body), { status, headers: { "Content-Type": "application/json" } }));

afterEach(() => vi.restoreAllMocks());

describe("api client", () => {
  it("unwraps the data envelope", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(() => okJson({ data: { hello: "world" } }));
    await expect(get<{ hello: string }>("/api/x")).resolves.toEqual({ hello: "world" });
  });

  it("sends the CSRF header and JSON body on mutations", async () => {
    const spy = vi.spyOn(globalThis, "fetch").mockImplementation(() => okJson({ data: { ok: true } }));
    await send("POST", "/api/things", { a: 1 });
    const init = spy.mock.calls[0]![1]!;
    const headers = init.headers as Record<string, string>;
    expect(headers["X-GoLearn-CSRF"]).toBe("1");
    expect(headers["Content-Type"]).toBe("application/json");
    expect(init.body).toBe(JSON.stringify({ a: 1 }));
    expect(init.credentials).toBe("same-origin");
  });

  it("turns error envelopes into ApiError with the server message", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(() =>
      okJson({ error: { code: "rate_limited", message: "Slow down." } }, 429),
    );
    const err = await get("/api/x").catch((e) => e);
    expect(err).toBeInstanceOf(ApiError);
    expect(err).toMatchObject({ status: 429, code: "rate_limited", message: "Slow down." });
  });

  it("falls back to a generic message for non-JSON failures", async () => {
    vi.spyOn(globalThis, "fetch").mockImplementation(() => Promise.resolve(new Response("boom", { status: 502 })));
    await expect(get("/api/x")).rejects.toMatchObject({ status: 502, message: "Request failed (502)." });
  });

  it("reports network failures clearly", async () => {
    vi.spyOn(globalThis, "fetch").mockRejectedValue(new TypeError("Failed to fetch"));
    await expect(get("/api/x")).rejects.toMatchObject({ code: "network" });
  });
});

describe("qs", () => {
  it("skips empty values and encodes the rest", () => {
    expect(qs({ a: "x y", b: "", c: undefined, d: 3 })).toBe("?a=x+y&d=3");
    expect(qs({})).toBe("");
  });
});
