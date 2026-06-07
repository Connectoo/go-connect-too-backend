export type VerificationStatus = "pending" | "approved" | "rejected";

// Mirrors EmployeeProfileResponse in internal/app/spec/openapi.yaml (~line 2888).
export type EmployeeProfile = {
  id: string;
  user_id: string;
  display_name?: string | null;
  phone?: string | null;
  bio?: string | null;
  experience_years: number;
  profile_photo_url?: string | null;
  location_text?: string | null;
  latitude?: number | null;
  longitude?: number | null;
  service_area_radius_km?: number | null;
  languages: string[];
  skills: string[];
  verification_status: VerificationStatus;
  created_at: string;
  updated_at: string;
};

// Mirrors UpdateEmployeeProfileRequest in openapi.yaml (~line 2946).
export type UpdateEmployeeProfileRequest = {
  display_name: string;
  phone: string;
  bio?: string | null;
  experience_years?: number;
  profile_photo_url?: string | null;
  location_text?: string | null;
  latitude?: number;
  longitude?: number;
  service_area_radius_km?: number;
  languages?: string[];
  skills?: string[];
};

export type ChangePasswordRequest = {
  current_password: string;
  new_password: string;
};
