"use client";

import { useState } from "react";
import Link from "next/link";
import { DataTable, type Column } from "@/components/admin/data-table";
import { Pagination } from "@/components/admin/pagination";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useAdminBookings } from "@/hooks/use-admin";
import type { Booking } from "@/types/admin";

export default function BookingsPage() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState("");
  const limit = 20;

  const { data, isLoading } = useAdminBookings({
    page,
    limit,
    status: status || undefined,
  });

  const columns: Column<Booking>[] = [
    {
      key: "date",
      header: "Date",
      cell: (row) => `${row.booking_date} ${row.start_time}`,
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
      key: "ids",
      header: "Refs",
      cell: (row) => (
        <span className="text-xs text-muted-foreground">
          S:{row.service_id.slice(0, 8)}…
        </span>
      ),
    },
    {
      key: "actions",
      header: "",
      cell: (row) => (
        <Button asChild size="sm" variant="outline">
          <Link href={`/bookings/${row.id}`}>View</Link>
        </Button>
      ),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Bookings</h1>
          <p className="text-sm text-muted-foreground">All marketplace bookings.</p>
        </div>
        <select
          className="h-10 rounded-md border border-input bg-background px-3 text-sm"
          value={status}
          onChange={(e) => {
            setStatus(e.target.value);
            setPage(1);
          }}
        >
          <option value="">All statuses</option>
          <option value="pending">Pending</option>
          <option value="accepted">Accepted</option>
          <option value="in_progress">In progress</option>
          <option value="completed">Completed</option>
          <option value="cancelled">Cancelled</option>
          <option value="rejected">Rejected</option>
        </select>
      </div>

      <DataTable columns={columns} data={data?.items ?? []} isLoading={isLoading} />

      {data && (
        <div className="mt-4">
          <Pagination
            page={data.page}
            limit={data.limit}
            total={data.total}
            onPageChange={setPage}
          />
        </div>
      )}
    </div>
  );
}
