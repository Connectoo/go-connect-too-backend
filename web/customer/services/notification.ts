import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type { NotificationListResult } from "@/types/notification";

function authOptions(extra?: RequestInit) {
  return { token: getAccessToken(), ...extra };
}

export function fetchNotifications(params?: { page?: number; limit?: number }) {
  return apiRequest<NotificationListResult>("/notifications", {
    ...authOptions(),
    params,
  });
}

export function markNotificationRead(id: string) {
  return apiRequest(`/notifications/${id}/read`, {
    ...authOptions({ method: "PATCH" }),
  });
}

export function markAllNotificationsRead() {
  return apiRequest<{ updated: number }>("/notifications/read-all", {
    ...authOptions({ method: "PATCH" }),
  });
}
