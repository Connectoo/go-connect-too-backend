"use client";

import { useState } from "react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import { ConfirmDialog } from "@/components/admin/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { useAdminBooking, useBookingStatusMutation } from "@/hooks/use-admin";
import { BOOKING_STATUSES } from "@/types/admin";

const STATUS_LABELS: Record<string, string> = {
  pending: "Pending",
  accepted: "Accepted",
  in_progress: "In progress",
  completed: "Completed",
  rejected: "Rejected",
  cancelled: "Cancelled",
  no_show: "No show",
};

function DetailRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-1 border-b py-3 last:border-0 sm:flex-row sm:items-center sm:justify-between">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className="text-sm font-medium">{value}</span>
    </div>
  );
}

export default function BookingDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data: booking, isLoading, isError, error } = useAdminBooking(id);
  const statusMutation = useBookingStatusMutation(id);

  const [nextStatus, setNextStatus] = useState("");
  const [confirmOpen, setConfirmOpen] = useState(false);

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <Link
          href="/bookings"
          className="mb-3 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to bookings
        </Link>
        <h1 className="text-2xl font-bold">Booking detail</h1>
        <p className="text-sm text-muted-foreground">
          Review booking information and override its status.
        </p>
      </div>

      {isLoading && (
        <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">
          Loading...
        </div>
      )}

      {isError && (
        <div className="rounded-lg border border-destructive/40 bg-destructive/10 p-4 text-sm text-destructive">
          {error instanceof Error ? error.message : "Failed to load booking."}
        </div>
      )}

      {booking && (
        <div className="grid gap-6 lg:grid-cols-3">
          <Card className="lg:col-span-2">
            <CardHeader className="flex flex-row items-center justify-between space-y-0">
              <CardTitle>Overview</CardTitle>
              <Badge variant="secondary">
                {STATUS_LABELS[booking.status] ?? booking.status}
              </Badge>
            </CardHeader>
            <CardContent>
              <DetailRow
                label="Date & time"
                value={`${booking.booking_date} · ${booking.start_time} – ${booking.end_time}`}
              />
              <DetailRow
                label="Amount"
                value={`₹${booking.total_amount.toLocaleString("en-IN")}`}
              />
              <DetailRow label="Customer ID" value={booking.customer_id} />
              <DetailRow label="Employee ID" value={booking.employee_id} />
              <DetailRow label="Service ID" value={booking.service_id} />
              <DetailRow
                label="Customer notes"
                value={booking.customer_notes ?? "—"}
              />
              <DetailRow
                label="Employee notes"
                value={booking.employee_notes ?? "—"}
              />
              <DetailRow
                label="Created"
                value={new Date(booking.created_at).toLocaleString("en-IN")}
              />
              <DetailRow
                label="Updated"
                value={new Date(booking.updated_at).toLocaleString("en-IN")}
              />
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Override status</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="status">New status</Label>
                <select
                  id="status"
                  className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
                  value={nextStatus}
                  onChange={(e) => setNextStatus(e.target.value)}
                >
                  <option value="">Select a status…</option>
                  {BOOKING_STATUSES.map((status) => (
                    <option
                      key={status}
                      value={status}
                      disabled={status === booking.status}
                    >
                      {STATUS_LABELS[status]}
                      {status === booking.status ? " (current)" : ""}
                    </option>
                  ))}
                </select>
              </div>

              {statusMutation.isError && (
                <p className="text-sm text-destructive">
                  {statusMutation.error instanceof Error
                    ? statusMutation.error.message
                    : "Failed to update status."}
                </p>
              )}

              <Button
                className="w-full"
                disabled={!nextStatus || nextStatus === booking.status}
                onClick={() => setConfirmOpen(true)}
              >
                Apply status change
              </Button>
            </CardContent>
          </Card>
        </div>
      )}

      <ConfirmDialog
        open={confirmOpen}
        onOpenChange={(open) => !open && setConfirmOpen(false)}
        title="Override booking status?"
        description={`This will force the booking status to "${
          STATUS_LABELS[nextStatus] ?? nextStatus
        }". This bypasses normal transition rules.`}
        confirmLabel="Apply"
        loading={statusMutation.isPending}
        onConfirm={async () => {
          if (!nextStatus) return;
          await statusMutation.mutateAsync(nextStatus);
          setConfirmOpen(false);
          setNextStatus("");
        }}
      />
    </div>
  );
}
