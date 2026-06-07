"use client";

import { useState } from "react";
import { DataTable, type Column } from "@/components/admin/data-table";
import { Pagination } from "@/components/admin/pagination";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useAdminReports, useReportActions } from "@/hooks/use-ops";
import type { Report } from "@/types/ops";

export default function ReportsPage() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState("open");
  const limit = 20;

  const { data, isLoading } = useAdminReports({
    page,
    limit,
    status: status || undefined,
  });
  const actions = useReportActions();

  const columns: Column<Report>[] = [
    {
      key: "reason",
      header: "Reason",
      cell: (row) => row.reason,
    },
    {
      key: "description",
      header: "Description",
      cell: (row) => (
        <span className="line-clamp-2 max-w-xs text-sm">
          {row.description ?? "—"}
        </span>
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => <Badge variant="secondary">{row.status}</Badge>,
    },
    {
      key: "date",
      header: "Reported",
      cell: (row) => new Date(row.created_at).toLocaleDateString(),
    },
    {
      key: "actions",
      header: "",
      cell: (row) =>
        row.status === "open" ? (
          <Button size="sm" onClick={() => actions.resolve.mutate(row.id)}>
            Resolve
          </Button>
        ) : null,
    },
  ];

  async function handleExport() {
    const blob = await actions.exportCsv.mutateAsync();
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = "reports.csv";
    a.click();
    URL.revokeObjectURL(url);
  }

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Reports</h1>
          <p className="text-sm text-muted-foreground">User and content reports queue.</p>
        </div>
        <div className="flex gap-2">
          <select
            className="h-10 rounded-md border border-input bg-background px-3 text-sm"
            value={status}
            onChange={(e) => {
              setStatus(e.target.value);
              setPage(1);
            }}
          >
            <option value="">All</option>
            <option value="open">Open</option>
            <option value="resolved">Resolved</option>
          </select>
          <Button
            variant="outline"
            disabled={actions.exportCsv.isPending}
            onClick={() => void handleExport()}
          >
            Export CSV
          </Button>
        </div>
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
