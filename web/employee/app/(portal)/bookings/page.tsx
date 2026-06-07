"use client";

import Link from "next/link";
import { useMemo, useState } from "react";
import { DataTable, type Column } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { useBookings } from "@/hooks/use-bookings";
import { type Booking, type BookingTab, tabForStatus } from "@/types/booking";

const TABS: { key: BookingTab; label: string }[] = [
  { key: "pending", label: "Pending" },
  { key: "active", label: "Active" },
  { key: "completed", label: "Completed" },
];

export default function BookingsPage() {
  const [tab, setTab] = useState<BookingTab>("pending");
  const { data, isLoading, isError } = useBookings();

  const grouped = useMemo(() => {
    const result: Record<BookingTab, Booking[]> = {
      pending: [],
      active: [],
      completed: [],
    };
    (data ?? []).forEach((booking) => {
      result[tabForStatus(booking.status)].push(booking);
    });
    return result;
  }, [data]);

  const columns: Column<Booking>[] = [
    {
      key: "date",
      header: "Date & time",
      cell: (row) => (
        <Link
          href={`/bookings/${row.id}`}
          className="font-medium text-primary hover:underline"
        >
          {row.booking_date} · {row.start_time}–{row.end_time}
        </Link>
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => <Badge variant="secondary">{row.status}</Badge>,
    },
    {
      key: "amount",
      header: "Amount",
      cell: (row) => `₹${row.total_amount.toLocaleString("en-IN")}`,
    },
    {
      key: "view",
      header: "",
      className: "text-right",
      cell: (row) => (
        <Button asChild variant="outline" size="sm">
          <Link href={`/bookings/${row.id}`}>View</Link>
        </Button>
      ),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Bookings</h1>
        <p className="text-sm text-muted-foreground">
          Manage incoming and active booking requests.
        </p>
      </div>

      {isError && (
        <p className="mb-4 text-sm text-destructive">
          Failed to load bookings. Ensure you are signed in and the API is running.
        </p>
      )}

      <div className="mb-4 flex gap-2 border-b">
        {TABS.map(({ key, label }) => (
          <button
            key={key}
            type="button"
            onClick={() => setTab(key)}
            className={cn(
              "-mb-px border-b-2 px-4 py-2 text-sm font-medium transition-colors",
              tab === key
                ? "border-primary text-foreground"
                : "border-transparent text-muted-foreground hover:text-foreground",
            )}
          >
            {label}
            <span className="ml-2 text-xs text-muted-foreground">
              {grouped[key].length}
            </span>
          </button>
        ))}
      </div>

      <DataTable
        columns={columns}
        data={grouped[tab]}
        isLoading={isLoading}
        emptyMessage="No bookings in this tab."
      />
    </div>
  );
}
