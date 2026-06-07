"use client";

import { useQuery } from "@tanstack/react-query";
import { getEmployeeAvailability, getService } from "@/services/catalog";

export function useService(id: string) {
  return useQuery({
    queryKey: ["service", id],
    queryFn: () => getService(id),
    enabled: Boolean(id),
  });
}

export function useEmployeeAvailability(employeeId: string | undefined) {
  return useQuery({
    queryKey: ["availability", employeeId],
    queryFn: () => getEmployeeAvailability(employeeId as string),
    enabled: Boolean(employeeId),
  });
}
