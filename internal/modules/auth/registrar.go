package auth

import (
	"context"
	"database/sql"

	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/customers"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/employees"
	"github.com/MustafaKheda/go-connect-too-backend/internal/modules/users"
	"github.com/MustafaKheda/go-connect-too-backend/internal/platform/database"
)

// AccountRegistrar creates a user and role-specific profile atomically.
type AccountRegistrar interface {
	RegisterCustomer(ctx context.Context, user *users.User) error
	RegisterEmployee(ctx context.Context, user *users.User) error
}

type registrar struct {
	db        *sql.DB
	users     *users.Repository
	customers *customers.Repository
	employees *employees.Repository
}

// NewRegistrar wires transactional account registration.
func NewRegistrar(db *sql.DB, userRepo *users.Repository, customerRepo *customers.Repository, employeeRepo *employees.Repository) AccountRegistrar {
	return &registrar{
		db:        db,
		users:     userRepo,
		customers: customerRepo,
		employees: employeeRepo,
	}
}

func (r *registrar) RegisterCustomer(ctx context.Context, user *users.User) error {
	return database.RunInTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := r.users.CreateInTx(ctx, tx, user); err != nil {
			return err
		}
		return r.customers.CreateForUserInTx(ctx, tx, user.ID, user.CreatedAt)
	})
}

func (r *registrar) RegisterEmployee(ctx context.Context, user *users.User) error {
	return database.RunInTx(ctx, r.db, func(tx *sql.Tx) error {
		if err := r.users.CreateInTx(ctx, tx, user); err != nil {
			return err
		}
		return r.employees.CreateForUserInTx(ctx, tx, user.ID, user.CreatedAt)
	})
}
