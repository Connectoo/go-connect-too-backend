// Shapes mirror the Go API DTOs (internal/modules/services/dto.go ServiceResponse
// and internal/modules/availability/dto.go AvailabilityResponse). The OpenAPI spec
// only documents a generic SuccessEnvelope, so the DTOs are the source of truth.

export type ServiceDetail = {
  id: string;
  employee_id: string;
  category_id: string;
  title: string;
  description?: string | null;
  price: number;
  duration_minutes: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

// Weekly recurring availability slot. day_of_week is 0 (Sunday) .. 6 (Saturday),
// matching Go's time.Weekday. start_time / end_time are "HH:MM" clock times.
export type AvailabilitySlot = {
  id: string;
  employee_id: string;
  day_of_week: number;
  start_time: string;
  end_time: string;
  is_available: boolean;
  created_at: string;
  updated_at: string;
};
