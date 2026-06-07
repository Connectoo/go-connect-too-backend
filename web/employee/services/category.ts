import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type { Category } from "@/types/category";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchCategories() {
  return apiRequest<Category[]>("/categories", authOptions());
}
