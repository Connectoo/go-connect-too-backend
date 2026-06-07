"use client";

import { useState } from "react";
import { Trash2 } from "lucide-react";
import { ConfirmDialog } from "@/components/shared/confirm-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import {
  useAvailability,
  useAvailabilityMutations,
} from "@/hooks/use-availability";
import type { Availability } from "@/types/availability";

// day_of_week matches Go time.Weekday: Sunday = 0 .. Saturday = 6.
const WEEKDAYS = [
  "Sunday",
  "Monday",
  "Tuesday",
  "Wednesday",
  "Thursday",
  "Friday",
  "Saturday",
];

export default function AvailabilityPage() {
  const { data, isLoading, isError, error } = useAvailability();
  const { create, update, remove } = useAvailabilityMutations();

  const [day, setDay] = useState(1);
  const [startTime, setStartTime] = useState("09:00");
  const [endTime, setEndTime] = useState("17:00");
  const [formError, setFormError] = useState<string | null>(null);
  const [deleteId, setDeleteId] = useState<string | null>(null);

  const slotsByDay = (d: number): Availability[] =>
    (data ?? [])
      .filter((s) => s.day_of_week === d)
      .sort((a, b) => a.start_time.localeCompare(b.start_time));

  const handleAdd = () => {
    setFormError(null);
    if (startTime >= endTime) {
      setFormError("End time must be after start time.");
      return;
    }
    create.mutate(
      { day_of_week: day, start_time: startTime, end_time: endTime },
      {
        onError: (err) =>
          setFormError((err as Error)?.message || "Failed to add slot"),
      },
    );
  };

  return (
    <div className="p-6 md:p-8">
      <div className="mb-6">
        <h1 className="text-2xl font-bold">Availability</h1>
        <p className="text-sm text-muted-foreground">
          Set the weekly hours you accept bookings.
        </p>
      </div>

      <Card className="mb-6 max-w-3xl">
        <CardHeader>
          <CardTitle>Add a slot</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid items-end gap-4 md:grid-cols-4">
            <div className="space-y-2">
              <Label htmlFor="day">Day</Label>
              <select
                id="day"
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                value={day}
                onChange={(e) => setDay(Number(e.target.value))}
              >
                {WEEKDAYS.map((label, index) => (
                  <option key={label} value={index}>
                    {label}
                  </option>
                ))}
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="start">Start</Label>
              <input
                id="start"
                type="time"
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                value={startTime}
                onChange={(e) => setStartTime(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="end">End</Label>
              <input
                id="end"
                type="time"
                className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                value={endTime}
                onChange={(e) => setEndTime(e.target.value)}
              />
            </div>
            <Button onClick={handleAdd} disabled={create.isPending}>
              {create.isPending ? "Adding..." : "Add slot"}
            </Button>
          </div>
          {formError && (
            <p className="mt-3 text-sm text-destructive">{formError}</p>
          )}
        </CardContent>
      </Card>

      {isError ? (
        <div className="rounded-lg border border-destructive/40 p-8 text-center text-sm text-destructive">
          {(error as Error)?.message || "Failed to load availability"}
        </div>
      ) : isLoading ? (
        <div className="rounded-lg border p-8 text-center text-sm text-muted-foreground">
          Loading availability...
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {WEEKDAYS.map((label, index) => {
            const slots = slotsByDay(index);
            return (
              <Card key={label}>
                <CardHeader>
                  <CardTitle className="text-base">{label}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  {slots.length === 0 ? (
                    <p className="text-sm text-muted-foreground">
                      No slots.
                    </p>
                  ) : (
                    slots.map((slot) => (
                      <div
                        key={slot.id}
                        className="flex items-center justify-between gap-2 rounded-md border px-3 py-2 text-sm"
                      >
                        <span>
                          {slot.start_time} – {slot.end_time}
                        </span>
                        <div className="flex items-center gap-2">
                          <button
                            type="button"
                            onClick={() =>
                              update.mutate({
                                id: slot.id,
                                body: {
                                  day_of_week: slot.day_of_week,
                                  start_time: slot.start_time,
                                  end_time: slot.end_time,
                                  is_available: !slot.is_available,
                                },
                              })
                            }
                            disabled={update.isPending}
                          >
                            <Badge
                              variant={
                                slot.is_available ? "default" : "secondary"
                              }
                            >
                              {slot.is_available ? "Available" : "Off"}
                            </Badge>
                          </button>
                          <Button
                            size="sm"
                            variant="ghost"
                            onClick={() => setDeleteId(slot.id)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    ))
                  )}
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      <ConfirmDialog
        open={Boolean(deleteId)}
        onOpenChange={(open) => !open && setDeleteId(null)}
        title="Delete slot?"
        description="This removes the availability slot."
        confirmLabel="Delete"
        variant="destructive"
        loading={remove.isPending}
        onConfirm={async () => {
          if (!deleteId) return;
          await remove.mutateAsync(deleteId);
          setDeleteId(null);
        }}
      />
    </div>
  );
}
