import type { AuthResponse } from "@/types/auth";

export const CUSTOMER_AUTH_KEY = "go_connect_customer_auth";
export const CUSTOMER_TOKEN_COOKIE = "customer_access_token";

export type StoredCustomerAuth = {
  user: NonNullable<AuthResponse["user"]>;
  tokens: AuthResponse["tokens"];
};

export function saveCustomerAuth(auth: StoredCustomerAuth) {
  if (typeof window === "undefined") return;
  localStorage.setItem(CUSTOMER_AUTH_KEY, JSON.stringify(auth));
  document.cookie = `${CUSTOMER_TOKEN_COOKIE}=${auth.tokens.access_token}; path=/; max-age=${60 * 60 * 8}; SameSite=Lax`;
}

export function loadCustomerAuth(): StoredCustomerAuth | null {
  if (typeof window === "undefined") return null;
  const raw = localStorage.getItem(CUSTOMER_AUTH_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as StoredCustomerAuth;
  } catch {
    return null;
  }
}

export function getAccessToken(): string | null {
  const auth = loadCustomerAuth();
  return auth?.tokens.access_token ?? null;
}

export function clearCustomerAuth() {
  if (typeof window === "undefined") return;
  localStorage.removeItem(CUSTOMER_AUTH_KEY);
  document.cookie = `${CUSTOMER_TOKEN_COOKIE}=; path=/; max-age=0`;
}

export function isCustomerUser(user: StoredCustomerAuth["user"]) {
  return user.role === "customer";
}
