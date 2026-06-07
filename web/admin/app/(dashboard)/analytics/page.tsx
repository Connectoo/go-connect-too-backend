"use client";

import {
  Bar,
  BarChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useAdminAnalyticsSummary,
  useAdminBookingsAnalytics,
  useAdminCategoriesAnalytics,
  useAdminRevenueAnalytics,
} from "@/hooks/use-ops";

function formatInrPaise(amount: number) {
  return `₹${(amount / 100).toLocaleString("en-IN")}`;
}

export default function AnalyticsPage() {
  const { data: summary, isLoading: summaryLoading } = useAdminAnalyticsSummary();
  const { data: revenue, isLoading: revenueLoading } = useAdminRevenueAnalytics();
  const { data: bookings, isLoading: bookingsLoading } = useAdminBookingsAnalytics();
  const { data: categories, isLoading: categoriesLoading } =
    useAdminCategoriesAnalytics();

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Analytics</h1>
        <p className="text-sm text-muted-foreground">Platform KPIs and trends.</p>
      </div>

      {summaryLoading ? (
        <Skeleton className="mb-6 h-24 w-full" />
      ) : (
        <div className="mb-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <Card>
            <CardContent className="pt-6">
              <p className="text-sm text-muted-foreground">Users</p>
              <p className="text-2xl font-bold">{summary?.total_users ?? 0}</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <p className="text-sm text-muted-foreground">Active subscriptions</p>
              <p className="text-2xl font-bold">{summary?.active_subscriptions ?? 0}</p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <p className="text-sm text-muted-foreground">MRR</p>
              <p className="text-2xl font-bold">
                {formatInrPaise(summary?.monthly_recurring_revenue ?? 0)}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardContent className="pt-6">
              <p className="text-sm text-muted-foreground">Booking volume</p>
              <p className="text-2xl font-bold">{summary?.booking_volume ?? 0}</p>
            </CardContent>
          </Card>
        </div>
      )}

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Subscription revenue (daily)</CardTitle>
          </CardHeader>
          <CardContent className="h-72">
            {revenueLoading ? (
              <Skeleton className="h-full w-full" />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={revenue?.daily_subscription_revenue ?? []}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" tickFormatter={(v) => v.slice(5)} />
                  <YAxis tickFormatter={(v) => `₹${v / 100}`} />
                  <Tooltip formatter={(v) => formatInrPaise(Number(v))} />
                  <Bar dataKey="amount" fill="hsl(var(--primary))" />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Bookings by status</CardTitle>
          </CardHeader>
          <CardContent className="h-72">
            {bookingsLoading ? (
              <Skeleton className="h-full w-full" />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={bookings?.by_status ?? []}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="status" />
                  <YAxis allowDecimals={false} />
                  <Tooltip />
                  <Bar dataKey="count" fill="hsl(var(--primary))" />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>

        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Top categories</CardTitle>
          </CardHeader>
          <CardContent className="h-72">
            {categoriesLoading ? (
              <Skeleton className="h-full w-full" />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={categories?.categories ?? []}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="category_name" />
                  <YAxis allowDecimals={false} />
                  <Tooltip />
                  <Bar dataKey="booking_count" fill="hsl(var(--primary))" />
                </BarChart>
              </ResponsiveContainer>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
