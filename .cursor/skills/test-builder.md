# Test Builder

Use when adding tests.

Rules:
- write table-driven tests
- test service layer first
- mock repositories for service tests
- cover success and failure cases
- include validation, unauthorized, not found, conflict, database error
- run go test ./...

Output:
- tests added
- cases covered
- remaining gaps