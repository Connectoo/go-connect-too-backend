import type { AuthResponse } from "@/types/auth";

export const ADMIN_AUTH_KEY = "go_connect_admin_auth";
export const ADMIN_TOKEN_COOKIE = "admin_access_token";

export type StoredAdminAuth = {
  user: NonNullable<AuthResponse["user"]>;
  tokens: AuthResponse["tokens"];
};

export function saveAdminAuth(auth: StoredAdminAuth) {
  if (typeof window === "undefined") return;
  localStorage.setItem(ADMIN_AUTH_KEY, JSON.stringify(auth));
  document.cookie = `${ADMIN_TOKEN_COOKIE}=${auth.tokens.access_token}; path=/; max-age=${60 * 60 * 8}; SameSite=Lax`;
}

export function loadAdminAuth(): StoredAdminAuth | null {
  if (typeof window === "undefined") return null;
  const raw = localStorage.getItem(ADMIN_AUTH_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as StoredAdminAuth;
  } catch {
    return null;
  }
}

export function getAccessToken(): string | null {
  const auth = loadAdminAuth();
  return auth?.tokens.access_token ?? null;
}

export function clearAdminAuth() {
  if (typeof window === "undefined") return;
  localStorage.removeItem(ADMIN_AUTH_KEY);
  document.cookie = `${ADMIN_TOKEN_COOKIE}=; path=/; max-age=0`;
}

export function isAdminUser(user: StoredAdminAuth["user"]) {
  return user.role === "admin";
}
