export type Category = {
  id: string;
  name: string;
  description?: string | null;
  is_active: boolean;
  created_at: string;
  updated_at: string;
};

export type Provider = {
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
  rating?: number | null;
  total_reviews: number;
};

export type Service = {
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

export type HomeStats = {
  categories_count: number;
  providers_count: number;
  services_count: number;
};

export type HomeData = {
  featured_categories: Category[];
  featured_providers: Provider[];
  featured_services: Service[];
  stats: HomeStats;
};
