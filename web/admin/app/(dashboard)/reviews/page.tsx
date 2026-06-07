"use client";

import { useState } from "react";
import { DataTable, type Column } from "@/components/admin/data-table";
import { Pagination } from "@/components/admin/pagination";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useAdminReviews, useReviewModeration } from "@/hooks/use-ops";
import type { AdminReview } from "@/types/ops";

export default function ReviewsPage() {
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState("");
  const limit = 20;

  const { data, isLoading } = useAdminReviews({
    page,
    limit,
    status: status || undefined,
  });
  const moderation = useReviewModeration();

  const columns: Column<AdminReview>[] = [
    {
      key: "rating",
      header: "Rating",
      cell: (row) => `${row.rating} ★`,
    },
    {
      key: "comment",
      header: "Comment",
      cell: (row) => (
        <span className="line-clamp-2 max-w-xs text-sm">
          {row.comment ?? "—"}
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
      header: "Date",
      cell: (row) => new Date(row.created_at).toLocaleDateString(),
    },
    {
      key: "actions",
      header: "Actions",
      cell: (row) => (
        <div className="flex gap-2">
          {row.status !== "approved" && (
            <Button
              size="sm"
              onClick={() => moderation.approve.mutate(row.id)}
            >
              Approve
            </Button>
          )}
          {row.status !== "hidden" && (
            <Button
              size="sm"
              variant="outline"
              onClick={() => moderation.hide.mutate(row.id)}
            >
              Hide
            </Button>
          )}
        </div>
      ),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Reviews</h1>
          <p className="text-sm text-muted-foreground">Moderate customer reviews.</p>
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
          <option value="approved">Approved</option>
          <option value="hidden">Hidden</option>
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
