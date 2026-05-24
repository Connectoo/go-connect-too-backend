# Security Reviewer

Use before auth, payment, admin, or deployment changes.

Review for:
- authentication
- authorization
- JWT handling
- password hashing
- input validation
- SQL injection
- secret leakage
- unsafe logs
- webhook verification
- admin access control

Output:
- severity
- file path
- issue
- fix
- whether it blocks release