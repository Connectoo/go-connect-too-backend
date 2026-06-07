"use client";

import Link from "next/link";
import { useState } from "react";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { DataTable, type Column } from "@/components/admin/data-table";
import { Pagination } from "@/components/admin/pagination";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAdminEmployees, useEmployeeActions } from "@/hooks/use-admin";
import type { AdminEmployee } from "@/types/admin";

export default function EmployeesPage() {
  const [page, setPage] = useState(1);
  const [q, setQ] = useState("");
  const [status, setStatus] = useState("pending");
  const [confirm, setConfirm] = useState<{
    id: string;
    action: "approve" | "reject";
  } | null>(null);

  const limit = 20;
  const { data, isLoading } = useAdminEmployees({
    page,
    limit,
    verification_status: status || undefined,
    q: q || undefined,
  });
  const actions = useEmployeeActions();

  const columns: Column<AdminEmployee>[] = [
    {
      key: "name",
      header: "Provider",
      cell: (row) => (
        <div>
          <Link
            href={`/employees/${row.id}`}
            className="font-medium hover:underline"
          >
            {row.display_name ?? row.user_name}
          </Link>
          <p className="text-xs text-muted-foreground">{row.user_email}</p>
        </div>
      ),
    },
    {
      key: "status",
      header: "Verification",
      cell: (row) => <Badge variant="secondary">{row.verification_status}</Badge>,
    },
    {
      key: "user_status",
      header: "Account",
      cell: (row) => row.user_status,
    },
    {
      key: "actions",
      header: "Actions",
      cell: (row) =>
        row.verification_status === "pending" ? (
          <div className="flex gap-2">
            <Button
              size="sm"
              onClick={() => setConfirm({ id: row.id, action: "approve" })}
            >
              Approve
            </Button>
            <Button
              size="sm"
              variant="outline"
              onClick={() => setConfirm({ id: row.id, action: "reject" })}
            >
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
        <h1 className="text-2xl font-bold">Employee approvals</h1>
        <p className="text-sm text-muted-foreground">
          Review and approve provider profiles (KYC / verification).
        </p>
      </div>

      <div className="mb-4 flex flex-wrap gap-3">
        <Input
          placeholder="Search name or email..."
          className="max-w-xs"
          value={q}
          onChange={(e) => {
            setQ(e.target.value);
            setPage(1);
          }}
        />
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

      <ConfirmDialog
        open={Boolean(confirm)}
        onOpenChange={(open) => !open && setConfirm(null)}
        title={confirm?.action === "approve" ? "Approve employee?" : "Reject employee?"}
        description="This updates the provider verification status and may affect their visibility on the marketplace."
        confirmLabel={confirm?.action === "approve" ? "Approve" : "Reject"}
        variant={confirm?.action === "reject" ? "destructive" : "default"}
        loading={actions.approve.isPending || actions.reject.isPending}
        onConfirm={async () => {
          if (!confirm) return;
          if (confirm.action === "approve") {
            await actions.approve.mutateAsync(confirm.id);
          } else {
            await actions.reject.mutateAsync(confirm.id);
          }
          setConfirm(null);
        }}
      />
    </div>
  );
}
