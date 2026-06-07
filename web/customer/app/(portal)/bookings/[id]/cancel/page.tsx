"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useState } from "react";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { useBooking, useBookingMutations } from "@/hooks/use-bookings";
import { formatBookingDate } from "@/lib/booking-slots";

export default function CancelBookingPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const router = useRouter();
  const { data: booking } = useBooking(id);
  const { cancel } = useBookingMutations(id);
  const [reason, setReason] = useState("");
  const [confirmOpen, setConfirmOpen] = useState(false);

  return (
    <div className="p-6 md:p-8">
      <Link
        href={`/bookings/${id}`}
        className="mb-4 inline-block text-sm text-muted-foreground hover:text-foreground"
      >
        ← Back to booking
      </Link>

      <Card className="max-w-md">
        <CardHeader>
          <CardTitle>Cancel booking</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {booking && (
            <p className="text-sm text-muted-foreground">
              {formatBookingDate(booking.booking_date)} · {booking.start_time} –{" "}
              {booking.end_time}
            </p>
          )}

          <div className="space-y-2">
            <Label htmlFor="reason">Reason (optional)</Label>
            <textarea
              id="reason"
              rows={3}
              value={reason}
              onChange={(e) => setReason(e.target.value)}
              className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
              placeholder="Why are you cancelling?"
            />
          </div>

          {cancel.isError && (
            <p className="text-sm text-destructive">
              {(cancel.error as Error)?.message || "Failed to cancel booking"}
            </p>
          )}

          <div className="flex gap-3">
            <Button variant="destructive" onClick={() => setConfirmOpen(true)}>
              Cancel booking
            </Button>
            <Button variant="outline" asChild>
              <Link href={`/bookings/${id}`}>Keep booking</Link>
            </Button>
          </div>
        </CardContent>
      </Card>

      <ConfirmDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title="Confirm cancellation?"
        description="This action cannot be undone. The provider will be notified."
        confirmLabel="Yes, cancel"
        variant="destructive"
        loading={cancel.isPending}
        onConfirm={async () => {
          await cancel.mutateAsync(reason.trim() ? { reason: reason.trim() } : undefined);
          router.push(`/bookings/${id}`);
        }}
      />
    </div>
  );
}
