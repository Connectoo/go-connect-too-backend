import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type { CreateReviewInput, Review } from "@/types/review";

function authOptions(extra?: RequestInit) {
  return { token: getAccessToken(), ...extra };
}

export function createReview(bookingId: string, body: CreateReviewInput) {
  return apiRequest<Review>(`/bookings/${bookingId}/review`, {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}
