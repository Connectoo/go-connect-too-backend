import type { NextRequest } from "next/server";

/** Build an absolute URL using proxy headers (nginx → Next.js on 127.0.0.1). */
export function absoluteUrl(request: NextRequest, pathname: string): URL {
  const host =
    request.headers.get("x-forwarded-host")?.split(",")[0]?.trim() ??
    request.headers.get("host");
  const proto =
    request.headers.get("x-forwarded-proto")?.split(",")[0]?.trim() ?? "https";

  if (host) {
    return new URL(pathname, `${proto}://${host}`);
  }

  return new URL(pathname, request.url);
}
