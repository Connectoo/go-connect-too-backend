# Database Migration Builder

Use when creating PostgreSQL migrations.

Inputs:
- table name
- fields
- relations
- indexes
- constraints

Rules:
- create up/down migration files
- use UUID primary keys
- include created_at and updated_at
- add foreign key indexes
- use NOT NULL where required
- explain every important index
- never drop data without warning

Output:
- migration files created
- schema summary
- rollback behavior
- sqlc query changes needed