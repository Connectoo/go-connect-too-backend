import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type {
  EmployeeBookingsAnalytics,
  EmployeeReviewsAnalytics,
  EmployeeSummary,
} from "@/types/analytics";

function authOptions(extra?: RequestInit) {
  return {
    token: getAccessToken(),
    ...extra,
  };
}

export function fetchSummary() {
  return apiRequest<EmployeeSummary>("/employee/analytics/summary", authOptions());
}

export function fetchBookingsAnalytics() {
  return apiRequest<EmployeeBookingsAnalytics>(
    "/employee/analytics/bookings",
    authOptions(),
  );
}

export function fetchReviewsAnalytics() {
  return apiRequest<EmployeeReviewsAnalytics>(
    "/employee/analytics/reviews",
    authOptions(),
  );
}
