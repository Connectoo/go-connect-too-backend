"use client";

import { useState } from "react";
import { DataTable, type Column } from "@/components/admin/data-table";
import { Pagination } from "@/components/admin/pagination";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAdminKYC, useKYCActions } from "@/hooks/use-admin";
import type { AdminKYCRecord } from "@/types/admin";

export default function KYCPage() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState("pending");
  const [rejectTarget, setRejectTarget] = useState<AdminKYCRecord | null>(null);
  const [rejectReason, setRejectReason] = useState("");

  const limit = 20;
  const { data, isLoading } = useAdminKYC({
    page,
    limit,
    status: status || undefined,
  });
  const actions = useKYCActions();

  const columns: Column<AdminKYCRecord>[] = [
    {
      key: "provider",
      header: "Provider",
      cell: (row) => (
        <div>
          <p className="font-medium">{row.employee_display_name ?? row.user_name}</p>
          <p className="text-xs text-muted-foreground">{row.user_email}</p>
        </div>
      ),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => <Badge variant="secondary">{row.status}</Badge>,
    },
    {
      key: "submitted",
      header: "Submitted",
      cell: (row) => new Date(row.created_at).toLocaleDateString(),
    },
    {
      key: "docs",
      header: "Documents",
      cell: (row) => (
        <div className="flex gap-2 text-xs">
          <a
            href={row.id_proof_url}
            target="_blank"
            rel="noreferrer"
            className="text-primary underline"
          >
            ID
          </a>
          <a
            href={row.address_proof_url}
            target="_blank"
            rel="noreferrer"
            className="text-primary underline"
          >
            Address
          </a>
        </div>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      cell: (row) =>
        row.status === "pending" ? (
          <div className="flex gap-2">
            <Button size="sm" onClick={() => actions.approve.mutate(row.id)}>
              Approve
            </Button>
            <Button size="sm" variant="outline" onClick={() => setRejectTarget(row)}>
              Reject
            </Button>
          </div>
        ) : (
          <span className="text-xs text-muted-foreground">—</span>
        ),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">KYC review</h1>
        <p className="text-sm text-muted-foreground">
          Review identity documents submitted by providers.
        </p>
      </div>

      <div className="mb-4">
        <select
          className="h-10 rounded-md border border-input bg-background px-3 text-sm"
          value={status}
          onChange={(e) => {
            setStatus(e.target.value);
            setPage(1);
          }}
        >
          <option value="pending">Pending</option>
          <option value="approved">Approved</option>
          <option value="rejected">Rejected</option>
          <option value="">All</option>
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

      {rejectTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
          <div className="w-full max-w-md rounded-lg border bg-background p-6 shadow-lg">
            <h2 className="text-lg font-semibold">Reject KYC</h2>
            <p className="mt-2 text-sm text-muted-foreground">
              Provide a reason for rejection. The provider will see this message.
            </p>
            <Input
              className="mt-4"
              placeholder="Reason for rejection"
              value={rejectReason}
              onChange={(e) => setRejectReason(e.target.value)}
            />
            <div className="mt-6 flex justify-end gap-2">
              <Button
                variant="outline"
                onClick={() => {
                  setRejectTarget(null);
                  setRejectReason("");
                }}
              >
                Cancel
              </Button>
              <Button
                variant="destructive"
                disabled={!rejectReason.trim() || actions.reject.isPending}
                onClick={async () => {
                  if (!rejectTarget || !rejectReason.trim()) return;
                  await actions.reject.mutateAsync({
                    id: rejectTarget.id,
                    reason: rejectReason.trim(),
                  });
                  setRejectTarget(null);
                  setRejectReason("");
                }}
              >
                {actions.reject.isPending ? "Working..." : "Reject"}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
