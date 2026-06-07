"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { createBooking, getBooking, getBookings } from "@/services/booking";

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

export function useCreateBooking() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: createBooking,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["bookings"] }),
  });
}
