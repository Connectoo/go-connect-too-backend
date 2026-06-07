"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  acceptBooking,
  cancelBooking,
  completeBooking,
  fetchBookings,
  noShowBooking,
  rejectBooking,
  rescheduleBooking,
  startBooking,
} from "@/services/booking";
import type { RescheduleRequest } from "@/types/booking";

const BOOKINGS_KEY = ["employee", "bookings"] as const;

export function useBookings() {
  return useQuery({
    queryKey: BOOKINGS_KEY,
    queryFn: fetchBookings,
  });
}

export function useBookingActions() {
  const qc = useQueryClient();
  const invalidate = () => qc.invalidateQueries({ queryKey: BOOKINGS_KEY });

  return {
    accept: useMutation({
      mutationFn: (id: string) => acceptBooking(id),
      onSuccess: invalidate,
    }),
    reject: useMutation({
      mutationFn: ({ id, reason }: { id: string; reason?: string }) =>
        rejectBooking(id, reason),
      onSuccess: invalidate,
    }),
    start: useMutation({
      mutationFn: (id: string) => startBooking(id),
      onSuccess: invalidate,
    }),
    complete: useMutation({
      mutationFn: (id: string) => completeBooking(id),
      onSuccess: invalidate,
    }),
    cancel: useMutation({
      mutationFn: ({ id, reason }: { id: string; reason?: string }) =>
        cancelBooking(id, reason),
      onSuccess: invalidate,
    }),
    noShow: useMutation({
      mutationFn: ({ id, reason }: { id: string; reason?: string }) =>
        noShowBooking(id, reason),
      onSuccess: invalidate,
    }),
    reschedule: useMutation({
      mutationFn: ({ id, body }: { id: string; body: RescheduleRequest }) =>
        rescheduleBooking(id, body),
      onSuccess: invalidate,
    }),
  };
}
