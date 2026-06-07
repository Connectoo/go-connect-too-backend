"use client";

import { useState } from "react";
import { DataTable, type Column } from "@/components/admin/data-table";
import { Pagination } from "@/components/admin/pagination";
import { Badge } from "@/components/ui/badge";
import { useAdminSubscriptions } from "@/hooks/use-ops";
import type { Subscription } from "@/types/ops";

export default function SubscriptionsPage() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState("");
  const limit = 20;

  const { data, isLoading } = useAdminSubscriptions({
    page,
    limit,
    status: status || undefined,
  });

  const columns: Column<Subscription>[] = [
    {
      key: "plan",
      header: "Plan",
      cell: (row) => row.plan_name,
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => <Badge variant="secondary">{row.status}</Badge>,
    },
    {
      key: "employee",
      header: "Employee",
      cell: (row) => (
        <span className="text-xs text-muted-foreground">
          {row.employee_id.slice(0, 8)}…
        </span>
      ),
    },
    {
      key: "expires",
      header: "Expires",
      cell: (row) =>
        row.expires_at ? new Date(row.expires_at).toLocaleDateString() : "—",
    },
    {
      key: "auto",
      header: "Auto-renew",
      cell: (row) => (row.auto_renew ? "Yes" : "No"),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Subscriptions</h1>
          <p className="text-sm text-muted-foreground">Provider subscription records.</p>
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
          <option value="active">Active</option>
          <option value="expired">Expired</option>
          <option value="cancelled">Cancelled</option>
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
