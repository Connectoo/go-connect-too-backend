import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type {
  Booking,
  EmployeeActionRequest,
  RescheduleRequest,
} from "@/types/booking";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchBookings() {
  return apiRequest<Booking[]>("/employee/bookings", authOptions());
}

function action(
  id: string,
  verb: string,
  body?: EmployeeActionRequest,
) {
  return apiRequest<Booking>(`/employee/bookings/${id}/${verb}`, {
    ...authOptions({
      method: "PATCH",
      ...(body ? { body: JSON.stringify(body) } : {}),
    }),
  });
}

export function acceptBooking(id: string) {
  return action(id, "accept");
}

export function rejectBooking(id: string, reason?: string) {
  return action(id, "reject", reason ? { reason } : undefined);
}

export function startBooking(id: string) {
  return action(id, "start");
}

export function completeBooking(id: string) {
  return action(id, "complete");
}

export function cancelBooking(id: string, reason?: string) {
  return action(id, "cancel", reason ? { reason } : undefined);
}

export function noShowBooking(id: string, reason?: string) {
  return action(id, "no-show", reason ? { reason } : undefined);
}

export function rescheduleBooking(id: string, body: RescheduleRequest) {
  return apiRequest<Booking>(`/employee/bookings/${id}/reschedule`, {
    ...authOptions({ method: "PATCH", body: JSON.stringify(body) }),
  });
}
