package settings

import (
	"context"
	"errors"
	"testing"
	"time"
)

type mockSettingStore struct {
	setting *Setting
	err     error
}

func (m mockSettingStore) GetByKey(_ context.Context, _ string) (*Setting, error) {
	return m.setting, m.err
}

func (m mockSettingStore) Upsert(_ context.Context, _ string, value []byte, _ time.Time) (*Setting, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &Setting{Key: GeneralKey, Value: value}, nil
}

func TestService_UpdateSettings_requiresSiteName(t *testing.T) {
	svc := NewService(mockSettingStore{})

	_, err := svc.UpdateSettings(context.Background(), UpdateSettingsRequest{
		General: GeneralSettings{SupportEmail: "support@example.com"},
	})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("UpdateSettings() error = %v, want ErrValidation", err)
	}
}

func TestService_GetSettings_returnsDefaultsWhenMissing(t *testing.T) {
	svc := NewService(mockSettingStore{err: ErrNotFound})

	res, err := svc.GetSettings(context.Background())
	if err != nil {
		t.Fatalf("GetSettings() error = %v", err)
	}
	if res.General.SiteName == "" {
		t.Fatalf("GetSettings() = %+v, want default site name", res)
	}
}
