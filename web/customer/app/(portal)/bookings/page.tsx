"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { formatCurrency } from "@/lib/utils";
import { useBookings } from "@/hooks/use-bookings";
import type { Booking, BookingStatus } from "@/types/booking";

type Tab = "upcoming" | "past" | "cancelled";

const TAB_STATUSES: Record<Tab, BookingStatus[]> = {
  upcoming: ["pending", "accepted", "in_progress"],
  past: ["completed", "no_show"],
  cancelled: ["cancelled", "rejected"],
};

const TABS: { key: Tab; label: string }[] = [
  { key: "upcoming", label: "Upcoming" },
  { key: "past", label: "Past" },
  { key: "cancelled", label: "Cancelled" },
];

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
    weekday: "short",
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}

function BookingCard({ booking }: { booking: Booking }) {
  const badge = statusBadge(booking.status);
  return (
    <Link href={`/bookings/${booking.id}`}>
      <Card className="transition-colors hover:bg-muted/40">
        <CardContent className="flex items-center justify-between gap-4 p-4">
          <div className="space-y-1">
            <p className="font-medium">{formatDate(booking.booking_date)}</p>
            <p className="text-sm text-muted-foreground">
              {booking.start_time} – {booking.end_time}
            </p>
          </div>
          <div className="flex flex-col items-end gap-2">
            <Badge variant={badge.variant} className={badge.className}>
              {formatStatus(booking.status)}
            </Badge>
            <span className="text-sm font-medium">
              {formatCurrency(booking.total_amount)}
            </span>
          </div>
        </CardContent>
      </Card>
    </Link>
  );
}

export default function BookingsPage() {
  const { data, isLoading, isError, error } = useBookings();
  const [tab, setTab] = useState<Tab>("upcoming");

  const grouped = useMemo(() => {
    const bookings = data ?? [];
    return {
      upcoming: bookings.filter((b) => TAB_STATUSES.upcoming.includes(b.status)),
      past: bookings.filter((b) => TAB_STATUSES.past.includes(b.status)),
      cancelled: bookings.filter((b) => TAB_STATUSES.cancelled.includes(b.status)),
    };
  }, [data]);

  const visible = grouped[tab];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">My bookings</h1>
        <p className="text-sm text-muted-foreground">
          Track and manage your appointments.
        </p>
      </div>

      <div className="mb-6 flex gap-1 border-b">
        {TABS.map(({ key, label }) => (
          <button
            key={key}
            type="button"
            onClick={() => setTab(key)}
            className={
              "border-b-2 px-4 py-2 text-sm font-medium transition-colors " +
              (tab === key
                ? "border-primary text-foreground"
                : "border-transparent text-muted-foreground hover:text-foreground")
            }
          >
            {label}
            {!isLoading ? ` (${grouped[key].length})` : ""}
          </button>
        ))}
      </div>

      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <Skeleton key={i} className="h-20 w-full" />
          ))}
        </div>
      ) : isError ? (
        <div className="rounded-lg border p-8 text-center text-sm text-destructive">
          {(error as Error)?.message || "Failed to load bookings."}
        </div>
      ) : visible.length === 0 ? (
        <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">
          No {tab} bookings.
        </div>
      ) : (
        <div className="space-y-3">
          {visible.map((booking) => (
            <BookingCard key={booking.id} booking={booking} />
          ))}
        </div>
      )}
    </div>
  );
}
