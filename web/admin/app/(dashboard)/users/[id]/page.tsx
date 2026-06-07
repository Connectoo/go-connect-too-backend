"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { useState } from "react";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAdminUser, useUserActions } from "@/hooks/use-admin";

function DetailRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1 border-b py-3 last:border-0 sm:flex-row sm:items-center sm:justify-between">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className="text-sm font-medium">{value}</span>
    </div>
  );
}

export default function UserDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const { data: user, isLoading, isError, error } = useAdminUser(id);
  const actions = useUserActions(id);
  const [confirmAction, setConfirmAction] = useState<"suspend" | "activate" | null>(
    null,
  );

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <Link
          href="/users"
          className="mb-3 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to users
        </Link>
        <h1 className="text-2xl font-bold">User detail</h1>
      </div>

      {isLoading && (
        <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">
          Loading...
        </div>
      )}

      {isError && (
        <div className="rounded-lg border border-destructive/40 p-4 text-sm text-destructive">
          {error instanceof Error ? error.message : "Failed to load user."}
        </div>
      )}

      {user && (
        <div className="grid gap-6 lg:grid-cols-3">
          <Card className="lg:col-span-2">
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <CardTitle>{user.name}</CardTitle>
              <div className="flex gap-2">
                <Badge variant="outline">{user.role}</Badge>
                <Badge variant="secondary">{user.status}</Badge>
              </div>
            </CardHeader>
            <CardContent>
              <DetailRow label="Email" value={user.email} />
              <DetailRow label="Phone" value={user.phone ?? "—"} />
              <DetailRow
                label="Joined"
                value={new Date(user.created_at).toLocaleDateString()}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Actions</CardTitle>
            </CardHeader>
            <CardContent className="space-y-2">
              {user.status === "suspended" || user.status === "inactive" ? (
                <Button onClick={() => setConfirmAction("activate")}>
                  Activate user
                </Button>
              ) : (
                <Button variant="destructive" onClick={() => setConfirmAction("suspend")}>
                  Suspend user
                </Button>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      <ConfirmDialog
        open={confirmAction === "suspend"}
        onOpenChange={(open) => !open && setConfirmAction(null)}
        title="Suspend user?"
        description="The user will be unable to sign in until reactivated."
        confirmLabel="Suspend"
        variant="destructive"
        loading={actions.suspend.isPending}
        onConfirm={async () => {
          await actions.suspend.mutateAsync();
          setConfirmAction(null);
        }}
      />

      <ConfirmDialog
        open={confirmAction === "activate"}
        onOpenChange={(open) => !open && setConfirmAction(null)}
        title="Activate user?"
        description="The user will be able to sign in again."
        confirmLabel="Activate"
        loading={actions.activate.isPending}
        onConfirm={async () => {
          await actions.activate.mutateAsync();
          setConfirmAction(null);
        }}
      />
    </div>
  );
}
