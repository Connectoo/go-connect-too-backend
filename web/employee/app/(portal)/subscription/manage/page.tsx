"use client";

import { useState } from "react";
import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useCurrentSubscription,
  useSubscriptionMutations,
  useSubscriptionPlans,
} from "@/hooks/use-subscription";
import { ApiError } from "@/lib/api-client";

export default function SubscriptionManagePage() {
  const { data: subscription, isLoading, error } = useCurrentSubscription();
  const { data: plans } = useSubscriptionPlans();
  const mutations = useSubscriptionMutations();
  const [cancelReason, setCancelReason] = useState("");
  const [selectedPlanId, setSelectedPlanId] = useState("");

  const noSubscription = error instanceof ApiError && error.status === 404;

  if (isLoading) {
    return (
      <div className="p-6 md:p-8">
        <Skeleton className="h-64 max-w-xl" />
      </div>
    );
  }

  if (noSubscription) {
    return (
      <div className="p-6 md:p-8">
        <p className="text-sm text-muted-foreground">No subscription to manage.</p>
        <Button asChild className="mt-4">
          <Link href="/subscription/plans">Choose a plan</Link>
        </Button>
      </div>
    );
  }

  if (!subscription) {
    return (
      <div className="p-6 md:p-8">
        <p className="text-sm text-destructive">Failed to load subscription.</p>
      </div>
    );
  }

  const otherPlans = (plans ?? []).filter((p) => p.id !== subscription.plan_id);

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Manage subscription</h1>
        <p className="text-sm text-muted-foreground">
          Cancel, change plan, or toggle auto-renew.
        </p>
      </div>

      <div className="grid max-w-2xl gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>{subscription.plan_name}</CardTitle>
            <Badge variant="secondary">{subscription.status}</Badge>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between rounded-md border p-3">
              <div>
                <p className="font-medium">Auto-renew</p>
                <p className="text-xs text-muted-foreground">
                  Renew automatically when the plan expires.
                </p>
              </div>
              <Button
                variant="outline"
                size="sm"
                disabled={mutations.setAutoRenew.isPending}
                onClick={() =>
                  mutations.setAutoRenew.mutate(!subscription.auto_renew)
                }
              >
                {subscription.auto_renew ? "Turn off" : "Turn on"}
              </Button>
            </div>

            {otherPlans.length > 0 && subscription.status === "active" && (
              <div className="space-y-2">
                <Label htmlFor="change-plan">Change plan</Label>
                <div className="flex gap-2">
                  <select
                    id="change-plan"
                    className="h-10 flex-1 rounded-md border border-input bg-background px-3 text-sm"
                    value={selectedPlanId}
                    onChange={(e) => setSelectedPlanId(e.target.value)}
                  >
                    <option value="">Select a plan</option>
                    {otherPlans.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.name}
                      </option>
                    ))}
                  </select>
                  <Button
                    disabled={!selectedPlanId || mutations.changePlan.isPending}
                    onClick={() => {
                      if (selectedPlanId) {
                        mutations.changePlan.mutate(selectedPlanId);
                      }
                    }}
                  >
                    Change
                  </Button>
                </div>
                <p className="text-xs text-muted-foreground">
                  Plan changes may require a new payment at checkout.
                </p>
              </div>
            )}
          </CardContent>
        </Card>

        {subscription.status === "active" && (
          <Card>
            <CardHeader>
              <CardTitle className="text-destructive">Cancel subscription</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3">
              <div className="space-y-2">
                <Label htmlFor="reason">Reason (optional)</Label>
                <Input
                  id="reason"
                  value={cancelReason}
                  onChange={(e) => setCancelReason(e.target.value)}
                  placeholder="Why are you cancelling?"
                />
              </div>
              <Button
                variant="destructive"
                disabled={mutations.cancel.isPending}
                onClick={() =>
                  mutations.cancel.mutate(cancelReason || undefined)
                }
              >
                Cancel subscription
              </Button>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
