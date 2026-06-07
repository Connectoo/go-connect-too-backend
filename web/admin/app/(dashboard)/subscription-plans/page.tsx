"use client";

import { useState } from "react";
import { DataTable, type Column } from "@/components/admin/data-table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { usePlanMutations, useSubscriptionPlans } from "@/hooks/use-ops";
import type { SubscriptionPlan } from "@/types/ops";

function formatPrice(amount: number, currency: string) {
  if (currency === "INR") return `₹${(amount / 100).toLocaleString("en-IN")}`;
  return `${(amount / 100).toFixed(2)} ${currency}`;
}

export default function SubscriptionPlansPage() {
  const { data, isLoading } = useSubscriptionPlans();
  const mutations = usePlanMutations();
  const [form, setForm] = useState({
    name: "",
    price: "",
    currency: "INR",
    duration_days: "30",
    service_limit: "5",
    is_featured_allowed: false,
    is_priority_allowed: false,
  });

  const columns: Column<SubscriptionPlan>[] = [
    { key: "name", header: "Name", cell: (row) => row.name },
    {
      key: "price",
      header: "Price",
      cell: (row) => formatPrice(row.price, row.currency),
    },
    {
      key: "duration",
      header: "Duration",
      cell: (row) => `${row.duration_days} days`,
    },
    {
      key: "limit",
      header: "Service limit",
      cell: (row) => row.service_limit,
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
        <Button
          size="sm"
          variant="outline"
          disabled={mutations.update.isPending}
          onClick={() =>
            mutations.update.mutate({
              id: row.id,
              name: row.name,
              price: row.price,
              currency: row.currency,
              duration_days: row.duration_days,
              service_limit: row.service_limit,
              is_featured_allowed: row.is_featured_allowed,
              is_priority_allowed: row.is_priority_allowed,
              is_active: !row.is_active,
            })
          }
        >
          {row.is_active ? "Deactivate" : "Activate"}
        </Button>
      ),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Subscription plans</h1>
        <p className="text-sm text-muted-foreground">Create and manage provider plans.</p>
      </div>

      <form
        className="mb-6 grid gap-4 rounded-lg border p-4 md:grid-cols-3"
        onSubmit={(e) => {
          e.preventDefault();
          mutations.create.mutate(
            {
              name: form.name,
              price: Math.round(parseFloat(form.price) * 100),
              currency: form.currency,
              duration_days: parseInt(form.duration_days, 10),
              service_limit: parseInt(form.service_limit, 10),
              is_featured_allowed: form.is_featured_allowed,
              is_priority_allowed: form.is_priority_allowed,
              is_active: true,
            },
            {
              onSuccess: () =>
                setForm({
                  name: "",
                  price: "",
                  currency: "INR",
                  duration_days: "30",
                  service_limit: "5",
                  is_featured_allowed: false,
                  is_priority_allowed: false,
                }),
            },
          );
        }}
      >
        <div className="space-y-2">
          <Label htmlFor="name">Name</Label>
          <Input
            id="name"
            value={form.name}
            onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
            required
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="price">Price (₹)</Label>
          <Input
            id="price"
            type="number"
            min="0"
            step="0.01"
            value={form.price}
            onChange={(e) => setForm((f) => ({ ...f, price: e.target.value }))}
            required
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="duration">Duration (days)</Label>
          <Input
            id="duration"
            type="number"
            min="1"
            value={form.duration_days}
            onChange={(e) => setForm((f) => ({ ...f, duration_days: e.target.value }))}
            required
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="limit">Service limit</Label>
          <Input
            id="limit"
            type="number"
            min="1"
            value={form.service_limit}
            onChange={(e) => setForm((f) => ({ ...f, service_limit: e.target.value }))}
            required
          />
        </div>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={form.is_featured_allowed}
            onChange={(e) =>
              setForm((f) => ({ ...f, is_featured_allowed: e.target.checked }))
            }
          />
          Featured allowed
        </label>
        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={form.is_priority_allowed}
            onChange={(e) =>
              setForm((f) => ({ ...f, is_priority_allowed: e.target.checked }))
            }
          />
          Priority allowed
        </label>
        <Button type="submit" disabled={mutations.create.isPending}>
          Create plan
        </Button>
      </form>

      <DataTable columns={columns} data={data ?? []} isLoading={isLoading} />
    </div>
  );
}
