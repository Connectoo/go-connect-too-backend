import { apiRequest } from "@/lib/api-client";
import type { AuthResponse, LoginInput, RegisterInput } from "@/types/auth";

export function registerCustomer(input: RegisterInput) {
  return apiRequest<AuthResponse>("/auth/register/customer", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function loginCustomer(input: LoginInput) {
  return apiRequest<AuthResponse>("/auth/login/customer", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export const AUTH_STORAGE_KEY = "go_connect_customer_auth";

export type StoredAuth = {
  user: NonNullable<AuthResponse["user"]>;
  tokens: AuthResponse["tokens"];
};

export function saveAuth(auth: StoredAuth) {
  if (typeof window !== "undefined") {
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(auth));
  }
}

export function loadAuth(): StoredAuth | null {
  if (typeof window === "undefined") return null;
  const raw = localStorage.getItem(AUTH_STORAGE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as StoredAuth;
  } catch {
    return null;
  }
}

export function clearAuth() {
  if (typeof window !== "undefined") {
    localStorage.removeItem(AUTH_STORAGE_KEY);
  }
}
