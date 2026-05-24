# Cursor Prompts

## Create New Module

Create a new module named [module_name] in internal/modules.

Use this structure:
- handler.go
- service.go
- repository.go
- model.go
- dto.go
- routes.go

Follow project rules:
- handler only handles HTTP
- service contains business logic
- repository handles DB only
- use context.Context
- use standard API response format

## Add API Endpoint

Add endpoint:

Method:
Path:
Auth required:
Role:
Request body:
Response:
Business rules:

Follow existing code style and update routes.

## Create Database Migration

Create PostgreSQL migration for:

Table:
Fields:
Relations:
Indexes:
Constraints:

Also explain why each index is needed.

## Review Code

Review this code for:
- Go best practices
- security issues
- architecture violations
- database transaction problems
- missing tests
- performance problems

Give fixes with file paths.

## Add Tests

Write tests for this module.

Include:
- success case
- validation error
- unauthorized case
- not found case
- database error case

Use table-driven tests.

## Refactor Safely

Refactor this code without changing behavior.

Rules:
- keep public API same
- improve readability
- reduce duplication
- add small helper functions only when useful
- run gofmt