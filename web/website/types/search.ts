export type ServiceSearchItem = {
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
  employee_display_name?: string | null;
  employee_location?: string | null;
  rating?: number | null;
};

export type EmployeeSearchItem = {
  id: string;
  display_name?: string | null;
  bio?: string | null;
  experience_years: number;
  profile_photo_url?: string | null;
  location_text?: string | null;
  latitude?: number | null;
  longitude?: number | null;
  service_area_radius_km?: number | null;
  languages: string[];
  skills: string[];
  distance_km?: number | null;
  rating?: number | null;
};
