"use client";

import { useQuery } from "@tanstack/react-query";
import { fetchPayments } from "@/services/payment";

export function usePayments() {
  return useQuery({
    queryKey: ["employee", "payments"],
    queryFn: fetchPayments,
  });
}
