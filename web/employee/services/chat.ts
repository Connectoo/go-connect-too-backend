import { apiRequest } from "@/lib/api-client";
import { getAccessToken } from "@/lib/auth";
import type {
  Conversation,
  MessageListResult,
  SendMessageInput,
} from "@/types/chat";

function authOptions(extra?: RequestInit) {
  return { token: getAccessToken(), ...extra };
}

export function fetchConversations() {
  return apiRequest<Conversation[]>("/chat/conversations", authOptions());
}

export function fetchMessages(
  conversationId: string,
  params?: { page?: number; limit?: number },
) {
  return apiRequest<MessageListResult>(
    `/chat/conversations/${conversationId}/messages`,
    { ...authOptions(), params },
  );
}

export function sendMessage(conversationId: string, body: SendMessageInput) {
  return apiRequest(`/chat/conversations/${conversationId}/messages`, {
    ...authOptions({ method: "POST", body: JSON.stringify(body) }),
  });
}

export function markMessageRead(conversationId: string, messageId: string) {
  return apiRequest(
    `/chat/conversations/${conversationId}/messages/${messageId}/read`,
    { ...authOptions({ method: "PATCH" }) },
  );
}
