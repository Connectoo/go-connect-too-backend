import type { PaginatedResult } from "@/types/api";

export type Conversation = {
  id: string;
  customer_id: string;
  employee_id: string;
  booking_id?: string | null;
  created_at: string;
  updated_at: string;
};

export type ChatMessage = {
  id: string;
  conversation_id: string;
  sender_id: string;
  message: string;
  attachment_url?: string | null;
  content_type?: string | null;
  read_at?: string | null;
  created_at: string;
};

export type MessageListResult = PaginatedResult<ChatMessage>;

export type SendMessageInput = {
  message: string;
  attachment_url?: string;
  content_type?: string;
};
