"use client";

import { useState } from "react";
import { DataTable, type Column } from "@/components/admin/data-table";
import { Pagination } from "@/components/admin/pagination";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useAdminServices, useServiceModeration } from "@/hooks/use-ops";
import type { AdminService } from "@/types/ops";

export default function ServicesPage() {
  const [page, setPage] = useState(1);
  const [isActive, setIsActive] = useState("");
  const [q, setQ] = useState("");
  const limit = 20;

  const { data, isLoading } = useAdminServices({
    page,
    limit,
    is_active: isActive || undefined,
    q: q || undefined,
  });
  const moderation = useServiceModeration();

  const columns: Column<AdminService>[] = [
    { key: "title", header: "Title", cell: (row) => row.title },
    {
      key: "price",
      header: "Price",
      cell: (row) => `₹${row.price.toLocaleString("en-IN")}`,
    },
    {
      key: "duration",
      header: "Duration",
      cell: (row) => `${row.duration_minutes} min`,
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => (
        <Badge variant={row.is_active ? "default" : "secondary"}>
          {row.is_active ? "Active" : "Inactive"}
        </Badge>
      ),
    },
    {
      key: "actions",
      header: "Actions",
      cell: (row) => (
        <Button
          size="sm"
          variant="outline"
          disabled={moderation.activate.isPending || moderation.deactivate.isPending}
          onClick={() =>
            row.is_active
              ? moderation.deactivate.mutate(row.id)
              : moderation.activate.mutate(row.id)
          }
        >
          {row.is_active ? "Deactivate" : "Activate"}
        </Button>
      ),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="text-2xl font-bold">Services</h1>
          <p className="text-sm text-muted-foreground">Moderate provider service listings.</p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Input
            placeholder="Search…"
            value={q}
            onChange={(e) => {
              setQ(e.target.value);
              setPage(1);
            }}
            className="w-48"
          />
          <select
            className="h-10 rounded-md border border-input bg-background px-3 text-sm"
            value={isActive}
            onChange={(e) => {
              setIsActive(e.target.value);
              setPage(1);
            }}
          >
            <option value="">All</option>
            <option value="true">Active</option>
            <option value="false">Inactive</option>
          </select>
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
