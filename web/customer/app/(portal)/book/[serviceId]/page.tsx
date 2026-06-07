"use client";

import { ArrowLeft } from "lucide-react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { useCreateBooking } from "@/hooks/use-bookings";
import { useEmployeeAvailability, useService } from "@/hooks/use-catalog";
import { cn, formatCurrency } from "@/lib/utils";
import type { AvailabilitySlot } from "@/types/service";

type Slot = { start: string; end: string };

function toMinutes(value: string) {
  const [h, m] = value.split(":").map(Number);
  return h * 60 + m;
}

function toClock(minutes: number) {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}`;
}

// day_of_week from a "YYYY-MM-DD" string, parsed as a local date (0 = Sunday).
function dayOfWeek(date: string) {
  const [y, m, d] = date.split("-").map(Number);
  return new Date(y, m - 1, d).getDay();
}

function buildSlots(
  availability: AvailabilitySlot[],
  date: string,
  durationMinutes: number,
): Slot[] {
  if (!date || durationMinutes <= 0) return [];
  const dow = dayOfWeek(date);
  const slots: Slot[] = [];
  availability
    .filter((a) => a.is_available && a.day_of_week === dow)
    .forEach((window) => {
      const windowEnd = toMinutes(window.end_time);
      let cursor = toMinutes(window.start_time);
      while (cursor + durationMinutes <= windowEnd) {
        slots.push({
          start: toClock(cursor),
          end: toClock(cursor + durationMinutes),
        });
        cursor += durationMinutes;
      }
    });
  return slots.sort((a, b) => toMinutes(a.start) - toMinutes(b.start));
}

function formatDate(value: string) {
  const [y, m, d] = value.split("-").map(Number);
  if (!y || !m || !d) return value;
  return new Date(y, m - 1, d).toLocaleDateString(undefined, {
    weekday: "long",
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

const STEPS = ["Date", "Time", "Notes", "Confirm"];

export default function BookServicePage() {
  const params = useParams<{ serviceId: string }>();
  const serviceId = params.serviceId;
  const router = useRouter();

  const { data: service, isLoading: serviceLoading, isError: serviceError } =
    useService(serviceId);
  const { data: availability, isLoading: availabilityLoading } =
    useEmployeeAvailability(service?.employee_id);
  const createBooking = useCreateBooking();

  const [step, setStep] = useState(0);
  const [date, setDate] = useState("");
  const [slot, setSlot] = useState<Slot | null>(null);
  const [notes, setNotes] = useState("");

  const today = useMemo(() => {
    const now = new Date();
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(
      now.getDate(),
    ).padStart(2, "0")}`;
  }, []);

  const slots = useMemo(
    () => buildSlots(availability ?? [], date, service?.duration_minutes ?? 0),
    [availability, date, service?.duration_minutes],
  );

  if (serviceLoading) {
    return (
      <div className="space-y-4 p-6 md:p-8">
        <Skeleton className="h-8 w-48" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (serviceError || !service) {
    return (
      <div className="p-6 md:p-8">
        <div className="rounded-lg border p-8 text-center text-sm text-destructive">
          This service is unavailable.
        </div>
      </div>
    );
  }

  function submit() {
    if (!slot) return;
    createBooking.mutate(
      {
        service_id: serviceId,
        booking_date: date,
        start_time: slot.start,
        end_time: slot.end,
        customer_notes: notes.trim() || undefined,
      },
      {
        onSuccess: (booking) => router.push(`/bookings/${booking.id}`),
      },
    );
  }

  const canNext =
    (step === 0 && Boolean(date)) ||
    (step === 1 && Boolean(slot)) ||
    step === 2;

  return (
    <div className="p-6 md:p-8">
      <Link
        href="/bookings"
        className="mb-4 inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-4 w-4" /> Bookings
      </Link>

      <div className="mb-6">
        <h1 className="text-2xl font-bold">Book {service.title}</h1>
        <p className="text-sm text-muted-foreground">
          {service.duration_minutes} min · {formatCurrency(service.price)}
        </p>
      </div>

      <div className="mb-6 flex items-center gap-2">
        {STEPS.map((label, i) => (
          <div key={label} className="flex items-center gap-2">
            <span
              className={cn(
                "flex h-7 w-7 items-center justify-center rounded-full text-xs font-medium",
                i === step
                  ? "bg-primary text-primary-foreground"
                  : i < step
                    ? "bg-green-600 text-white"
                    : "bg-muted text-muted-foreground",
              )}
            >
              {i + 1}
            </span>
            <span
              className={cn(
                "text-sm",
                i === step ? "font-medium" : "text-muted-foreground",
              )}
            >
              {label}
            </span>
            {i < STEPS.length - 1 ? (
              <span className="mx-1 h-px w-6 bg-border" />
            ) : null}
          </div>
        ))}
      </div>

      <Card className="max-w-2xl">
        <CardHeader>
          <CardTitle>{STEPS[step]}</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {step === 0 ? (
            <div className="space-y-2">
              <Label htmlFor="date">Choose a date</Label>
              <Input
                id="date"
                type="date"
                min={today}
                value={date}
                onChange={(e) => {
                  setDate(e.target.value);
                  setSlot(null);
                }}
              />
            </div>
          ) : null}

          {step === 1 ? (
            availabilityLoading ? (
              <div className="grid grid-cols-3 gap-2">
                {Array.from({ length: 6 }).map((_, i) => (
                  <Skeleton key={i} className="h-10 w-full" />
                ))}
              </div>
            ) : slots.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                No available time slots for {formatDate(date)}. Try another date.
              </p>
            ) : (
              <div className="grid grid-cols-3 gap-2">
                {slots.map((s) => (
                  <button
                    key={s.start}
                    type="button"
                    onClick={() => setSlot(s)}
                    className={cn(
                      "rounded-md border px-3 py-2 text-sm transition-colors",
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
          ) : null}

          {step === 2 ? (
            <div className="space-y-2">
              <Label htmlFor="notes">Notes (optional)</Label>
              <textarea
                id="notes"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                rows={4}
                placeholder="Anything the provider should know?"
                className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
              />
            </div>
          ) : null}

          {step === 3 ? (
            <div className="space-y-3 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Service</span>
                <span className="font-medium">{service.title}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Date</span>
                <span className="font-medium">{formatDate(date)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Time</span>
                <span className="font-medium">
                  {slot ? `${slot.start} – ${slot.end}` : "—"}
                </span>
              </div>
              {notes.trim() ? (
                <div className="flex justify-between gap-4">
                  <span className="text-muted-foreground">Notes</span>
                  <span className="text-right font-medium">{notes.trim()}</span>
                </div>
              ) : null}
              <div className="flex justify-between border-t pt-3">
                <span className="text-muted-foreground">Total</span>
                <span className="font-semibold">{formatCurrency(service.price)}</span>
              </div>
              {createBooking.isError ? (
                <p className="text-sm text-destructive">
                  {(createBooking.error as Error)?.message ||
                    "Could not create booking."}
                </p>
              ) : null}
            </div>
          ) : null}

          <div className="flex justify-between pt-2">
            <Button
              variant="outline"
              onClick={() => setStep((s) => Math.max(0, s - 1))}
              disabled={step === 0 || createBooking.isPending}
            >
              Back
            </Button>
            {step < STEPS.length - 1 ? (
              <Button onClick={() => setStep((s) => s + 1)} disabled={!canNext}>
                Next
              </Button>
            ) : (
              <Button onClick={submit} disabled={!slot || createBooking.isPending}>
                {createBooking.isPending ? "Booking..." : "Confirm booking"}
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
