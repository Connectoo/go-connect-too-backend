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
