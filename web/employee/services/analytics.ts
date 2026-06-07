import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type { EmployeeSummary } from "@/types/analytics";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchSummary() {
  return apiRequest<EmployeeSummary>("/employee/analytics/summary", authOptions());
}
