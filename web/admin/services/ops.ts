import { apiDownload, apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type {
  AdminAnalyticsSummary,
  AdminBookingsAnalytics,
  AdminCategoriesAnalytics,
  AdminRevenueAnalytics,
  AdminReview,
  AdminService,
  Payment,
  PaymentListResult,
  PlatformSettings,
  ReportListResult,
  ReviewListResult,
  ServiceListResult,
  Subscription,
  SubscriptionListResult,
  SubscriptionPlan,
  SupportTicketDetail,
  TicketListResult,
} from "@/types/ops";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchPayments(params: {
  page?: number;
  limit?: number;
  status?: string;
}) {
  return apiRequest<PaymentListResult>("/admin/payments", {
    ...authOptions(),
    params,
  });
}

export function refundPayment(
  id: string,
  body?: { amount?: number; reason?: string },
) {
  return apiRequest(`/admin/payments/${id}/refund`, {
    ...authOptions({
      method: "POST",
      body: JSON.stringify(body ?? {}),
    }),
  });
}

export function fetchSubscriptions(params: {
  page?: number;
  limit?: number;
  status?: string;
}) {
  return apiRequest<SubscriptionListResult>("/admin/subscriptions", {
    ...authOptions(),
    params,
  });
}

export function fetchSubscriptionPlans() {
  return apiRequest<SubscriptionPlan[]>("/subscription-plans", authOptions());
}

export function createSubscriptionPlan(body: {
  name: string;
  price: number;
  currency: string;
  duration_days: number;
  service_limit: number;
  is_featured_allowed: boolean;
  is_priority_allowed: boolean;
  is_active?: boolean;
}) {
  return apiRequest<SubscriptionPlan>("/admin/subscription-plans", {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}

export function updateSubscriptionPlan(
  id: string,
  body: {
    name: string;
    price: number;
    currency: string;
    duration_days: number;
    service_limit: number;
    is_featured_allowed: boolean;
    is_priority_allowed: boolean;
    is_active?: boolean;
  },
) {
  return apiRequest<SubscriptionPlan>(`/admin/subscription-plans/${id}`, {
    ...authOptions({ method: "PUT", body: JSON.stringify(body) }),
  });
}

export function fetchAdminAnalyticsSummary(params?: {
  from?: string;
  to?: string;
}) {
  return apiRequest<AdminAnalyticsSummary>("/admin/analytics/summary", {
    ...authOptions(),
    params,
  });
}

export function fetchAdminRevenueAnalytics(params?: {
  from?: string;
  to?: string;
}) {
  return apiRequest<AdminRevenueAnalytics>("/admin/analytics/revenue", {
    ...authOptions(),
    params,
  });
}

export function fetchAdminBookingsAnalytics(params?: {
  from?: string;
  to?: string;
}) {
  return apiRequest<AdminBookingsAnalytics>("/admin/analytics/bookings", {
    ...authOptions(),
    params,
  });
}

export function fetchAdminCategoriesAnalytics(params?: {
  from?: string;
  to?: string;
}) {
  return apiRequest<AdminCategoriesAnalytics>("/admin/analytics/categories", {
    ...authOptions(),
    params,
  });
}

export function fetchAdminServices(params: {
  page?: number;
  limit?: number;
  category_id?: string;
  is_active?: string;
  q?: string;
}) {
  return apiRequest<ServiceListResult>("/admin/services", {
    ...authOptions(),
    params,
  });
}

export function activateService(id: string) {
  return apiRequest<AdminService>(`/admin/services/${id}/activate`, {
    ...authOptions({ method: "PATCH" }),
  });
}

export function deactivateService(id: string) {
  return apiRequest<AdminService>(`/admin/services/${id}/deactivate`, {
    ...authOptions({ method: "PATCH" }),
  });
}

export function fetchAdminReviews(params: {
  page?: number;
  limit?: number;
  status?: string;
}) {
  return apiRequest<ReviewListResult>("/admin/reviews", {
    ...authOptions(),
    params,
  });
}

export function approveReview(id: string) {
  return apiRequest<AdminReview>(`/admin/reviews/${id}/approve`, {
    ...authOptions({ method: "PATCH" }),
  });
}

export function hideReview(id: string) {
  return apiRequest<AdminReview>(`/admin/reviews/${id}/hide`, {
    ...authOptions({ method: "PATCH" }),
  });
}

export function fetchReports(params: {
  page?: number;
  limit?: number;
  status?: string;
}) {
  return apiRequest<ReportListResult>("/admin/reports", {
    ...authOptions(),
    params,
  });
}

export function resolveReport(id: string) {
  return apiRequest(`/admin/reports/${id}/resolve`, {
    ...authOptions({ method: "PATCH" }),
  });
}

export function exportReportsCsv() {
  return apiDownload("/admin/reports/export", authOptions());
}

export function fetchSupportTickets(params: {
  page?: number;
  limit?: number;
  status?: string;
}) {
  return apiRequest<TicketListResult>("/admin/support/tickets", {
    ...authOptions(),
    params,
  });
}

export function fetchSupportTicket(id: string) {
  return apiRequest<SupportTicketDetail>(`/admin/support/tickets/${id}`, authOptions());
}

export function updateSupportTicket(
  id: string,
  body: { status?: string; priority?: string },
) {
  return apiRequest<SupportTicketDetail>(`/admin/support/tickets/${id}`, {
    ...authOptions({ method: "PATCH", body: JSON.stringify(body) }),
  });
}

export function replySupportTicket(id: string, message: string) {
  return apiRequest(`/admin/support/tickets/${id}/messages`, {
    ...authOptions({ method: "POST", body: JSON.stringify({ message }) }),
  });
}

export function fetchSettings() {
  return apiRequest<PlatformSettings>("/admin/settings", authOptions());
}

export function updateSettings(body: PlatformSettings) {
  return apiRequest<PlatformSettings>("/admin/settings", {
    ...authOptions({ method: "PUT", body: JSON.stringify(body) }),
  });
}
