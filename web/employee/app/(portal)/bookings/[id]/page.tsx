"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { useState } from "react";
import { ArrowLeft } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { useBookingActions, useBookings } from "@/hooks/use-bookings";
import type { Booking } from "@/types/booking";

type DestructiveAction = "reject" | "cancel" | "noShow";

const DESTRUCTIVE_COPY: Record<
  DestructiveAction,
  { title: string; description: string; confirmLabel: string }
> = {
  reject: {
    title: "Reject booking",
    description: "The customer will be notified that you declined this request.",
    confirmLabel: "Reject",
  },
  cancel: {
    title: "Cancel booking",
    description: "This will cancel the accepted booking. This cannot be undone.",
    confirmLabel: "Cancel booking",
  },
  noShow: {
    title: "Mark as no-show",
    description: "Mark this booking as a customer no-show.",
    confirmLabel: "Mark no-show",
  },
};

function DetailRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex justify-between gap-4 border-b py-2 last:border-0">
      <span className="text-sm text-muted-foreground">{label}</span>
      <span className="text-sm font-medium">{value}</span>
    </div>
  );
}

export default function BookingDetailPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;

  const { data, isLoading, isError } = useBookings();
  const actions = useBookingActions();

  const [confirm, setConfirm] = useState<DestructiveAction | null>(null);
  const [showReschedule, setShowReschedule] = useState(false);
  const [rescheduleForm, setRescheduleForm] = useState({
    booking_date: "",
    start_time: "",
    end_time: "",
    reason: "",
  });

  const booking: Booking | undefined = data?.find((b) => b.id === id);

  if (isLoading) {
    return (
      <div className="p-6 md:p-8">
        <Skeleton className="mb-4 h-8 w-48" />
        <Skeleton className="h-64" />
      </div>
    );
  }

  if (isError || !booking) {
    return (
      <div className="p-6 md:p-8">
        <Link
          href="/bookings"
          className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Back to bookings
        </Link>
        <p className="text-sm text-destructive">
          Booking not found. It may not belong to your account.
        </p>
      </div>
    );
  }

  const anyPending =
    actions.accept.isPending ||
    actions.reject.isPending ||
    actions.start.isPending ||
    actions.complete.isPending ||
    actions.cancel.isPending ||
    actions.noShow.isPending ||
    actions.reschedule.isPending;

  function runDestructive() {
    if (!confirm) return;
    const onSettled = () => setConfirm(null);
    if (confirm === "reject") {
      actions.reject.mutate({ id }, { onSuccess: onSettled });
    } else if (confirm === "cancel") {
      actions.cancel.mutate({ id }, { onSuccess: onSettled });
    } else if (confirm === "noShow") {
      actions.noShow.mutate({ id }, { onSuccess: onSettled });
    }
  }

  function submitReschedule(e: React.FormEvent) {
    e.preventDefault();
    actions.reschedule.mutate(
      {
        id,
        body: {
          booking_date: rescheduleForm.booking_date,
          start_time: rescheduleForm.start_time,
          end_time: rescheduleForm.end_time,
          reason: rescheduleForm.reason || undefined,
        },
      },
      { onSuccess: () => setShowReschedule(false) },
    );
  }

  const status = booking.status;
  const destructiveCopy = confirm ? DESTRUCTIVE_COPY[confirm] : null;

  return (
    <div className="p-6 md:p-8">
      <Link
        href="/bookings"
        className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" /> Back to bookings
      </Link>

      <div className="mb-6 flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-2xl font-bold">Booking detail</h1>
        <Badge variant="secondary">{status}</Badge>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Schedule</CardTitle>
          </CardHeader>
          <CardContent>
            <DetailRow label="Date" value={booking.booking_date} />
            <DetailRow
              label="Time"
              value={`${booking.start_time} – ${booking.end_time}`}
            />
            <DetailRow
              label="Amount"
              value={`₹${booking.total_amount.toLocaleString("en-IN")}`}
            />
            <DetailRow label="Service ID" value={booking.service_id} />
            <DetailRow label="Customer ID" value={booking.customer_id} />
            {booking.customer_notes && (
              <DetailRow label="Customer notes" value={booking.customer_notes} />
            )}
            {booking.employee_notes && (
              <DetailRow label="Your notes" value={booking.employee_notes} />
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">Actions</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-wrap gap-2">
            {status === "pending" && (
              <>
                <Button
                  onClick={() => actions.accept.mutate(id)}
                  disabled={anyPending}
                >
                  Accept
                </Button>
                <Button
                  variant="destructive"
                  onClick={() => setConfirm("reject")}
                  disabled={anyPending}
                >
                  Reject
                </Button>
              </>
            )}

            {status === "accepted" && (
              <>
                <Button
                  onClick={() => actions.start.mutate(id)}
                  disabled={anyPending}
                >
                  Start
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setShowReschedule((v) => !v)}
                  disabled={anyPending}
                >
                  Reschedule
                </Button>
                <Button
                  variant="destructive"
                  onClick={() => setConfirm("cancel")}
                  disabled={anyPending}
                >
                  Cancel
                </Button>
                <Button
                  variant="destructive"
                  onClick={() => setConfirm("noShow")}
                  disabled={anyPending}
                >
                  No-show
                </Button>
              </>
            )}

            {status === "in_progress" && (
              <>
                <Button
                  onClick={() => actions.complete.mutate(id)}
                  disabled={anyPending}
                >
                  Complete
                </Button>
                <Button
                  variant="destructive"
                  onClick={() => setConfirm("noShow")}
                  disabled={anyPending}
                >
                  No-show
                </Button>
              </>
            )}

            {["completed", "rejected", "cancelled", "no_show"].includes(
              status,
            ) && (
              <p className="text-sm text-muted-foreground">
                No actions available for this booking.
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      {showReschedule && status === "accepted" && (
        <Card className="mt-6">
          <CardHeader>
            <CardTitle className="text-base">Reschedule booking</CardTitle>
          </CardHeader>
          <CardContent>
            <form
              onSubmit={submitReschedule}
              className="grid gap-4 sm:grid-cols-2"
            >
              <div className="space-y-1">
                <Label htmlFor="booking_date">Date</Label>
                <Input
                  id="booking_date"
                  type="date"
                  required
                  value={rescheduleForm.booking_date}
                  onChange={(e) =>
                    setRescheduleForm((f) => ({
                      ...f,
                      booking_date: e.target.value,
                    }))
                  }
                />
              </div>
              <div className="space-y-1">
                <Label htmlFor="reason">Reason (optional)</Label>
                <Input
                  id="reason"
                  value={rescheduleForm.reason}
                  onChange={(e) =>
                    setRescheduleForm((f) => ({ ...f, reason: e.target.value }))
                  }
                />
              </div>
              <div className="space-y-1">
                <Label htmlFor="start_time">Start time</Label>
                <Input
                  id="start_time"
                  type="time"
                  required
                  value={rescheduleForm.start_time}
                  onChange={(e) =>
                    setRescheduleForm((f) => ({
                      ...f,
                      start_time: e.target.value,
                    }))
                  }
                />
              </div>
              <div className="space-y-1">
                <Label htmlFor="end_time">End time</Label>
                <Input
                  id="end_time"
                  type="time"
                  required
                  value={rescheduleForm.end_time}
                  onChange={(e) =>
                    setRescheduleForm((f) => ({
                      ...f,
                      end_time: e.target.value,
                    }))
                  }
                />
              </div>
              <div className="sm:col-span-2 flex gap-2">
                <Button type="submit" disabled={anyPending}>
                  {actions.reschedule.isPending ? "Saving..." : "Save new time"}
                </Button>
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => setShowReschedule(false)}
                >
                  Cancel
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      <ConfirmDialog
        open={confirm !== null}
        onOpenChange={(open) => {
          if (!open) setConfirm(null);
        }}
        title={destructiveCopy?.title ?? ""}
        description={destructiveCopy?.description ?? ""}
        confirmLabel={destructiveCopy?.confirmLabel}
        variant="destructive"
        loading={anyPending}
        onConfirm={runDestructive}
      />
    </div>
  );
}
