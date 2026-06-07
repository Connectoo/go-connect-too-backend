export type WebSocketMessage = {
  type: string;
  payload: Record<string, unknown>;
};
