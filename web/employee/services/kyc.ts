import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type { ChangePasswordRequest } from "@/types/profile";
import type { KYCRecord, SubmitKYCRequest } from "@/types/kyc";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchKYC() {
  return apiRequest<KYCRecord>("/employee/kyc", authOptions());
}

export function submitKYC(body: SubmitKYCRequest) {
  return apiRequest<KYCRecord>("/employee/kyc", {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}

export function changePassword(body: ChangePasswordRequest) {
  return apiRequest("/auth/change-password", {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}
