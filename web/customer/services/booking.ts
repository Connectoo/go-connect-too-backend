import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type { Booking, CreateBookingInput } from "@/types/booking";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

// GET /bookings returns the authenticated customer's bookings as a bare array.
export function getBookings() {
  return apiRequest<Booking[]>("/bookings", authOptions());
}

export function getBooking(id: string) {
  return apiRequest<Booking>(`/bookings/${id}`, authOptions());
}

export function createBooking(body: CreateBookingInput) {
  return apiRequest<Booking>("/bookings", {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}
