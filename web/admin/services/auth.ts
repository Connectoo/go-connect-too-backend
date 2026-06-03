import { apiRequest } from "@/lib/api-client";
import type { AuthResponse, LoginInput } from "@/types/auth";

export function loginAdmin(input: LoginInput) {
  return apiRequest<AuthResponse>("/auth/login/admin", {
    method: "POST",
    body: JSON.stringify(input),
  });
}
