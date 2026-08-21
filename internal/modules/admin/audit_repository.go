package admin

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MustafaKheda/go-connect-too-backend/internal/shared/middleware"
)

// AuditRepository persists admin audit logs.
type AuditRepository struct {
	db *sql.DB
}

// NewAuditRepository creates an admin audit repository.
func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// InsertAuditLog stores an admin audit entry.
func (r *AuditRepository) InsertAuditLog(ctx context.Context, entry middleware.AdminAuditEntry) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO admin_audit_logs (id, admin_user_id, action, resource_type, resource_id, details, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.ID, entry.AdminUserID, entry.Action, entry.ResourceType, entry.ResourceID, entry.Details, entry.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert admin audit log: %w", err)
	}
	return nil
}
