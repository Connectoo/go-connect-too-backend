"use client";

import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import {
  useBooking,
  useBookingMutations,
  useRebookPreview,
} from "@/hooks/use-bookings";
import { useEmployeeAvailability, useService } from "@/hooks/use-catalog";
import {
  buildSlots,
  formatBookingDate,
  todayISO,
  type Slot,
} from "@/lib/booking-slots";
import { cn, formatCurrency } from "@/lib/utils";

export default function RebookPage() {
  const params = useParams<{ id: string }>();
  const id = params.id;
  const router = useRouter();

  const { data: booking } = useBooking(id);
  const { data: preview, isLoading: previewLoading } = useRebookPreview(id);
  const { data: service } = useService(preview?.service_id ?? booking?.service_id ?? "");
  const { data: availability, isLoading: slotsLoading } = useEmployeeAvailability(
    preview?.employee_id ?? booking?.employee_id,
  );
  const { rebook } = useBookingMutations(id);

  const [date, setDate] = useState("");
  const [slot, setSlot] = useState<Slot | null>(null);
  const [notes, setNotes] = useState("");

  useEffect(() => {
    if (!preview) return;
    if (preview.suggested_booking_date) setDate(preview.suggested_booking_date);
    if (preview.suggested_start_time && preview.suggested_end_time) {
      setSlot({
        start: preview.suggested_start_time,
        end: preview.suggested_end_time,
      });
    }
  }, [preview]);

  const slots = useMemo(
    () => buildSlots(availability ?? [], date, service?.duration_minutes ?? 0),
    [availability, date, service?.duration_minutes],
  );

  if (previewLoading) {
    return (
      <div className="p-6 md:p-8">
        <Skeleton className="h-64 w-full max-w-xl" />
      </div>
    );
  }

  if (!preview?.can_rebook) {
    return (
      <div className="p-6 md:p-8">
        <p className="text-sm text-muted-foreground">
          This booking cannot be rebooked. The service or provider may no longer be
          available.
        </p>
        <Link href={`/bookings/${id}`} className="mt-4 inline-block text-sm underline">
          Back to booking
        </Link>
      </div>
    );
  }

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
          <CardTitle>Rebook service</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {service && (
            <p className="text-sm text-muted-foreground">
              {service.title} · {formatCurrency(preview.current_price ?? service.price)}
            </p>
          )}

          <div className="space-y-2">
            <Label htmlFor="date">Date</Label>
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

          {date &&
            (slotsLoading ? (
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
            ))}

          <div className="space-y-2">
            <Label htmlFor="notes">Notes (optional)</Label>
            <textarea
              id="notes"
              rows={3}
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm"
            />
          </div>

          {rebook.isError && (
            <p className="text-sm text-destructive">
              {(rebook.error as Error)?.message || "Failed to rebook"}
            </p>
          )}

          <Button
            disabled={!date || !slot || rebook.isPending}
            onClick={() => {
              if (!slot) return;
              rebook.mutate(
                {
                  source_booking_id: id,
                  booking_date: date,
                  start_time: slot.start,
                  end_time: slot.end,
                  customer_notes: notes.trim() || undefined,
                },
                {
                  onSuccess: (newBooking) =>
                    router.push(`/bookings/${newBooking.id}`),
                },
              );
            }}
          >
            {rebook.isPending ? "Booking..." : "Confirm rebook"}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
