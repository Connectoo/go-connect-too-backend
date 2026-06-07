"use client";

import { useQuery } from "@tanstack/react-query";
import {
  fetchBookingsAnalytics,
  fetchReviewsAnalytics,
  fetchSummary,
} from "@/services/analytics";

export function useEmployeeSummary() {
  return useQuery({
    queryKey: ["employee", "analytics", "summary"],
    queryFn: fetchSummary,
  });
}

export function useEmployeeBookingsAnalytics() {
  return useQuery({
    queryKey: ["employee", "analytics", "bookings"],
    queryFn: fetchBookingsAnalytics,
  });
}

export function useEmployeeReviewsAnalytics() {
  return useQuery({
    queryKey: ["employee", "analytics", "reviews"],
    queryFn: fetchReviewsAnalytics,
  });
}
