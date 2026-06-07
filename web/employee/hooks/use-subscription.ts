"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ApiError } from "@/lib/api-client";
import {
  cancelSubscription,
  changeSubscriptionPlan,
  createSubscriptionOrder,
  fetchCurrentSubscription,
  fetchSubscriptionPlans,
  setSubscriptionAutoRenew,
  verifySubscriptionPayment,
} from "@/services/subscription";
import type { VerifyPaymentInput } from "@/types/subscription";

export function useSubscriptionPlans() {
  return useQuery({
    queryKey: ["employee", "subscription-plans"],
    queryFn: fetchSubscriptionPlans,
  });
}

export function useCurrentSubscription() {
  return useQuery({
    queryKey: ["employee", "subscription", "current"],
    queryFn: fetchCurrentSubscription,
    retry: (_count, error) => !(error instanceof ApiError && error.status === 404),
  });
}

export function useSubscriptionMutations() {
  const qc = useQueryClient();
  const invalidate = () => {
    qc.invalidateQueries({ queryKey: ["employee", "subscription"] });
    qc.invalidateQueries({ queryKey: ["employee", "payments"] });
  };

  return {
    createOrder: useMutation({
      mutationFn: (planId: string) => createSubscriptionOrder(planId),
    }),
    verifyPayment: useMutation({
      mutationFn: (body: VerifyPaymentInput) => verifySubscriptionPayment(body),
      onSuccess: invalidate,
    }),
    cancel: useMutation({
      mutationFn: (reason?: string) => cancelSubscription(reason),
      onSuccess: invalidate,
    }),
    changePlan: useMutation({
      mutationFn: (planId: string) => changeSubscriptionPlan(planId),
      onSuccess: invalidate,
    }),
    setAutoRenew: useMutation({
      mutationFn: (autoRenew: boolean) => setSubscriptionAutoRenew(autoRenew),
      onSuccess: invalidate,
    }),
  };
}
