"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  approveEmployee,
  createCategory,
  deleteCategory,
  fetchBookings,
  fetchCategories,
  fetchDashboardSummary,
  fetchEmployees,
  rejectEmployee,
  updateCategory,
} from "@/services/admin";

export function useDashboardSummary() {
  return useQuery({
    queryKey: ["admin", "dashboard"],
    queryFn: fetchDashboardSummary,
  });
}

export function useAdminEmployees(params: {
  page: number;
  limit: number;
  verification_status?: string;
  q?: string;
}) {
  return useQuery({
    queryKey: ["admin", "employees", params],
    queryFn: () => fetchEmployees(params),
  });
}

export function useEmployeeActions() {
  const qc = useQueryClient();
  return {
    approve: useMutation({
      mutationFn: approveEmployee,
      onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "employees"] }),
    }),
    reject: useMutation({
      mutationFn: rejectEmployee,
      onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "employees"] }),
    }),
  };
}

export function useAdminCategories() {
  return useQuery({
    queryKey: ["admin", "categories"],
    queryFn: fetchCategories,
  });
}

export function useCategoryMutations() {
  const qc = useQueryClient();
  return {
    create: useMutation({
      mutationFn: createCategory,
      onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "categories"] }),
    }),
    update: useMutation({
      mutationFn: ({ id, ...body }: { id: string; name: string; description?: string; is_active?: boolean }) =>
        updateCategory(id, body),
      onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "categories"] }),
    }),
    remove: useMutation({
      mutationFn: deleteCategory,
      onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "categories"] }),
    }),
  };
}

export function useAdminBookings(params: {
  page: number;
  limit: number;
  status?: string;
}) {
  return useQuery({
    queryKey: ["admin", "bookings", params],
    queryFn: () => fetchBookings(params),
  });
}
