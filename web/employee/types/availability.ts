// Mirrors AvailabilityResponse in internal/modules/availability/dto.go.
// day_of_week: 0 (Sunday) .. 6 (Saturday). start_time/end_time are "HH:MM".
export type Availability = {
  id: string;
  employee_id: string;
  day_of_week: number;
  start_time: string;
  end_time: string;
  is_available: boolean;
  created_at: string;
  updated_at: string;
};

// Mirrors CreateAvailabilityRequest in internal/modules/availability/dto.go.
export type CreateAvailabilityRequest = {
  day_of_week: number;
  start_time: string;
  end_time: string;
  is_available?: boolean;
};

// Mirrors UpdateAvailabilityRequest in internal/modules/availability/dto.go.
export type UpdateAvailabilityRequest = {
  day_of_week: number;
  start_time: string;
  end_time: string;
  is_available: boolean;
};
