import { beforeEach, describe, expect, it, vi } from "vitest";
import { api, getTokens, setTokens, allPages, ApiError } from "./client";
const storage = new Map<string, string>();
vi.stubGlobal("localStorage", {
  getItem: (key: string) => storage.get(key) || null,
  setItem: (key: string, value: string) => storage.set(key, value),
  removeItem: (key: string) => storage.delete(key),
});
vi.stubGlobal("window", new EventTarget());
const old = {
  access_token: "old",
  refresh_token: "refresh-old",
  token_type: "Bearer" as const,
  expires_in: 900 as const,
};
beforeEach(() => {
  storage.clear();
  vi.restoreAllMocks();
});
describe("API integration client", () => {
  it("rotates refresh tokens once for simultaneous 401s and retries both requests", async () => {
    setTokens(old);
    let refreshCount = 0;
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string, options: RequestInit) => {
        if (url.endsWith("/auth/refresh")) {
          refreshCount++;
          expect(JSON.parse(options.body as string)).toEqual({
            token: "refresh-old",
          });
          await new Promise((r) => setTimeout(r, 10));
          return Response.json({
            ...old,
            access_token: "new",
            refresh_token: "refresh-new",
          });
        }
        if (new Headers(options.headers).get("Authorization") === "Bearer old")
          return Response.json({ message: "expired" }, { status: 401 });
        return Response.json({ id: 1 });
      }),
    );
    expect(
      await Promise.all([api("/users/me"), api("/users/me/teams")]),
    ).toEqual([{ id: 1 }, { id: 1 }]);
    expect(refreshCount).toBe(1);
    expect(getTokens()?.refresh_token).toBe("refresh-new");
  });
  it("clears revoked sessions without an infinite retry", async () => {
    setTokens(old);
    const fetcher = vi.fn(async () =>
      Response.json({ message: "revoked" }, { status: 401 }),
    );
    vi.stubGlobal("fetch", fetcher);
    await expect(api("/users/me")).rejects.toBeInstanceOf(ApiError);
    expect(getTokens()).toBeNull();
    expect(fetcher).toHaveBeenCalledTimes(2);
  });
  it("does not attach credentials or refresh on a failed public login", async () => {
    setTokens(old);
    const fetcher = vi.fn(async (_url: string, options: RequestInit) => {
      expect(new Headers(options.headers).has("Authorization")).toBe(false);
      return Response.json({ message: "Wrong password" }, { status: 401 });
    });
    vi.stubGlobal("fetch", fetcher);
    await expect(
      api(
        "/auth/login",
        { method: "POST", body: { email: "a@b.test", password: "x" } },
        false,
      ),
    ).rejects.toThrow("Wrong password");
    expect(fetcher).toHaveBeenCalledTimes(1);
  });
  it("handles empty 202 and 204 responses", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response(null, { status: 204 })),
    );
    expect(
      await api("/auth/verify-email", { method: "POST" }, false),
    ).toBeUndefined();
  });
  it("preserves the session on transient refresh failures", async () => {
    setTokens(old);
    vi.stubGlobal(
      "fetch",
      vi.fn(async (url: string) =>
        Response.json(
          { message: "unavailable" },
          { status: url.endsWith("/auth/refresh") ? 503 : 401 },
        ),
      ),
    );
    await expect(api("/users/me")).rejects.toThrow("unavailable");
    expect(getTokens()).toEqual(old);
  });
  it("loads subsequent pages without dropping existing query parameters", async () => {
    const fetcher = vi.fn(async (url: string) => {
      expect(url).toContain("discipline_id=chess&limit=100&offset=");
      const second = url.endsWith("offset=100");
      return Response.json({
        items: second ? [100] : Array.from({ length: 100 }, (_, i) => i),
        total: 101,
      });
    });
    vi.stubGlobal("fetch", fetcher);
    expect(await allPages("/ratings?discipline_id=chess", false)).toHaveLength(
      101,
    );
    expect(fetcher).toHaveBeenCalledTimes(2);
  });
});
