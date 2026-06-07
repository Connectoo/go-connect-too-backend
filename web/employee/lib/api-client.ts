import type { ApiEnvelope } from "@/types/api";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

type RequestOptions = RequestInit & {
  token?: string | null;
  params?: Record<string, string | number | undefined | null>;
};

function buildUrl(path: string, params?: RequestOptions["params"]) {
  const base = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";
  const url = new URL(path.startsWith("http") ? path : `${base}${path}`);
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== "") {
        url.searchParams.set(key, String(value));
      }
    });
  }
  return url.toString();
}

export async function apiRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const { token, params, headers, ...init } = options;
  const url = buildUrl(path, params);

  const response = await fetch(url, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...headers,
    },
  });

  const body = (await response.json()) as ApiEnvelope<T>;

  if (!response.ok || !body.success) {
    throw new ApiError(
      body.message || "Request failed",
      response.status,
      body.error,
    );
  }

  return body.data as T;
}
