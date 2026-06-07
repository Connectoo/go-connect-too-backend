import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type { EmployeeService, EmployeeServiceRequest } from "@/types/service";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchServices() {
  return apiRequest<EmployeeService[]>("/employee/services", authOptions());
}

export function createService(body: EmployeeServiceRequest) {
  return apiRequest<EmployeeService>("/employee/services", {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}

export function updateService(id: string, body: EmployeeServiceRequest) {
  return apiRequest<EmployeeService>(`/employee/services/${id}`, {
    ...authOptions({ method: "PUT", body: JSON.stringify(body) }),
  });
}

export function updateServiceStatus(id: string, isActive: boolean) {
  return apiRequest<EmployeeService>(`/employee/services/${id}/status`, {
    ...authOptions({
      method: "PATCH",
      body: JSON.stringify({ is_active: isActive }),
    }),
  });
}

export function deleteService(id: string) {
  return apiRequest(`/employee/services/${id}`, {
    ...authOptions({ method: "DELETE" }),
  });
}
