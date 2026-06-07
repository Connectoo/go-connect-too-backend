import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type { ReplyInput, ReviewListResult } from "@/types/review";

function authOptions(extra?: RequestInit) {
  return { token: getAccessToken(), ...extra };
}

export function fetchReviews(params?: { page?: number; limit?: number }) {
  return apiRequest<ReviewListResult>("/employee/reviews", {
    ...authOptions(),
    params,
  });
}

export function replyToReview(id: string, body: ReplyInput) {
  return apiRequest(`/employee/reviews/${id}/reply`, {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}
