"use client";

import Link from "next/link";
import { DataTable, type Column } from "@/components/shared/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { useCategories, useServices } from "@/hooks/use-services";
import type { EmployeeService } from "@/types/service";

export default function ServicesPage() {
  const { data, isLoading, isError, error } = useServices();
  const { data: categories } = useCategories();

  const categoryName = (id: string) =>
    categories?.find((c) => c.id === id)?.name ?? "—";

  const columns: Column<EmployeeService>[] = [
    { key: "title", header: "Name", cell: (row) => row.title },
    {
      key: "category",
      header: "Category",
      cell: (row) => categoryName(row.category_id),
    },
    {
      key: "price",
      header: "Price",
      cell: (row) => row.price.toFixed(2),
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
      header: "",
      cell: (row) => (
        <Button asChild size="sm" variant="outline">
          <Link href={`/services/${row.id}`}>View</Link>
        </Button>
      ),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">My services</h1>
          <p className="text-sm text-muted-foreground">
            Services you offer to customers.
          </p>
        </div>
        <Button asChild>
          <Link href="/services/new">New service</Link>
        </Button>
      </div>

      {isError ? (
        <div className="rounded-lg border border-destructive/40 p-8 text-center text-sm text-destructive">
          {(error as Error)?.message || "Failed to load services"}
        </div>
      ) : (
        <DataTable
          columns={columns}
          data={data ?? []}
          isLoading={isLoading}
          emptyMessage="No services yet. Create your first service."
        />
      )}
    </div>
  );
}
