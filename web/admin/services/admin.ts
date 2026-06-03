import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type {
  BookingListResult,
  Category,
  DashboardSummary,
  EmployeeListResult,
} from "@/types/admin";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchDashboardSummary() {
  return apiRequest<DashboardSummary>("/admin/dashboard/summary", authOptions());
}

export function fetchEmployees(params: {
  page?: number;
  limit?: number;
  verification_status?: string;
  q?: string;
}) {
  return apiRequest<EmployeeListResult>("/admin/employees", {
    ...authOptions(),
    params,
  });
}

export function approveEmployee(id: string) {
  return apiRequest(`/admin/employees/${id}/approve`, {
    ...authOptions({ method: "PATCH" }),
  });
}

export function rejectEmployee(id: string) {
  return apiRequest(`/admin/employees/${id}/reject`, {
    ...authOptions({ method: "PATCH" }),
  });
}

export function fetchCategories() {
  return apiRequest<Category[]>("/categories", authOptions());
}

export function createCategory(body: {
  name: string;
  description?: string;
  is_active?: boolean;
}) {
  return apiRequest<Category>("/admin/categories", {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}

export function updateCategory(
  id: string,
  body: { name: string; description?: string; is_active?: boolean },
) {
  return apiRequest<Category>(`/admin/categories/${id}`, {
    ...authOptions({ method: "PUT", body: JSON.stringify(body) }),
  });
}

export function deleteCategory(id: string) {
  return apiRequest(`/admin/categories/${id}`, {
    ...authOptions({ method: "DELETE" }),
  });
}

export function fetchBookings(params: {
  page?: number;
  limit?: number;
  status?: string;
}) {
  return apiRequest<BookingListResult>("/admin/bookings", {
    ...authOptions(),
    params,
  });
}
