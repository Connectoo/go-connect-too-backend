import type { AvailabilitySlot } from "@/types/service";

export type Slot = { start: string; end: string };

export function toMinutes(value: string) {
  const [h, m] = value.split(":").map(Number);
  return h * 60 + m;
}

export function toClock(minutes: number) {
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return `${String(h).padStart(2, "0")}:${String(m).padStart(2, "0")}`;
}

export function dayOfWeek(date: string) {
  const [y, m, d] = date.split("-").map(Number);
  return new Date(y, m - 1, d).getDay();
}

export function buildSlots(
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

export function formatBookingDate(value: string) {
  const [y, m, d] = value.split("-").map(Number);
  if (!y || !m || !d) return value;
  return new Date(y, m - 1, d).toLocaleDateString(undefined, {
    weekday: "long",
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

export function todayISO() {
  const now = new Date();
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, "0")}-${String(
    now.getDate(),
  ).padStart(2, "0")}`;
}
