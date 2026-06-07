import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type { CreateReportInput } from "@/types/report";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function submitReport(body: CreateReportInput) {
  return apiRequest("/reports", {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}
