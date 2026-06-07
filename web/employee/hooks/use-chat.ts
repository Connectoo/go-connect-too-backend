"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  fetchConversations,
  fetchMessages,
  sendMessage,
} from "@/services/chat";
import type { SendMessageInput } from "@/types/chat";
import { useWebSocket } from "@/hooks/use-websocket";

export function useConversations() {
  return useQuery({
    queryKey: ["chat", "conversations"],
    queryFn: fetchConversations,
  });
}

export function useMessages(conversationId: string) {
  const qc = useQueryClient();

  useWebSocket((message) => {
    if (
      message.type === "chat.message" &&
      message.payload?.conversation_id === conversationId
    ) {
      qc.invalidateQueries({
        queryKey: ["chat", "messages", conversationId],
      });
      qc.invalidateQueries({ queryKey: ["chat", "conversations"] });
    }
  });

  return useQuery({
    queryKey: ["chat", "messages", conversationId],
    queryFn: () => fetchMessages(conversationId, { page: 1, limit: 50 }),
    enabled: Boolean(conversationId),
    refetchInterval: 15000,
  });
}

export function useChatMutations(conversationId: string) {
  const qc = useQueryClient();
  const invalidate = () => {
    qc.invalidateQueries({ queryKey: ["chat", "messages", conversationId] });
    qc.invalidateQueries({ queryKey: ["chat", "conversations"] });
  };

  return {
    send: useMutation({
      mutationFn: (body: SendMessageInput) => sendMessage(conversationId, body),
      onSuccess: invalidate,
    }),
  };
}
