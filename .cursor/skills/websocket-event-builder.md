# WebSocket Event Builder

Use for realtime booking, payment, provider, and notification events.

Rules:
- keep WebSocket logic in platform/websocket
- business modules should publish events, not manage connections
- authenticate WebSocket clients
- define event types clearly
- avoid sending secrets or private data
- support reconnect behavior where needed

Output:
- event names
- payload format
- publisher location
- subscriber/client behavior