"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchCategories } from "@/services/category";
import {
  createService,
  deleteService,
  fetchServices,
  updateService,
  updateServiceStatus,
} from "@/services/service";
import type { EmployeeServiceRequest } from "@/types/service";

export function useServices() {
  return useQuery({
    queryKey: ["employee", "services"],
    queryFn: fetchServices,
  });
}

export function useCategories() {
  return useQuery({
    queryKey: ["categories"],
    queryFn: fetchCategories,
  });
}

export function useServiceMutations() {
  const qc = useQueryClient();
  const invalidate = () =>
    qc.invalidateQueries({ queryKey: ["employee", "services"] });

  return {
    create: useMutation({
      mutationFn: createService,
      onSuccess: invalidate,
    }),
    update: useMutation({
      mutationFn: ({ id, body }: { id: string; body: EmployeeServiceRequest }) =>
        updateService(id, body),
      onSuccess: invalidate,
    }),
    setStatus: useMutation({
      mutationFn: ({ id, isActive }: { id: string; isActive: boolean }) =>
        updateServiceStatus(id, isActive),
      onSuccess: invalidate,
    }),
    remove: useMutation({
      mutationFn: deleteService,
      onSuccess: invalidate,
    }),
  };
}
