import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type { AvailabilitySlot, ServiceDetail } from "@/types/service";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

// GET /services/{id} is public; the token is harmless when present.
export function getService(id: string) {
  return apiRequest<ServiceDetail>(`/services/${id}`, authOptions());
}

// GET /employees/{id}/availability is public and returns a bare array.
export function getEmployeeAvailability(employeeId: string) {
  return apiRequest<AvailabilitySlot[]>(
    `/employees/${employeeId}/availability`,
    authOptions(),
  );
}
