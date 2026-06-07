"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  activateUser,
  approveEmployee,
  approveKYC,
  createCategory,
  deleteCategory,
  fetchBooking,
  fetchBookings,
  fetchCategories,
  fetchDashboardSummary,
  fetchEmployee,
  fetchEmployees,
  fetchKYCRecord,
  fetchKYCRecords,
  fetchUser,
  fetchUsers,
  rejectEmployee,
  rejectKYC,
  suspendEmployee,
  suspendUser,
  updateBookingStatus,
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

export function useAdminBooking(id: string) {
  return useQuery({
    queryKey: ["admin", "bookings", id],
    queryFn: () => fetchBooking(id),
    enabled: Boolean(id),
  });
}

export function useBookingStatusMutation(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (status: string) => updateBookingStatus(id, status),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin", "bookings"] });
      qc.invalidateQueries({ queryKey: ["admin", "bookings", id] });
    },
  });
}

export function useAdminEmployee(id: string) {
  return useQuery({
    queryKey: ["admin", "employees", id],
    queryFn: () => fetchEmployee(id),
    enabled: Boolean(id),
  });
}

export function useEmployeeSuspendMutation(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => suspendEmployee(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin", "employees"] });
      qc.invalidateQueries({ queryKey: ["admin", "employees", id] });
    },
  });
}

export function useAdminKYC(params: {
  page: number;
  limit: number;
  status?: string;
}) {
  return useQuery({
    queryKey: ["admin", "kyc", params],
    queryFn: () => fetchKYCRecords(params),
  });
}

export function useAdminKYCRecord(id: string) {
  return useQuery({
    queryKey: ["admin", "kyc", id],
    queryFn: () => fetchKYCRecord(id),
    enabled: Boolean(id),
  });
}

export function useKYCActions() {
  const qc = useQueryClient();
  return {
    approve: useMutation({
      mutationFn: approveKYC,
      onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "kyc"] }),
    }),
    reject: useMutation({
      mutationFn: ({ id, reason }: { id: string; reason: string }) =>
        rejectKYC(id, reason),
      onSuccess: () => qc.invalidateQueries({ queryKey: ["admin", "kyc"] }),
    }),
  };
}

export function useAdminUsers(params: {
  page: number;
  limit: number;
  role?: string;
  status?: string;
  q?: string;
}) {
  return useQuery({
    queryKey: ["admin", "users", params],
    queryFn: () => fetchUsers(params),
  });
}

export function useAdminUser(id: string) {
  return useQuery({
    queryKey: ["admin", "users", id],
    queryFn: () => fetchUser(id),
    enabled: Boolean(id),
  });
}

export function useUserActions(id: string) {
  const qc = useQueryClient();
  return {
    suspend: useMutation({
      mutationFn: () => suspendUser(id),
      onSuccess: () => {
        qc.invalidateQueries({ queryKey: ["admin", "users"] });
        qc.invalidateQueries({ queryKey: ["admin", "users", id] });
      },
    }),
    activate: useMutation({
      mutationFn: () => activateUser(id),
      onSuccess: () => {
        qc.invalidateQueries({ queryKey: ["admin", "users"] });
        qc.invalidateQueries({ queryKey: ["admin", "users", id] });
      },
    }),
  };
}
