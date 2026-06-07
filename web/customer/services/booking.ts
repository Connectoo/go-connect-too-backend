import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type {
  Booking,
  CancelInput,
  CreateBookingInput,
  RebookInput,
  RebookPreview,
  RescheduleInput,
} from "@/types/booking";

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

export function cancelBooking(id: string, body?: CancelInput) {
  return apiRequest<Booking>(`/bookings/${id}/cancel`, {
    ...authOptions({ method: "PATCH", body: JSON.stringify(body ?? {}) }),
  });
}

export function rescheduleBooking(id: string, body: RescheduleInput) {
  return apiRequest<Booking>(`/bookings/${id}/reschedule`, {
    ...authOptions({ method: "PATCH", body: JSON.stringify(body) }),
  });
}

export function getRebookPreview(id: string) {
  return apiRequest<RebookPreview>(`/bookings/${id}/rebook-preview`, authOptions());
}

export function rebookBooking(body: RebookInput) {
  return apiRequest<Booking>("/bookings/rebook", {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}
