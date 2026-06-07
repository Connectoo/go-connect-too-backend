"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { useBooking, useBookingMutations } from "@/hooks/use-bookings";
import { useEmployeeAvailability, useService } from "@/hooks/use-catalog";
import {
  buildSlots,
  formatBookingDate,
  todayISO,
  type Slot,
} from "@/lib/booking-slots";
import { cn } from "@/lib/utils";

export default function RescheduleBookingPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const router = useRouter();

  const { data: booking } = useBooking(id);
  const { data: service } = useService(booking?.service_id ?? "");
  const { data: availability, isLoading } = useEmployeeAvailability(booking?.employee_id);
  const { reschedule } = useBookingMutations(id);

  const [date, setDate] = useState("");
  const [slot, setSlot] = useState<Slot | null>(null);
  const [reason, setReason] = useState("");

  const slots = useMemo(
    () => buildSlots(availability ?? [], date, service?.duration_minutes ?? 0),
    [availability, date, service?.duration_minutes],
  );

  return (
    <div className="p-6 md:p-8">
      <Link
        href={`/bookings/${id}`}
        className="mb-4 inline-block text-sm text-muted-foreground hover:text-foreground"
      >
        ← Back to booking
      </Link>

      <Card className="max-w-xl">
        <CardHeader>
          <CardTitle>Reschedule booking</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="date">New date</Label>
            <Input
              id="date"
              type="date"
              min={todayISO()}
              value={date}
              onChange={(e) => {
                setDate(e.target.value);
                setSlot(null);
              }}
            />
          </div>

          {date && (
            isLoading ? (
              <Skeleton className="h-24 w-full" />
            ) : slots.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                No slots for {formatBookingDate(date)}.
              </p>
            ) : (
              <div className="grid grid-cols-3 gap-2">
                {slots.map((s) => (
                  <button
                    key={s.start}
                    type="button"
                    onClick={() => setSlot(s)}
                    className={cn(
                      "rounded-md border px-3 py-2 text-sm",
                      slot?.start === s.start
                        ? "border-primary bg-primary text-primary-foreground"
                        : "hover:bg-muted",
                    )}
                  >
                    {s.start}
                  </button>
                ))}
              </div>
            )
          )}

          <div className="space-y-2">
            <Label htmlFor="reason">Reason (optional)</Label>
            <Input
              id="reason"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
            />
          </div>

          {reschedule.isError && (
            <p className="text-sm text-destructive">
              {(reschedule.error as Error)?.message || "Failed to reschedule"}
            </p>
          )}

          <Button
            disabled={!date || !slot || reschedule.isPending}
            onClick={() => {
              if (!slot) return;
              reschedule.mutate(
                {
                  booking_date: date,
                  start_time: slot.start,
                  end_time: slot.end,
                  reason: reason.trim() || undefined,
                },
                { onSuccess: () => router.push(`/bookings/${id}`) },
              );
            }}
          >
            {reschedule.isPending ? "Saving..." : "Confirm reschedule"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
