"use client";

import Link from "next/link";
import { useState } from "react";
import { DataTable, type Column } from "@/components/admin/data-table";
import { Pagination } from "@/components/admin/pagination";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { useAdminUsers } from "@/hooks/use-admin";
import type { AdminUser } from "@/types/admin";

export default function UsersPage() {
  const [page, setPage] = useState(1);
  const [q, setQ] = useState("");
  const [role, setRole] = useState("");
  const [status, setStatus] = useState("");

  const limit = 20;
  const { data, isLoading } = useAdminUsers({
    page,
    limit,
    q: q || undefined,
    role: role || undefined,
    status: status || undefined,
  });

  const columns: Column<AdminUser>[] = [
    {
      key: "name",
      header: "User",
      cell: (row) => (
        <Link href={`/users/${row.id}`} className="font-medium hover:underline">
          {row.name}
        </Link>
      ),
    },
    { key: "email", header: "Email", cell: (row) => row.email },
    {
      key: "role",
      header: "Role",
      cell: (row) => <Badge variant="outline">{row.role}</Badge>,
    },
    {
      key: "status",
      header: "Status",
      cell: (row) => <Badge variant="secondary">{row.status}</Badge>,
    },
    {
      key: "created",
      header: "Joined",
      cell: (row) => new Date(row.created_at).toLocaleDateString(),
    },
  ];

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Users</h1>
        <p className="text-sm text-muted-foreground">
          Manage customer, provider, and admin accounts.
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
          value={role}
          onChange={(e) => {
            setRole(e.target.value);
            setPage(1);
          }}
        >
          <option value="">All roles</option>
          <option value="customer">Customer</option>
          <option value="employee">Employee</option>
          <option value="admin">Admin</option>
        </select>
        <select
          className="h-10 rounded-md border border-input bg-background px-3 text-sm"
          value={status}
          onChange={(e) => {
            setStatus(e.target.value);
            setPage(1);
          }}
        >
          <option value="">All statuses</option>
          <option value="active">Active</option>
          <option value="inactive">Inactive</option>
          <option value="suspended">Suspended</option>
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
