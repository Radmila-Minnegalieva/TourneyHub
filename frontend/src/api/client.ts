import type { components } from "./schema";
export type Schema = components["schemas"];
export const API_BASE = (
  import.meta.env.VITE_API_BASE_URL || "/api/v1"
).replace(/\/$/, "");
const KEY = "tourneyhub.tokens";
export function getTokens(): Schema["Tokens"] | null {
  try {
    return JSON.parse(localStorage.getItem(KEY) || "null");
  } catch {
    return null;
  }
}
export function setTokens(tokens: Schema["Tokens"] | null) {
  if (tokens) localStorage.setItem(KEY, JSON.stringify(tokens));
  else localStorage.removeItem(KEY);
  window.dispatchEvent(new Event("authchange"));
}
export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
  }
}
export const errorMessage = (error: unknown) =>
  error instanceof Error ? error.message : "Не удалось выполнить запрос";
let refreshing: Promise<void> | null = null;
async function refresh() {
  if (!refreshing)
    refreshing = (async () => {
      const token = getTokens()?.refresh_token;
      if (!token) throw new ApiError(401, "Войдите в аккаунт");
      try {
        const tokens = await api<Schema["Tokens"]>(
          "/auth/refresh",
          { method: "POST", body: { token } },
          false,
        );
        if (getTokens()?.refresh_token === token) setTokens(tokens);
      } catch (error) {
        if (error instanceof ApiError && [400, 401, 403].includes(error.status))
          setTokens(null);
        throw error;
      }
    })().finally(() => {
      refreshing = null;
    });
  return refreshing;
}
type Options = Omit<RequestInit, "body"> & { body?: unknown };
export async function api<T = void>(
  path: string,
  options: Options = {},
  authenticated = true,
  retry = true,
): Promise<T> {
  const headers = new Headers(options.headers);
  const access = authenticated ? getTokens()?.access_token : undefined;
  if (access) headers.set("Authorization", `Bearer ${access}`);
  const isForm = options.body instanceof FormData;
  if (options.body !== undefined && !isForm)
    headers.set("Content-Type", "application/json");
  let response: Response;
  try {
    response = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
      body:
        options.body === undefined
          ? undefined
          : isForm
            ? (options.body as FormData)
            : JSON.stringify(options.body),
    });
  } catch (error) {
    if (error instanceof DOMException && error.name === "AbortError")
      throw error;
    throw new Error(
      "Сервер недоступен. Проверьте подключение и повторите попытку.",
    );
  }
  if (
    response.status === 401 &&
    authenticated &&
    retry &&
    getTokens()?.refresh_token
  ) {
    // A concurrent request may already have rotated the tokens.
    if (getTokens()?.access_token === access) await refresh();
    return api<T>(path, options, authenticated, false);
  }
  if (!response.ok) {
    const data = await response.json().catch(() => null);
    if (response.status === 401 && authenticated) setTokens(null);
    const fields = data?.fields
      ?.map(
        (f: { field: string; message: string }) => `${f.field}: ${f.message}`,
      )
      .join("; ");
    throw new ApiError(
      response.status,
      [data?.message || `Ошибка запроса (${response.status})`, fields]
        .filter(Boolean)
        .join(": "),
    );
  }
  if (response.status === 204 || response.status === 202) return undefined as T;
  return response.json();
}
export async function allPages<T>(
  path: string,
  authenticated = true,
): Promise<T[]> {
  const items: T[] = [];
  for (let offset = 0; ; offset += 100) {
    const page = await api<{ items: T[]; total: number }>(
      `${path}${path.includes("?") ? "&" : "?"}limit=100&offset=${offset}`,
      {},
      authenticated,
    );
    items.push(...(page.items || []));
    if (!page.items?.length || items.length >= page.total) return items;
  }
}
