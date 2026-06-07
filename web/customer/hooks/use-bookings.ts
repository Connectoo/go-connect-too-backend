"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  cancelBooking,
  createBooking,
  getBooking,
  getBookings,
  getRebookPreview,
  rebookBooking,
  rescheduleBooking,
} from "@/services/booking";
import type {
  CancelInput,
  CreateBookingInput,
  RebookInput,
  RescheduleInput,
} from "@/types/booking";

export function useBookings() {
  return useQuery({
    queryKey: ["bookings"],
    queryFn: getBookings,
  });
}

export function useBooking(id: string) {
  return useQuery({
    queryKey: ["bookings", id],
    queryFn: () => getBooking(id),
    enabled: Boolean(id),
  });
}

export function useRebookPreview(id: string) {
  return useQuery({
    queryKey: ["bookings", id, "rebook-preview"],
    queryFn: () => getRebookPreview(id),
    enabled: Boolean(id),
  });
}

export function useCreateBooking() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createBooking,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["bookings"] }),
  });
}

export function useBookingMutations(id: string) {
  const qc = useQueryClient();
  const invalidate = () => {
    qc.invalidateQueries({ queryKey: ["bookings"] });
    qc.invalidateQueries({ queryKey: ["bookings", id] });
  };

  return {
    cancel: useMutation({
      mutationFn: (body?: CancelInput) => cancelBooking(id, body),
      onSuccess: invalidate,
    }),
    reschedule: useMutation({
      mutationFn: (body: RescheduleInput) => rescheduleBooking(id, body),
      onSuccess: invalidate,
    }),
    rebook: useMutation({
      mutationFn: (body: RebookInput) => rebookBooking(body),
      onSuccess: () => qc.invalidateQueries({ queryKey: ["bookings"] }),
    }),
  };
}
