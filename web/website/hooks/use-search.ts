"use client";

import { useQuery } from "@tanstack/react-query";
import { searchEmployees, searchServices } from "@/services/search";

export function useServiceSearch(params: {
  q?: string;
  category_id?: string;
  enabled?: boolean;
}) {
  return useQuery({
    queryKey: ["search", "services", params],
    queryFn: () =>
      searchServices({
        q: params.q,
        category_id: params.category_id,
        limit: 20,
      }),
    enabled: params.enabled !== false && Boolean(params.q?.trim()),
  });
}

export function useEmployeeSearch(params: { q?: string; enabled?: boolean }) {
  return useQuery({
    queryKey: ["search", "employees", params],
    queryFn: () => searchEmployees({ q: params.q, limit: 20 }),
    enabled: params.enabled !== false && Boolean(params.q?.trim()),
  });
}
