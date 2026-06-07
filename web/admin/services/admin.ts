import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type {
  AdminEmployee,
  AdminKYCRecord,
  AdminUser,
  Booking,
  BookingListResult,
  Category,
  DashboardSummary,
  EmployeeListResult,
  KYCListResult,
  UserListResult,
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

export function fetchBooking(id: string) {
  return apiRequest<Booking>(`/admin/bookings/${id}`, authOptions());
}

export function updateBookingStatus(id: string, status: string) {
  return apiRequest<Booking>(`/admin/bookings/${id}/status`, {
    ...authOptions({ method: "PATCH", body: JSON.stringify({ status }) }),
  });
}

export function fetchEmployee(id: string) {
  return apiRequest<AdminEmployee>(`/admin/employees/${id}`, authOptions());
}

export function suspendEmployee(id: string) {
  return apiRequest(`/admin/employees/${id}/suspend`, {
    ...authOptions({ method: "PATCH" }),
  });
}

export function fetchKYCRecords(params: {
  page?: number;
  limit?: number;
  status?: string;
}) {
  return apiRequest<KYCListResult>("/admin/kyc", {
    ...authOptions(),
    params,
  });
}

export function fetchKYCRecord(id: string) {
  return apiRequest<AdminKYCRecord>(`/admin/kyc/${id}`, authOptions());
}

export function approveKYC(id: string) {
  return apiRequest(`/admin/kyc/${id}/approve`, {
    ...authOptions({ method: "PATCH" }),
  });
}

export function rejectKYC(id: string, reason: string) {
  return apiRequest(`/admin/kyc/${id}/reject`, {
    ...authOptions({ method: "PATCH", body: JSON.stringify({ reason }) }),
  });
}

export function fetchUsers(params: {
  page?: number;
  limit?: number;
  role?: string;
  status?: string;
  q?: string;
}) {
  return apiRequest<UserListResult>("/admin/users", {
    ...authOptions(),
    params,
  });
}

export function fetchUser(id: string) {
  return apiRequest<AdminUser>(`/admin/users/${id}`, authOptions());
}

export function suspendUser(id: string) {
  return apiRequest(`/admin/users/${id}/suspend`, {
    ...authOptions({ method: "PATCH" }),
  });
}

export function activateUser(id: string) {
  return apiRequest(`/admin/users/${id}/activate`, {
    ...authOptions({ method: "PATCH" }),
  });
}
