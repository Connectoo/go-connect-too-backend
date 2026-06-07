import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type { CreateTicketInput, SupportTicket } from "@/types/support";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchSupportTickets() {
  return apiRequest<SupportTicket[]>("/support/tickets", authOptions());
}

export function createSupportTicket(body: CreateTicketInput) {
  return apiRequest<SupportTicket>("/support/tickets", {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}
