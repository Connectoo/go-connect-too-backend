import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type {
  CreateOrderResponse,
  Subscription,
  SubscriptionPlan,
  VerifyPaymentInput,
} from "@/types/subscription";
import type { Payment } from "@/types/payment";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchSubscriptionPlans() {
  return apiRequest<SubscriptionPlan[]>("/subscription-plans", authOptions());
}

export function fetchCurrentSubscription() {
  return apiRequest<Subscription>("/employee/subscriptions/current", authOptions());
}

export function createSubscriptionOrder(planId: string) {
  return apiRequest<CreateOrderResponse>("/employee/subscriptions/create-order", {
    ...authOptions({ method: "POST", body: JSON.stringify({ plan_id: planId }) }),
  });
}

export function verifySubscriptionPayment(body: VerifyPaymentInput) {
  return apiRequest<Payment>("/employee/subscriptions/verify-payment", {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}

export function cancelSubscription(reason?: string) {
  return apiRequest<Subscription>("/employee/subscriptions/cancel", {
    ...authOptions({
      method: "POST",
      body: JSON.stringify(reason ? { reason } : {}),
    }),
  });
}

export function changeSubscriptionPlan(planId: string) {
  return apiRequest<Subscription>("/employee/subscriptions/change-plan", {
    ...authOptions({ method: "POST", body: JSON.stringify({ plan_id: planId }) }),
  });
}

export function setSubscriptionAutoRenew(autoRenew: boolean) {
  return apiRequest<Subscription>("/employee/subscriptions/auto-renew", {
    ...authOptions({
      method: "PATCH",
      body: JSON.stringify({ auto_renew: autoRenew }),
    }),
  });
}
