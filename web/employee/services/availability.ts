import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type {
  Availability,
  CreateAvailabilityRequest,
  UpdateAvailabilityRequest,
} from "@/types/availability";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchAvailability() {
  return apiRequest<Availability[]>("/employee/availability", authOptions());
}

export function createAvailability(body: CreateAvailabilityRequest) {
  return apiRequest<Availability>("/employee/availability", {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}

export function updateAvailability(id: string, body: UpdateAvailabilityRequest) {
  return apiRequest<Availability>(`/employee/availability/${id}`, {
    ...authOptions({ method: "PUT", body: JSON.stringify(body) }),
  });
}

export function deleteAvailability(id: string) {
  return apiRequest(`/employee/availability/${id}`, {
    ...authOptions({ method: "DELETE" }),
  });
}
