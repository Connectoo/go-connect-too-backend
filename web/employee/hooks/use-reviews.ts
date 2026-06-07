"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchReviews, replyToReview } from "@/services/review";
import type { ReplyInput } from "@/types/review";

export function useReviews(page = 1, limit = 20) {
  return useQuery({
    queryKey: ["employee", "reviews", page, limit],
    queryFn: () => fetchReviews({ page, limit }),
  });
}

export function useReviewReply(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: ReplyInput) => replyToReview(id, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["employee", "reviews"] });
    },
  });
}
