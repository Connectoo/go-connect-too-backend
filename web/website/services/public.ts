import { apiRequest } from "@/lib/api-client";
import {
  mockCategoryList,
  mockHome,
  mockProviderById,
  mockProviderList,
} from "@/lib/mocks";
import type { Category, HomeData, Provider, Service } from "@/types/public";

async function withMockFallback<T>(fn: () => Promise<T>, fallback: T): Promise<T> {
  try {
    return await fn();
  } catch {
    return fallback;
  }
}

export function fetchHome() {
  return withMockFallback(
    () => apiRequest<HomeData>("/public/home"),
    mockHome,
  );
}

export function fetchCategories() {
  return withMockFallback(
    () => apiRequest<Category[]>("/public/categories"),
    mockCategoryList,
  );
}

export function fetchProviders(limit = 20) {
  return withMockFallback(
    () => apiRequest<Provider[]>("/public/providers", { params: { limit } }),
    mockProviderList,
  );
}

export function fetchProvider(id: string) {
  return withMockFallback(
    () => apiRequest<Provider>(`/public/providers/${id}`),
    mockProviderById(id) ?? mockProviderList[0],
  );
}

export function fetchServices(params?: { category_id?: string; limit?: number }) {
  return withMockFallback(
    () => apiRequest<Service[]>("/public/services", { params }),
    mockHome.featured_services,
  );
}

export function fetchService(id: string) {
  return withMockFallback(
    () => apiRequest<Service>(`/public/services/${id}`),
    mockHome.featured_services[0],
  );
}
