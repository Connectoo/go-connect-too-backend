"use client";

import Link from "next/link";
import { DataTable, type Column } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useSupportTickets } from "@/hooks/use-support";
import type { SupportTicket } from "@/types/support";

export default function SupportPage() {
  const { data, isLoading } = useSupportTickets();

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
          <Link href={`/support/${row.id}`}>View</Link>
        </Button>
      ),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Support</h1>
          <p className="text-sm text-muted-foreground">Your help requests.</p>
        </div>
        <Button asChild>
          <Link href="/support/new">New ticket</Link>
        </Button>
      </div>

      <DataTable columns={columns} data={data ?? []} isLoading={isLoading} />
    </div>
  );
}
