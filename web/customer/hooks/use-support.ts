"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createSupportTicket, fetchSupportTickets } from "@/services/support";
import type { CreateTicketInput } from "@/types/support";

export function useSupportTickets() {
  return useQuery({
    queryKey: ["customer", "support"],
    queryFn: fetchSupportTickets,
  });
}

export function useCreateSupportTicket() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateTicketInput) => createSupportTicket(body),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["customer", "support"] }),
  });
}
