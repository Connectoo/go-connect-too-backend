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
  useEmployeeBookingsAnalytics,
  useEmployeeReviewsAnalytics,
} from "@/hooks/use-analytics";

export default function AnalyticsPage() {
  const { data: bookings, isLoading: bookingsLoading } =
    useEmployeeBookingsAnalytics();
  const { data: reviews, isLoading: reviewsLoading } =
    useEmployeeReviewsAnalytics();

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Analytics</h1>
        <p className="text-sm text-muted-foreground">
          Booking volume and review performance.
        </p>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
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

        <Card>
          <CardHeader>
            <CardTitle>Daily booking volume</CardTitle>
          </CardHeader>
          <CardContent className="h-72">
            {bookingsLoading ? (
              <Skeleton className="h-full w-full" />
            ) : (
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={bookings?.daily_volume ?? []}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" tickFormatter={(v) => v.slice(5)} />
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
            <CardTitle>Review summary</CardTitle>
          </CardHeader>
          <CardContent>
            {reviewsLoading ? (
              <Skeleton className="h-20 w-full" />
            ) : (
              <div className="grid gap-4 sm:grid-cols-3">
                <div>
                  <p className="text-sm text-muted-foreground">Total reviews</p>
                  <p className="text-2xl font-bold">{reviews?.total_reviews ?? 0}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Average rating</p>
                  <p className="text-2xl font-bold">
                    {reviews?.average_rating?.toFixed(1) ?? "—"}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Period</p>
                  <p className="text-sm font-medium">
                    {reviews?.period.from} → {reviews?.period.to}
                  </p>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
