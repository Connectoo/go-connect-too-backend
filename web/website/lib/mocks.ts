import type { Category, HomeData, Provider, Service } from "@/types/public";

const mockCategories: Category[] = [
  {
    id: "00000000-0000-4000-8000-000000000001",
    name: "Home Cleaning",
    description: "Professional home cleaning services",
    is_active: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "00000000-0000-4000-8000-000000000002",
    name: "Plumbing",
    description: "Repairs and installations",
    is_active: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
  {
    id: "00000000-0000-4000-8000-000000000003",
    name: "Electrical",
    description: "Licensed electrical work",
    is_active: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
];

const mockProviders: Provider[] = [
  {
    id: "00000000-0000-4000-8000-000000000101",
    display_name: "Priya Sharma",
    bio: "Experienced home service professional with 8+ years in the field.",
    experience_years: 8,
    location_text: "Mumbai, IN",
    languages: ["English", "Hindi"],
    skills: ["Deep cleaning", "Sanitization"],
    rating: 4.8,
    total_reviews: 124,
  },
  {
    id: "00000000-0000-4000-8000-000000000102",
    display_name: "Rahul Verma",
    bio: "Certified plumber serving residential and small commercial clients.",
    experience_years: 6,
    location_text: "Delhi, IN",
    languages: ["English", "Hindi"],
    skills: ["Pipe repair", "Installation"],
    rating: 4.6,
    total_reviews: 89,
  },
];

const mockServices: Service[] = [
  {
    id: "00000000-0000-4000-8000-000000000201",
    employee_id: mockProviders[0].id,
    category_id: mockCategories[0].id,
    title: "Standard Home Cleaning",
    description: "2-hour standard cleaning for apartments up to 2 BHK.",
    price: 1499,
    duration_minutes: 120,
    is_active: true,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
  },
];

export const mockHome: HomeData = {
  featured_categories: mockCategories,
  featured_providers: mockProviders,
  featured_services: mockServices,
  stats: {
    categories_count: mockCategories.length,
    providers_count: mockProviders.length,
    services_count: mockServices.length,
  },
};

export const mockCategoryList = mockCategories;
export const mockProviderList = mockProviders;

export function mockProviderById(id: string): Provider | undefined {
  return mockProviders.find((p) => p.id === id);
}
