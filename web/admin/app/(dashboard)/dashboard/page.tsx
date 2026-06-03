"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useDashboardSummary } from "@/hooks/use-admin";

const statCards = [
  { key: "total_users", label: "Total users" },
  { key: "total_customers", label: "Customers" },
  { key: "total_employees", label: "Employees" },
  { key: "pending_employees", label: "Pending approvals" },
  { key: "total_bookings", label: "Bookings" },
  { key: "active_bookings", label: "Active bookings" },
  { key: "total_services", label: "Services" },
  { key: "total_revenue", label: "Revenue (INR)" },
] as const;

export default function DashboardPage() {
  const { data, isLoading, isError } = useDashboardSummary();

  return (
    <div className="p-6 md:p-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-sm text-muted-foreground">
          Marketplace overview and key metrics.
        </p>
      </div>

      {isError && (
        <p className="mb-4 text-sm text-destructive">
          Failed to load dashboard. Ensure you are signed in as admin and the API is running.
        </p>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {isLoading
          ? Array.from({ length: 8 }).map((_, i) => (
              <Skeleton key={i} className="h-28" />
            ))
          : statCards.map(({ key, label }) => (
              <Card key={key}>
                <CardHeader className="pb-2">
                  <CardTitle className="text-sm font-medium text-muted-foreground">
                    {label}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-2xl font-bold">
                    {key === "total_revenue"
                      ? `₹${(data?.[key] ?? 0).toLocaleString("en-IN")}`
                      : (data?.[key] ?? 0)}
                  </p>
                </CardContent>
              </Card>
            ))}
      </div>
    </div>
  );
}
