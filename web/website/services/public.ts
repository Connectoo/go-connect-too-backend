import { apiRequest } from "@/lib/api-client";
import type { Category, HomeData, Provider, Service } from "@/types/public";

export function fetchHome() {
  return apiRequest<HomeData>("/public/home");
}

export function fetchCategories() {
  return apiRequest<Category[]>("/public/categories");
}

export function fetchProviders(limit = 20) {
  return apiRequest<Provider[]>("/public/providers", { params: { limit } });
}

export function fetchProvider(id: string) {
  return apiRequest<Provider>(`/public/providers/${id}`);
}

export function fetchServices(params?: { category_id?: string; limit?: number }) {
  return apiRequest<Service[]>("/public/services", { params });
}

export function fetchService(id: string) {
  return apiRequest<Service>(`/public/services/${id}`);
}
