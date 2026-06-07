"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { createReview } from "@/services/review";
import type { CreateReviewInput } from "@/types/review";

export function useCreateReview(bookingId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateReviewInput) => createReview(bookingId, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["bookings", bookingId] });
      qc.invalidateQueries({ queryKey: ["bookings"] });
    },
  });
}
