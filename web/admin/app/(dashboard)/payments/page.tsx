"use client";

import { useState } from "react";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { DataTable, type Column } from "@/components/admin/data-table";
import { Pagination } from "@/components/admin/pagination";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useAdminPayments, usePaymentRefund } from "@/hooks/use-ops";
import type { Payment } from "@/types/ops";

function formatAmount(amount: number, currency: string) {
  if (currency === "INR") return `₹${(amount / 100).toLocaleString("en-IN")}`;
  return `${(amount / 100).toFixed(2)} ${currency}`;
}

export default function PaymentsPage() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState("");
  const [refundId, setRefundId] = useState<string | null>(null);
  const limit = 20;

  const { data, isLoading } = useAdminPayments({
    page,
    limit,
    status: status || undefined,
  });
  const refund = usePaymentRefund();

  const columns: Column<Payment>[] = [
    {
      key: "date",
      header: "Date",
      cell: (row) => new Date(row.created_at).toLocaleDateString(),
    },
    {
      key: "amount",
      header: "Amount",
      cell: (row) => formatAmount(row.amount, row.currency),
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => <Badge variant="secondary">{row.status}</Badge>,
    },
    {
      key: "provider",
      header: "Provider",
      cell: (row) => row.provider,
    },
    {
      key: "actions",
      header: "",
      cell: (row) =>
        row.status === "success" ? (
          <Button size="sm" variant="outline" onClick={() => setRefundId(row.id)}>
            Refund
          </Button>
        ) : null,
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Payments</h1>
          <p className="text-sm text-muted-foreground">Subscription payments and refunds.</p>
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
          <option value="success">Success</option>
          <option value="failed">Failed</option>
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
        open={Boolean(refundId)}
        onOpenChange={(open) => !open && setRefundId(null)}
        title="Issue refund?"
        description="This will initiate a full refund for the payment."
        confirmLabel="Refund"
        variant="destructive"
        loading={refund.isPending}
        onConfirm={async () => {
          if (!refundId) return;
          await refund.mutateAsync({ id: refundId });
          setRefundId(null);
        }}
      />
    </div>
  );
}
