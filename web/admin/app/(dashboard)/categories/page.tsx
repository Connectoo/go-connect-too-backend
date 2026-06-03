"use client";

import { useState } from "react";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { DataTable, type Column } from "@/components/admin/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAdminCategories, useCategoryMutations } from "@/hooks/use-admin";
import type { Category } from "@/types/admin";

export default function CategoriesPage() {
  const { data, isLoading } = useAdminCategories();
  const mutations = useCategoryMutations();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [deleteId, setDeleteId] = useState<string | null>(null);

  const columns: Column<Category>[] = [
    { key: "name", header: "Name", cell: (row) => row.name },
    {
      key: "description",
      header: "Description",
      cell: (row) => row.description ?? "—",
    },
    {
      key: "active",
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
        <Button size="sm" variant="destructive" onClick={() => setDeleteId(row.id)}>
          Delete
        </Button>
      ),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Categories</h1>
        <p className="text-sm text-muted-foreground">Manage service categories.</p>
      </div>

      <form
        className="mb-6 grid gap-4 rounded-lg border p-4 md:grid-cols-3"
        onSubmit={(e) => {
          e.preventDefault();
          mutations.create.mutate(
            { name, description: description || undefined, is_active: true },
            {
              onSuccess: () => {
                setName("");
                setDescription("");
              },
            },
          );
        }}
      >
        <div className="space-y-2">
          <Label htmlFor="name">Name</Label>
          <Input id="name" value={name} onChange={(e) => setName(e.target.value)} required />
        </div>
        <div className="space-y-2 md:col-span-2">
          <Label htmlFor="description">Description</Label>
          <Input
            id="description"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
          />
        </div>
        <Button type="submit" disabled={mutations.create.isPending}>
          Add category
        </Button>
      </form>

      <DataTable columns={columns} data={data ?? []} isLoading={isLoading} />

      <ConfirmDialog
        open={Boolean(deleteId)}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title="Delete category?"
        description="This cannot be undone if services are linked. Use with caution."
        confirmLabel="Delete"
        variant="destructive"
        loading={mutations.remove.isPending}
        onConfirm={async () => {
          if (!deleteId) return;
          await mutations.remove.mutateAsync(deleteId);
          setDeleteId(null);
        }}
      />
    </div>
  );
}
