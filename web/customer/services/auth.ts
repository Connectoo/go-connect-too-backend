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
