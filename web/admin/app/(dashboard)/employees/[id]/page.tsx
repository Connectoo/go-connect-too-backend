"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { useState } from "react";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  useAdminEmployee,
  useEmployeeSuspendMutation,
} from "@/hooks/use-admin";

function DetailRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1 border-b py-3 last:border-0 sm:flex-row sm:items-center sm:justify-between">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className="text-sm font-medium">{value}</span>
    </div>
  );
}

export default function EmployeeDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const { data: employee, isLoading, isError, error } = useAdminEmployee(id);
  const suspend = useEmployeeSuspendMutation(id);
  const [confirmOpen, setConfirmOpen] = useState(false);

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <Link
          href="/employees"
          className="mb-3 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to employees
        </Link>
        <h1 className="text-2xl font-bold">Provider detail</h1>
      </div>

      {isLoading && (
        <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">
          Loading...
        </div>
      )}

      {isError && (
        <div className="rounded-lg border border-destructive/40 p-4 text-sm text-destructive">
          {error instanceof Error ? error.message : "Failed to load employee."}
        </div>
      )}

      {employee && (
        <div className="grid gap-6 lg:grid-cols-3">
          <Card className="lg:col-span-2">
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <CardTitle>{employee.display_name ?? employee.user_name}</CardTitle>
              <Badge variant="secondary">{employee.verification_status}</Badge>
            </CardHeader>
            <CardContent>
              <DetailRow label="Email" value={employee.user_email} />
              <DetailRow label="Phone" value={employee.phone ?? "—"} />
              <DetailRow label="Account status" value={employee.user_status} />
              <DetailRow label="Experience" value={`${employee.experience_years} years`} />
              <DetailRow label="Bio" value={employee.bio ?? "—"} />
              <DetailRow
                label="Skills"
                value={employee.skills?.length ? employee.skills.join(", ") : "—"}
              />
              <DetailRow
                label="Languages"
                value={employee.languages?.length ? employee.languages.join(", ") : "—"}
              />
              <DetailRow label="Location" value={employee.location_text ?? "—"} />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Actions</CardTitle>
            </CardHeader>
            <CardContent>
              {employee.user_status === "suspended" ? (
                <p className="text-sm text-muted-foreground">
                  This provider account is suspended.
                </p>
              ) : (
                <Button variant="destructive" onClick={() => setConfirmOpen(true)}>
                  Suspend provider
                </Button>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      <ConfirmDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title="Suspend provider?"
        description="The provider will be unable to sign in or receive new bookings."
        confirmLabel="Suspend"
        variant="destructive"
        loading={suspend.isPending}
        onConfirm={async () => {
          await suspend.mutateAsync();
          setConfirmOpen(false);
        }}
      />
    </div>
  );
}
