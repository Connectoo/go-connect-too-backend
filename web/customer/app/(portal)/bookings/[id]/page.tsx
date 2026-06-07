"use client";

import { ArrowLeft, Check } from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useBooking } from "@/hooks/use-bookings";
import { useService } from "@/hooks/use-catalog";
import { cn, formatCurrency } from "@/lib/utils";
import type { BookingStatus } from "@/types/booking";

const FLOW: BookingStatus[] = ["pending", "accepted", "in_progress", "completed"];
const CANCELLABLE: BookingStatus[] = ["pending", "accepted"];
const RESCHEDULABLE: BookingStatus[] = ["pending", "accepted"];

function formatStatus(status: BookingStatus) {
  return status.replace(/_/g, " ").replace(/^\w/, (c) => c.toUpperCase());
}

function statusBadge(status: BookingStatus): {
  variant: "default" | "secondary" | "outline";
  className?: string;
} {
  switch (status) {
    case "pending":
      return { variant: "secondary" };
    case "accepted":
      return { variant: "default" };
    case "in_progress":
      return { variant: "default", className: "bg-blue-600" };
    case "completed":
      return { variant: "default", className: "bg-green-600" };
    default:
      return { variant: "outline", className: "border-destructive text-destructive" };
  }
}

function formatDate(value: string) {
  const [y, m, d] = value.split("-").map(Number);
  if (!y || !m || !d) return value;
  return new Date(y, m - 1, d).toLocaleDateString(undefined, {
    weekday: "long",
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

function StatusTimeline({ status }: { status: BookingStatus }) {
  const inFlow = FLOW.indexOf(status);
  // For terminal off-flow states (cancelled/rejected/no_show) we don't have
  // history from the API, so show a simple two-step timeline.
  const steps: { label: string; done: boolean; current: boolean }[] =
    inFlow >= 0
      ? FLOW.map((s, i) => ({
          label: formatStatus(s),
          done: i < inFlow,
          current: i === inFlow,
        }))
      : [
          { label: "Pending", done: true, current: false },
          { label: formatStatus(status), done: false, current: true },
        ];

  return (
    <ol className="space-y-4">
      {steps.map((step, i) => (
        <li key={i} className="flex items-center gap-3">
          <span
            className={cn(
              "flex h-6 w-6 items-center justify-center rounded-full border text-xs",
              step.done && "border-green-600 bg-green-600 text-white",
              step.current && "border-primary bg-primary text-primary-foreground",
              !step.done && !step.current && "border-muted-foreground/30 text-muted-foreground",
            )}
          >
            {step.done ? <Check className="h-3 w-3" /> : i + 1}
          </span>
          <span
            className={cn(
              "text-sm",
              step.current ? "font-medium text-foreground" : "text-muted-foreground",
            )}
          >
            {step.label}
          </span>
        </li>
      ))}
    </ol>
  );
}

export default function BookingDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const { data: booking, isLoading, isError, error } = useBooking(id);
  const { data: service } = useService(booking?.service_id ?? "");

  if (isLoading) {
    return (
      <div className="space-y-4 p-6 md:p-8">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-40 w-full" />
        <Skeleton className="h-40 w-full" />
      </div>
    );
  }

  if (isError || !booking) {
    return (
      <div className="p-6 md:p-8">
        <Link
          href="/bookings"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Back to bookings
        </Link>
        <div className="rounded-lg border p-8 text-center text-sm text-destructive">
          {(error as Error)?.message || "Booking not found."}
        </div>
      </div>
    );
  }

  const badge = statusBadge(booking.status);
  const canCancel = CANCELLABLE.includes(booking.status);
  const canReschedule = RESCHEDULABLE.includes(booking.status);

  return (
    <div className="p-6 md:p-8">
      <Link
        href="/bookings"
        className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" /> Back to bookings
      </Link>

      <div className="mb-6 flex items-start justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">
            {service?.title ?? "Booking"}
          </h1>
          <p className="text-sm text-muted-foreground">
            {formatDate(booking.booking_date)}
          </p>
        </div>
        <Badge variant={badge.variant} className={badge.className}>
          {formatStatus(booking.status)}
        </Badge>
      </div>

      <div className="grid gap-6 md:grid-cols-3">
        <div className="space-y-6 md:col-span-2">
          <Card>
            <CardHeader>
              <CardTitle>Details</CardTitle>
            </CardHeader>
            <CardContent className="space-y-3 text-sm">
              {service ? (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Service</span>
                  <span className="font-medium">{service.title}</span>
                </div>
              ) : null}
              <div className="flex justify-between">
                <span className="text-muted-foreground">Date</span>
                <span className="font-medium">{formatDate(booking.booking_date)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Time</span>
                <span className="font-medium">
                  {booking.start_time} – {booking.end_time}
                </span>
              </div>
              {service ? (
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Duration</span>
                  <span className="font-medium">{service.duration_minutes} min</span>
                </div>
              ) : null}
              <div className="flex justify-between">
                <span className="text-muted-foreground">Amount</span>
                <span className="font-medium">{formatCurrency(booking.total_amount)}</span>
              </div>
            </CardContent>
          </Card>

          {booking.customer_notes ? (
            <Card>
              <CardHeader>
                <CardTitle>Your notes</CardTitle>
              </CardHeader>
              <CardContent className="text-sm text-muted-foreground">
                {booking.customer_notes}
              </CardContent>
            </Card>
          ) : null}

          {(canCancel || canReschedule) && (
            <div className="flex flex-wrap gap-3">
              {canReschedule ? (
                <Button asChild variant="outline">
                  <Link href={`/bookings/${booking.id}/reschedule`}>Reschedule</Link>
                </Button>
              ) : null}
              {canCancel ? (
                <Button asChild variant="destructive">
                  <Link href={`/bookings/${booking.id}/cancel`}>Cancel booking</Link>
                </Button>
              ) : null}
            </div>
          )}
        </div>

        <Card>
          <CardHeader>
            <CardTitle>Status</CardTitle>
          </CardHeader>
          <CardContent>
            <StatusTimeline status={booking.status} />
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
