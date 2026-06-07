"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useSubscriptionMutations, useSubscriptionPlans } from "@/hooks/use-subscription";
import { formatPlanPrice, openRazorpayCheckout } from "@/lib/razorpay";

export function CheckoutContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const planId = searchParams.get("planId") ?? "";
  const { data: plans, isLoading } = useSubscriptionPlans();
  const mutations = useSubscriptionMutations();
  const [error, setError] = useState<string | null>(null);
  const [paying, setPaying] = useState(false);

  const plan = plans?.find((p) => p.id === planId);

  useEffect(() => {
    if (!isLoading && planId && !plan) {
      setError("Plan not found. Choose a plan from the list.");
    }
  }, [isLoading, planId, plan]);

  async function startCheckout() {
    if (!plan) return;
    setError(null);
    setPaying(true);
    try {
      const order = await mutations.createOrder.mutateAsync(plan.id);
      if (!order.razorpay_key_id) {
        setError("Payment gateway is not configured. Contact support.");
        setPaying(false);
        return;
      }

      await openRazorpayCheckout({
        keyId: order.razorpay_key_id,
        amount: order.amount,
        currency: order.currency,
        orderId: order.provider_order_id,
        planName: plan.name,
        onDismiss: () => setPaying(false),
        onSuccess: async (response) => {
          try {
            await mutations.verifyPayment.mutateAsync({
              payment_id: order.payment_id,
              provider_order_id: response.razorpay_order_id,
              provider_payment_id: response.razorpay_payment_id,
              signature: response.razorpay_signature,
            });
            router.push("/subscription");
          } catch (err) {
            setError(err instanceof Error ? err.message : "Payment verification failed");
            setPaying(false);
          }
        },
      });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not start checkout");
      setPaying(false);
    }
  }

  if (!planId) {
    return (
      <div>
        <p className="text-sm text-muted-foreground">No plan selected.</p>
        <Button asChild className="mt-4">
          <Link href="/subscription/plans">View plans</Link>
        </Button>
      </div>
    );
  }

  if (isLoading) {
    return <Skeleton className="h-48 max-w-md" />;
  }

  if (plan) {
    return (
      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>{plan.name}</CardTitle>
          <p className="text-2xl font-bold">
            {formatPlanPrice(plan.price, plan.currency)}
          </p>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            {plan.duration_days} days · up to {plan.service_limit} services
          </p>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <div className="flex gap-2">
            <Button disabled={paying} onClick={() => void startCheckout()}>
              {paying ? "Opening Razorpay…" : "Pay with Razorpay"}
            </Button>
            <Button asChild variant="outline">
              <Link href="/subscription/plans">Back</Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    );
  }

  return (
    <div>
      <p className="text-sm text-destructive">{error ?? "Plan not found."}</p>
      <Button asChild className="mt-4" variant="outline">
        <Link href="/subscription/plans">View plans</Link>
      </Button>
    </div>
  );
}
