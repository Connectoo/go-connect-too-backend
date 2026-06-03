package settings

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SettingStore loads and saves settings rows.
type SettingStore interface {
	GetByKey(ctx context.Context, key string) (*Setting, error)
	Upsert(ctx context.Context, key string, value []byte, at time.Time) (*Setting, error)
}

// Service handles platform settings business logic.
type Service struct {
	store SettingStore
	now   func() time.Time
}

// NewService creates a settings service.
func NewService(store SettingStore) *Service {
	return &Service{
		store: store,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

// GetSettings returns current platform settings.
func (s *Service) GetSettings(ctx context.Context) (*SettingsResponse, error) {
	general, err := s.loadGeneral(ctx)
	if err != nil {
		return nil, err
	}
	return &SettingsResponse{General: general}, nil
}

// UpdateSettings replaces platform settings.
func (s *Service) UpdateSettings(ctx context.Context, req UpdateSettingsRequest) (*SettingsResponse, error) {
	if err := validateGeneral(req.General); err != nil {
		return nil, err
	}

	raw, err := encodeGeneral(req.General)
	if err != nil {
		return nil, fmt.Errorf("encode settings: %w", err)
	}

	if _, err := s.store.Upsert(ctx, GeneralKey, raw, s.now()); err != nil {
		return nil, err
	}

	return &SettingsResponse{General: req.General}, nil
}

func (s *Service) loadGeneral(ctx context.Context) (GeneralSettings, error) {
	setting, err := s.store.GetByKey(ctx, GeneralKey)
	if err != nil {
		if err == ErrNotFound {
			return defaultGeneral(), nil
		}
		return GeneralSettings{}, err
	}
	return decodeGeneral(setting.Value)
}

func validateGeneral(settings GeneralSettings) error {
	if strings.TrimSpace(settings.SiteName) == "" {
		return fmt.Errorf("%w: site_name is required", ErrValidation)
	}
	if strings.TrimSpace(settings.SupportEmail) == "" {
		return fmt.Errorf("%w: support_email is required", ErrValidation)
	}
	return nil
}
