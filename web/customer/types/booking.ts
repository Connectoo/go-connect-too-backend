// Shapes mirror the Go API DTO internal/modules/bookings/dto.go (BookingResponse
// and CreateBookingRequest). GET /bookings returns a bare array (no pagination
// wrapper) and GET /bookings/{id} returns a single object.

export type BookingStatus =
  | "pending"
  | "accepted"
  | "in_progress"
  | "completed"
  | "rejected"
  | "cancelled"
  | "no_show";

export type Booking = {
  id: string;
  customer_id: string;
  employee_id: string;
  service_id: string;
  booking_date: string;
  start_time: string;
  end_time: string;
  status: BookingStatus;
  customer_notes?: string | null;
  employee_notes?: string | null;
  total_amount: number;
  source_booking_id?: string | null;
  rescheduled_from_id?: string | null;
  created_at: string;
  updated_at: string;
};

// The detail endpoint returns the same shape as a list item.
export type BookingDetail = Booking;

export type BookingListResult = Booking[];

export type CreateBookingInput = {
  service_id: string;
  booking_date: string;
  start_time: string;
  end_time: string;
  customer_notes?: string;
};
