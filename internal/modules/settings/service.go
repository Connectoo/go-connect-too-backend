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
	provider, err := s.loadProvider(ctx)
	if err != nil {
		return nil, err
	}
	return &SettingsResponse{General: general, Provider: redactProvider(provider)}, nil
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

	if req.Provider != nil {
		provider, err := s.mergeProvider(ctx, *req.Provider)
		if err != nil {
			return nil, err
		}
		providerRaw, err := encodeProvider(provider)
		if err != nil {
			return nil, fmt.Errorf("encode provider settings: %w", err)
		}
		if _, err := s.store.Upsert(ctx, ProviderKey, providerRaw, s.now()); err != nil {
			return nil, err
		}
	}

	provider, err := s.loadProvider(ctx)
	if err != nil {
		return nil, err
	}
	return &SettingsResponse{General: req.General, Provider: redactProvider(provider)}, nil
}

func (s *Service) loadProvider(ctx context.Context) (ProviderSettings, error) {
	setting, err := s.store.GetByKey(ctx, ProviderKey)
	if err != nil {
		if err == ErrNotFound {
			return defaultProvider(), nil
		}
		return ProviderSettings{}, err
	}
	return decodeProvider(setting.Value)
}

func (s *Service) mergeProvider(ctx context.Context, incoming ProviderSettings) (ProviderSettings, error) {
	current, err := s.loadProvider(ctx)
	if err != nil {
		return ProviderSettings{}, err
	}
	if incoming.RazorpayKeyID != "" {
		current.RazorpayKeyID = incoming.RazorpayKeyID
	}
	if incoming.RazorpayKeySecret != "" && incoming.RazorpayKeySecret != "********" {
		current.RazorpayKeySecret = incoming.RazorpayKeySecret
	}
	if incoming.RazorpayWebhookSecret != "" && incoming.RazorpayWebhookSecret != "********" {
		current.RazorpayWebhookSecret = incoming.RazorpayWebhookSecret
	}
	if incoming.SMTPHost != "" {
		current.SMTPHost = incoming.SMTPHost
	}
	if incoming.SMTPUser != "" {
		current.SMTPUser = incoming.SMTPUser
	}
	if incoming.SMTPPass != "" && incoming.SMTPPass != "********" {
		current.SMTPPass = incoming.SMTPPass
	}
	if incoming.SMTPFrom != "" {
		current.SMTPFrom = incoming.SMTPFrom
	}
	if incoming.S3Bucket != "" {
		current.S3Bucket = incoming.S3Bucket
	}
	if incoming.S3Region != "" {
		current.S3Region = incoming.S3Region
	}
	if incoming.S3AccessKey != "" && incoming.S3AccessKey != "********" {
		current.S3AccessKey = incoming.S3AccessKey
	}
	if incoming.S3SecretKey != "" && incoming.S3SecretKey != "********" {
		current.S3SecretKey = incoming.S3SecretKey
	}
	return current, nil
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
