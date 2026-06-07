"use client";

import { useState } from "react";
import Link from "next/link";
import { DataTable, type Column } from "@/components/admin/data-table";
import { Pagination } from "@/components/admin/pagination";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useAdminSupportTickets } from "@/hooks/use-ops";
import type { SupportTicket } from "@/types/ops";

export default function SupportPage() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState("");
  const limit = 20;

  const { data, isLoading } = useAdminSupportTickets({
    page,
    limit,
    status: status || undefined,
  });

  const columns: Column<SupportTicket>[] = [
    { key: "subject", header: "Subject", cell: (row) => row.subject },
    {
      key: "status",
      header: "Status",
      cell: (row) => <Badge variant="secondary">{row.status}</Badge>,
    },
    {
      key: "priority",
      header: "Priority",
      cell: (row) => <Badge variant="outline">{row.priority}</Badge>,
    },
    {
      key: "updated",
      header: "Updated",
      cell: (row) => new Date(row.updated_at).toLocaleDateString(),
    },
    {
      key: "actions",
      header: "",
      cell: (row) => (
        <Button asChild size="sm" variant="outline">
          <Link href={`/support/${row.id}`}>Open</Link>
        </Button>
      ),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Support</h1>
          <p className="text-sm text-muted-foreground">Customer support tickets.</p>
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
          <option value="open">Open</option>
          <option value="in_progress">In progress</option>
          <option value="resolved">Resolved</option>
          <option value="closed">Closed</option>
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
