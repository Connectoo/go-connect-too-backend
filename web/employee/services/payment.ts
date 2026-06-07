import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type { Payment } from "@/types/payment";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchPayments() {
  return apiRequest<Payment[]>("/employee/payments", authOptions());
}
