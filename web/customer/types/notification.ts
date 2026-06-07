import type { PaginatedResult } from "@/types/api";

export type Notification = {
  id: string;
  type: string;
  title: string;
  body: string;
  data: Record<string, unknown>;
  read_at?: string | null;
  created_at: string;
};

export type NotificationListResult = PaginatedResult<Notification>;
