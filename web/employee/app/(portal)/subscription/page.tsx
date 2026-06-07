"use client";

import Link from "next/link";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useCurrentSubscription } from "@/hooks/use-subscription";
import { formatPlanPrice } from "@/lib/razorpay";
import { ApiError } from "@/lib/api-client";

export default function SubscriptionPage() {
  const { data, isLoading, error } = useCurrentSubscription();
  const noSubscription = error instanceof ApiError && error.status === 404;

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Subscription</h1>
          <p className="text-sm text-muted-foreground">
            Your provider plan and billing status.
          </p>
        </div>
        <div className="flex gap-2">
          <Button asChild variant="outline">
            <Link href="/subscription/plans">View plans</Link>
          </Button>
          {data && (
            <Button asChild>
              <Link href="/subscription/manage">Manage</Link>
            </Button>
          )}
        </div>
      </div>

      {isLoading ? (
        <Skeleton className="h-48 w-full max-w-xl" />
      ) : noSubscription ? (
        <Card className="max-w-xl">
          <CardHeader>
            <CardTitle>No active subscription</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <p className="text-sm text-muted-foreground">
              Choose a plan to list services and receive bookings on the marketplace.
            </p>
            <Button asChild>
              <Link href="/subscription/plans">Browse plans</Link>
            </Button>
          </CardContent>
        </Card>
      ) : data ? (
        <Card className="max-w-xl">
          <CardHeader className="flex flex-row items-center justify-between">
            <CardTitle>{data.plan_name}</CardTitle>
            <Badge variant="secondary">{data.status}</Badge>
          </CardHeader>
          <CardContent className="space-y-3 text-sm">
            {data.plan && (
              <p className="text-lg font-semibold">
                {formatPlanPrice(data.plan.price, data.plan.currency)}
                <span className="text-sm font-normal text-muted-foreground">
                  {" "}
                  / {data.plan.duration_days} days
                </span>
              </p>
            )}
            {data.starts_at && (
              <p>
                <span className="text-muted-foreground">Started:</span>{" "}
                {new Date(data.starts_at).toLocaleDateString()}
              </p>
            )}
            {data.expires_at && (
              <p>
                <span className="text-muted-foreground">Expires:</span>{" "}
                {new Date(data.expires_at).toLocaleDateString()}
              </p>
            )}
            <p>
              <span className="text-muted-foreground">Auto-renew:</span>{" "}
              {data.auto_renew ? "On" : "Off"}
            </p>
          </CardContent>
        </Card>
      ) : (
        <p className="text-sm text-destructive">Failed to load subscription.</p>
      )}
    </div>
  );
}
