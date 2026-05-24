# Push Notification Builder

Use for Firebase Cloud Messaging, email, SMS, and in-app notifications.

Rules:
- keep notification provider logic in platform/notifications
- keep templates separate
- store notification records when needed
- support retry-safe sending
- do not block main request on slow external notification calls
- avoid sending sensitive data in push payloads

Output:
- notification trigger
- template used
- delivery logic
- retry/failure behavior