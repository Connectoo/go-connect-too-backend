"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchSummary } from "@/services/analytics";

export function useEmployeeSummary() {
  return useQuery({
    queryKey: ["employee", "analytics", "summary"],
    queryFn: fetchSummary,
  });
}
