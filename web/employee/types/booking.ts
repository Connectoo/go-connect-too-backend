export type BookingStatus =
  | "pending"
  | "accepted"
  | "rejected"
  | "in_progress"
  | "completed"
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

export type EmployeeActionRequest = {
  employee_notes?: string;
  reason?: string;
};

export type RescheduleRequest = {
  booking_date: string;
  start_time: string;
  end_time: string;
  reason?: string;
};

export type BookingTab = "pending" | "active" | "completed";

export function tabForStatus(status: BookingStatus): BookingTab {
  switch (status) {
    case "pending":
      return "pending";
    case "accepted":
    case "in_progress":
      return "active";
    default:
      return "completed";
  }
}
