import { apiRequest } from "@/lib/api-client";
import type { EmployeeSearchItem, ServiceSearchItem } from "@/types/search";

export function searchServices(params: {
  q?: string;
  category_id?: string;
  limit?: number;
}) {
  return apiRequest<ServiceSearchItem[]>("/search/services", { params });
}

export function searchEmployees(params: { q?: string; limit?: number }) {
  return apiRequest<EmployeeSearchItem[]>("/search/employees", { params });
}
