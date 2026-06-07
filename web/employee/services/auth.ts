import { apiRequest } from "@/lib/api-client";
import type { AuthResponse, LoginInput, RegisterInput } from "@/types/auth";

export function loginEmployee(input: LoginInput) {
  return apiRequest<AuthResponse>("/auth/login/employee", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function registerEmployee(input: RegisterInput) {
  return apiRequest<AuthResponse>("/auth/register/employee", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function forgotPassword(email: string) {
  return apiRequest<null>("/auth/forgot-password", {
    method: "POST",
    body: JSON.stringify({ email, role: "employee" }),
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
