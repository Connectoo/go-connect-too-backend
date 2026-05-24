package employees

import "database/sql"

// Repository will persist employee profiles in a later phase.
// Employee registration currently creates a users row with role=employee via auth.
type Repository struct {
	db *sql.DB
}

// NewRepository creates an employee repository.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}
