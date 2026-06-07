import { apiRequest } from "@/lib/api-client";
import type { AuthResponse, LoginInput, RegisterInput } from "@/types/auth";

export function loginCustomer(input: LoginInput) {
  return apiRequest<AuthResponse>("/auth/login/customer", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function registerCustomer(input: RegisterInput) {
  return apiRequest<AuthResponse>("/auth/register/customer", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function forgotPassword(email: string) {
  return apiRequest<null>("/auth/forgot-password", {
    method: "POST",
    body: JSON.stringify({ email, role: "customer" }),
  });
}

export function resetPassword(token: string, newPassword: string) {
  return apiRequest<null>("/auth/reset-password", {
    method: "POST",
    body: JSON.stringify({ token, new_password: newPassword }),
  });
}

export function verifyEmail(token: string) {
  return apiRequest<null>("/auth/verify-email", {
    method: "POST",
    body: JSON.stringify({ token }),
  });
}
