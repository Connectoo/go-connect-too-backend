import type { AuthResponse } from "@/types/auth";

export const EMPLOYEE_AUTH_KEY = "go_connect_employee_auth";
export const EMPLOYEE_TOKEN_COOKIE = "employee_access_token";

export type StoredEmployeeAuth = {
  user: NonNullable<AuthResponse["user"]>;
  tokens: AuthResponse["tokens"];
};

export function saveEmployeeAuth(auth: StoredEmployeeAuth) {
  if (typeof window === "undefined") return;
  localStorage.setItem(EMPLOYEE_AUTH_KEY, JSON.stringify(auth));
  document.cookie = `${EMPLOYEE_TOKEN_COOKIE}=${auth.tokens.access_token}; path=/; max-age=${60 * 60 * 8}; SameSite=Lax`;
}

export function loadEmployeeAuth(): StoredEmployeeAuth | null {
  if (typeof window === "undefined") return null;
  const raw = localStorage.getItem(EMPLOYEE_AUTH_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as StoredEmployeeAuth;
  } catch {
    return null;
  }
}

export function getAccessToken(): string | null {
  const auth = loadEmployeeAuth();
  return auth?.tokens.access_token ?? null;
}

export function clearEmployeeAuth() {
  if (typeof window === "undefined") return;
  localStorage.removeItem(EMPLOYEE_AUTH_KEY);
  document.cookie = `${EMPLOYEE_TOKEN_COOKIE}=; path=/; max-age=0`;
}

export function isEmployeeUser(user: StoredEmployeeAuth["user"]) {
  return user.role === "employee";
}
