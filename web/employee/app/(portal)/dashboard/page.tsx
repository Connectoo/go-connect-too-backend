"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useEmployeeSummary } from "@/hooks/use-analytics";

function formatResponseTime(ms: number | null): string {
  if (ms === null || ms === undefined) return "—";
  const minutes = Math.round(ms / 60000);
  if (minutes < 60) return `${minutes}m`;
  const hours = Math.round(minutes / 60);
  return `${hours}h`;
}

export default function DashboardPage() {
  const { data, isLoading, isError } = useEmployeeSummary();

  const cards = [
    { label: "Profile views", value: data?.profile_views ?? 0 },
    { label: "Total bookings", value: data?.total_bookings ?? 0 },
    { label: "Completed", value: data?.completed_bookings ?? 0 },
    { label: "Cancelled", value: data?.cancelled_bookings ?? 0 },
    {
      label: "Est. revenue (INR)",
      value: `₹${data?.estimated_revenue ?? "0"}`,
    },
    {
      label: "Avg. response time",
      value: formatResponseTime(data?.average_response_time_ms ?? null),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-sm text-muted-foreground">
          Your provider KPIs and recent performance.
        </p>
      </div>

      {isError && (
        <p className="mb-4 text-sm text-destructive">
          Failed to load dashboard. Ensure you are signed in and the API is running.
        </p>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {isLoading
          ? Array.from({ length: 6 }).map((_, i) => (
              <Skeleton key={i} className="h-28" />
            ))
          : cards.map((card) => (
              <Card key={card.label}>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">
                    {card.label}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold">{card.value}</p>
                </CardContent>
              </Card>
            ))}
      </div>
    </div>
  );
}
