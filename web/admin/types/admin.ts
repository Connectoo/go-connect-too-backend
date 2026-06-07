import type { PaginatedResult } from "@/types/api";

export type DashboardSummary = {
  total_users: number;
  total_customers: number;
  total_employees: number;
  pending_employees: number;
  total_bookings: number;
  active_bookings: number;
  total_services: number;
  active_services: number;
  total_payments: number;
  completed_payments: number;
  total_revenue: number;
  active_subscriptions: number;
};

export type Category = {
  id: string;
  name: string;
  description?: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type AdminEmployee = {
  id: string;
  user_id: string;
  display_name?: string | null;
  phone?: string | null;
  bio?: string | null;
  experience_years: number;
  verification_status: string;
  location_text?: string | null;
  languages?: string[];
  skills?: string[];
  user_name: string;
  user_email: string;
  user_status: string;
  created_at: string;
  updated_at: string;
};

export type Booking = {
  id: string;
  customer_id: string;
  employee_id: string;
  service_id: string;
  booking_date: string;
  start_time: string;
  end_time: string;
  status: string;
  total_amount: number;
  customer_notes?: string | null;
  employee_notes?: string | null;
  source_booking_id?: string | null;
  rescheduled_from_id?: string | null;
  created_at: string;
  updated_at: string;
};

export const BOOKING_STATUSES = [
  "pending",
  "accepted",
  "in_progress",
  "completed",
  "rejected",
  "cancelled",
  "no_show",
] as const;

export type BookingStatus = (typeof BOOKING_STATUSES)[number];

export type EmployeeListResult = PaginatedResult<AdminEmployee>;
export type BookingListResult = PaginatedResult<Booking>;

export type AdminUser = {
  id: string;
  name: string;
  email: string;
  phone?: string | null;
  role: string;
  status: string;
  created_at: string;
  updated_at: string;
};

export type UserListResult = PaginatedResult<AdminUser>;

export type AdminKYCRecord = {
  id: string;
  employee_id: string;
  id_proof_url: string;
  address_proof_url: string;
  status: string;
  rejection_reason?: string | null;
  reviewed_by?: string | null;
  reviewed_at?: string | null;
  created_at: string;
  updated_at: string;
  employee_display_name?: string | null;
  user_name: string;
  user_email: string;
};

export type KYCListResult = PaginatedResult<AdminKYCRecord>;
