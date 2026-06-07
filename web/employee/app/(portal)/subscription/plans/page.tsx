"use client";

import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useSubscriptionPlans } from "@/hooks/use-subscription";
import { formatPlanPrice } from "@/lib/razorpay";

export default function SubscriptionPlansPage() {
  const { data: plans, isLoading } = useSubscriptionPlans();

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Choose a plan</h1>
        <p className="text-sm text-muted-foreground">
          Compare provider subscription tiers.
        </p>
      </div>

      {isLoading ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-64" />
          ))}
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {(plans ?? []).map((plan) => (
            <Card key={plan.id}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <CardTitle>{plan.name}</CardTitle>
                  {plan.is_featured_allowed && (
                    <Badge variant="secondary">Featured</Badge>
                  )}
                </div>
                <p className="text-2xl font-bold">
                  {formatPlanPrice(plan.price, plan.currency)}
                </p>
                <p className="text-sm text-muted-foreground">
                  {plan.duration_days} days · up to {plan.service_limit} services
                </p>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="space-y-1 text-sm text-muted-foreground">
                  <p>Priority listing: {plan.is_priority_allowed ? "Yes" : "No"}</p>
                  <p>Featured placement: {plan.is_featured_allowed ? "Yes" : "No"}</p>
                </div>
                <Button asChild className="w-full">
                  <Link href={`/subscription/checkout?planId=${plan.id}`}>
                    Subscribe
                  </Link>
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
