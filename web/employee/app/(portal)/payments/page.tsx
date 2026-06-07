"use client";

import { DataTable, type Column } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { usePayments } from "@/hooks/use-payments";
import { formatPlanPrice } from "@/lib/razorpay";
import type { Payment } from "@/types/payment";

export default function PaymentsPage() {
  const { data, isLoading } = usePayments();

  const columns: Column<Payment>[] = [
    {
      key: "date",
      header: "Date",
      cell: (row) => new Date(row.created_at).toLocaleDateString(),
    },
    {
      key: "amount",
      header: "Amount",
      cell: (row) => formatPlanPrice(row.amount, row.currency),
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
      key: "order",
      header: "Order ID",
      cell: (row) => (
        <span className="text-xs text-muted-foreground">
          {row.provider_order_id.slice(0, 12)}…
        </span>
      ),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Payment history</h1>
        <p className="text-sm text-muted-foreground">
          Subscription payments processed via Razorpay.
        </p>
      </div>

      <DataTable columns={columns} data={data ?? []} isLoading={isLoading} />
    </div>
  );
}
