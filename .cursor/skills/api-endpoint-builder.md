# API Endpoint Builder

Use when adding one REST endpoint.

Inputs:
- method
- path
- auth required
- role required
- request body
- response body
- business rules

Rules:
- follow /api/v1 format
- update routes.go
- keep handler thin
- put business logic in service.go
- put database work in repository.go
- use standard response format
- return proper HTTP status codes

Output:
- files changed
- endpoint added
- example request
- example response
- tests needed