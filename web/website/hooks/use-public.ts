"use client";

import { useQuery } from "@tanstack/react-query";
import {
  fetchCategories,
  fetchHome,
  fetchProvider,
  fetchProviders,
  fetchService,
  fetchServices,
} from "@/services/public";

export function useHome() {
  return useQuery({ queryKey: ["public", "home"], queryFn: fetchHome });
}

export function useCategories() {
  return useQuery({
    queryKey: ["public", "categories"],
    queryFn: fetchCategories,
  });
}

export function useProviders(limit = 20) {
  return useQuery({
    queryKey: ["public", "providers", limit],
    queryFn: () => fetchProviders(limit),
  });
}

export function useProvider(id: string) {
  return useQuery({
    queryKey: ["public", "provider", id],
    queryFn: () => fetchProvider(id),
    enabled: Boolean(id),
  });
}

export function useServices(params?: { category_id?: string; limit?: number }) {
  return useQuery({
    queryKey: ["public", "services", params],
    queryFn: () => fetchServices(params),
  });
}

export function useService(id: string) {
  return useQuery({
    queryKey: ["public", "service", id],
    queryFn: () => fetchService(id),
    enabled: Boolean(id),
  });
}
