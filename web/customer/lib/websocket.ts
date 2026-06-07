import { getAccessToken } from "@/lib/auth";
import type { WebSocketMessage } from "@/types/websocket";

export function buildWebSocketUrl(token: string) {
  const base =
    process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";
  const wsBase = base.replace(/^http/, "ws");
  return `${wsBase}/ws?token=${encodeURIComponent(token)}`;
}

export function parseWebSocketMessage(raw: string): WebSocketMessage | null {
  try {
    return JSON.parse(raw) as WebSocketMessage;
  } catch {
    return null;
  }
}

export function getWebSocketToken() {
  return getAccessToken();
}
