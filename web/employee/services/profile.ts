import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type {
  EmployeeProfile,
  UpdateEmployeeProfileRequest,
} from "@/types/profile";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchProfile() {
  return apiRequest<EmployeeProfile>("/employee/profile", authOptions());
}

export function updateProfile(body: UpdateEmployeeProfileRequest) {
  return apiRequest<EmployeeProfile>("/employee/profile", {
    ...authOptions({ method: "PUT", body: JSON.stringify(body) }),
  });
}
