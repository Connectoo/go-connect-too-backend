"use client";

import { useCallback, useEffect, useRef } from "react";
import { buildWebSocketUrl, getWebSocketToken, parseWebSocketMessage } from "@/lib/websocket";
import type { WebSocketMessage } from "@/types/websocket";

export function useWebSocket(onMessage?: (message: WebSocketMessage) => void) {
  const handlerRef = useRef(onMessage);
  handlerRef.current = onMessage;

  const stableHandler = useCallback((message: WebSocketMessage) => {
    handlerRef.current?.(message);
  }, []);

  useEffect(() => {
    const token = getWebSocketToken();
    if (!token) return;

    const ws = new WebSocket(buildWebSocketUrl(token));

    ws.onmessage = (event) => {
      const message = parseWebSocketMessage(event.data);
      if (message) stableHandler(message);
    };

    return () => ws.close();
  }, [stableHandler]);
}
