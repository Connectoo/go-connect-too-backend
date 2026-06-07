"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  activateService,
  approveReview,
  createSubscriptionPlan,
  deactivateService,
  exportReportsCsv,
  fetchAdminAnalyticsSummary,
  fetchAdminBookingsAnalytics,
  fetchAdminCategoriesAnalytics,
  fetchAdminRevenueAnalytics,
  fetchAdminReviews,
  fetchAdminServices,
  fetchPayments,
  fetchReports,
  fetchSettings,
  fetchSubscriptionPlans,
  fetchSubscriptions,
  fetchSupportTicket,
  fetchSupportTickets,
  hideReview,
  refundPayment,
  replySupportTicket,
  resolveReport,
  updateSettings,
  updateSubscriptionPlan,
  updateSupportTicket,
} from "@/services/ops";
import type { PlatformSettings } from "@/types/ops";

export function useAdminPayments(params: {
  page: number;
  limit: number;
  status?: string;
}) {
  return useQuery({
    queryKey: ["admin", "payments", params],
    queryFn: () => fetchPayments(params),
  });
}

export function usePaymentRefund() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason?: string }) =>
      refundPayment(id, reason ? { reason } : undefined),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "payments"] }),
  });
}

export function useAdminSubscriptions(params: {
  page: number;
  limit: number;
  status?: string;
}) {
  return useQuery({
    queryKey: ["admin", "subscriptions", params],
    queryFn: () => fetchSubscriptions(params),
  });
}

export function useSubscriptionPlans() {
  return useQuery({
    queryKey: ["admin", "subscription-plans"],
    queryFn: fetchSubscriptionPlans,
  });
}

export function usePlanMutations() {
  const qc = useQueryClient();
  return {
    create: useMutation({
      mutationFn: createSubscriptionPlan,
      onSuccess: () =>
        qc.invalidateQueries({ queryKey: ["admin", "subscription-plans"] }),
    }),
    update: useMutation({
      mutationFn: ({
        id,
        ...body
      }: {
        id: string;
        name: string;
        price: number;
        currency: string;
        duration_days: number;
        service_limit: number;
        is_featured_allowed: boolean;
        is_priority_allowed: boolean;
        is_active?: boolean;
      }) => updateSubscriptionPlan(id, body),
      onSuccess: () =>
        qc.invalidateQueries({ queryKey: ["admin", "subscription-plans"] }),
    }),
  };
}

export function useAdminAnalyticsSummary(from?: string, to?: string) {
  return useQuery({
    queryKey: ["admin", "analytics", "summary", from, to],
    queryFn: () => fetchAdminAnalyticsSummary({ from, to }),
  });
}

export function useAdminRevenueAnalytics(from?: string, to?: string) {
  return useQuery({
    queryKey: ["admin", "analytics", "revenue", from, to],
    queryFn: () => fetchAdminRevenueAnalytics({ from, to }),
  });
}

export function useAdminBookingsAnalytics(from?: string, to?: string) {
  return useQuery({
    queryKey: ["admin", "analytics", "bookings", from, to],
    queryFn: () => fetchAdminBookingsAnalytics({ from, to }),
  });
}

export function useAdminCategoriesAnalytics(from?: string, to?: string) {
  return useQuery({
    queryKey: ["admin", "analytics", "categories", from, to],
    queryFn: () => fetchAdminCategoriesAnalytics({ from, to }),
  });
}

export function useAdminServices(params: {
  page: number;
  limit: number;
  is_active?: string;
  q?: string;
}) {
  return useQuery({
    queryKey: ["admin", "services", params],
    queryFn: () => fetchAdminServices(params),
  });
}

export function useServiceModeration() {
  const qc = useQueryClient();
  return {
    activate: useMutation({
      mutationFn: activateService,
      onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "services"] }),
    }),
    deactivate: useMutation({
      mutationFn: deactivateService,
      onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "services"] }),
    }),
  };
}

export function useAdminReviews(params: {
  page: number;
  limit: number;
  status?: string;
}) {
  return useQuery({
    queryKey: ["admin", "reviews", params],
    queryFn: () => fetchAdminReviews(params),
  });
}

export function useReviewModeration() {
  const qc = useQueryClient();
  return {
    approve: useMutation({
      mutationFn: approveReview,
      onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "reviews"] }),
    }),
    hide: useMutation({
      mutationFn: hideReview,
      onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "reviews"] }),
    }),
  };
}

export function useAdminReports(params: {
  page: number;
  limit: number;
  status?: string;
}) {
  return useQuery({
    queryKey: ["admin", "reports", params],
    queryFn: () => fetchReports(params),
  });
}

export function useReportActions() {
  const qc = useQueryClient();
  return {
    resolve: useMutation({
      mutationFn: resolveReport,
      onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "reports"] }),
    }),
    exportCsv: useMutation({
      mutationFn: exportReportsCsv,
    }),
  };
}

export function useAdminSupportTickets(params: {
  page: number;
  limit: number;
  status?: string;
}) {
  return useQuery({
    queryKey: ["admin", "support", params],
    queryFn: () => fetchSupportTickets(params),
  });
}

export function useAdminSupportTicket(id: string) {
  return useQuery({
    queryKey: ["admin", "support", id],
    queryFn: () => fetchSupportTicket(id),
    enabled: Boolean(id),
  });
}

export function useSupportTicketActions(id: string) {
  const qc = useQueryClient();
  return {
    update: useMutation({
      mutationFn: (body: { status?: string; priority?: string }) =>
        updateSupportTicket(id, body),
      onSuccess: () => {
        qc.invalidateQueries({ queryKey: ["admin", "support"] });
        qc.invalidateQueries({ queryKey: ["admin", "support", id] });
      },
    }),
    reply: useMutation({
      mutationFn: (message: string) => replySupportTicket(id, message),
      onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "support", id] }),
    }),
  };
}

export function useAdminSettings() {
  return useQuery({
    queryKey: ["admin", "settings"],
    queryFn: fetchSettings,
  });
}

export function useSettingsMutation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: PlatformSettings) => updateSettings(body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "settings"] }),
  });
}
