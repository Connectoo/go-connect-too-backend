import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { absoluteUrl } from "@/lib/middleware-url";

const PUBLIC_PATHS = ["/login"];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const isPublic = PUBLIC_PATHS.some((p) => pathname.startsWith(p));
  const token = request.cookies.get("admin_access_token")?.value;

  if (!token && !isPublic) {
    const loginUrl = absoluteUrl(request, "/login");
    loginUrl.searchParams.set("from", pathname);
    return NextResponse.redirect(loginUrl);
  }

  if (token && pathname === "/login") {
    return NextResponse.redirect(absoluteUrl(request, "/dashboard"));
  }

  if (pathname === "/") {
    return NextResponse.redirect(
      absoluteUrl(request, token ? "/dashboard" : "/login"),
    );
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon.ico).*)"],
};
