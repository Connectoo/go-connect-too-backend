package settings

import "encoding/json"

// GeneralSettings holds editable platform configuration.
type GeneralSettings struct {
	SiteName        string `json:"site_name"`
	SupportEmail    string `json:"support_email"`
	MaintenanceMode bool   `json:"maintenance_mode"`
}

// ProviderSettings holds third-party provider configuration.
type ProviderSettings struct {
	RazorpayKeyID         string `json:"razorpay_key_id,omitempty"`
	RazorpayKeySecret     string `json:"razorpay_key_secret,omitempty"`
	RazorpayWebhookSecret string `json:"razorpay_webhook_secret,omitempty"`
	SMTPHost              string `json:"smtp_host,omitempty"`
	SMTPUser              string `json:"smtp_user,omitempty"`
	SMTPPass              string `json:"smtp_pass,omitempty"`
	SMTPFrom              string `json:"smtp_from,omitempty"`
	S3Bucket              string `json:"s3_bucket,omitempty"`
	S3Region              string `json:"s3_region,omitempty"`
	S3AccessKey           string `json:"s3_access_key,omitempty"`
	S3SecretKey           string `json:"s3_secret_key,omitempty"`
}

// SettingsResponse is the admin settings payload.
type SettingsResponse struct {
	General  GeneralSettings  `json:"general"`
	Provider ProviderSettings `json:"provider,omitempty"`
}

// UpdateSettingsRequest replaces platform settings.
type UpdateSettingsRequest struct {
	General  GeneralSettings   `json:"general"`
	Provider *ProviderSettings `json:"provider,omitempty"`
}

// defaultGeneral returns seeded defaults when no row exists yet.
func defaultGeneral() GeneralSettings {
	return GeneralSettings{
		SiteName:        "Go Connect",
		SupportEmail:    "support@example.com",
		MaintenanceMode: false,
	}
}

func encodeGeneral(settings GeneralSettings) (json.RawMessage, error) {
	return json.Marshal(settings)
}

func defaultProvider() ProviderSettings {
	return ProviderSettings{}
}

func encodeProvider(settings ProviderSettings) (json.RawMessage, error) {
	return json.Marshal(settings)
}

func decodeProvider(raw json.RawMessage) (ProviderSettings, error) {
	if len(raw) == 0 {
		return defaultProvider(), nil
	}
	var settings ProviderSettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return ProviderSettings{}, err
	}
	return settings, nil
}

func redactProvider(settings ProviderSettings) ProviderSettings {
	if settings.RazorpayKeySecret != "" {
		settings.RazorpayKeySecret = "********"
	}
	if settings.RazorpayWebhookSecret != "" {
		settings.RazorpayWebhookSecret = "********"
	}
	if settings.SMTPPass != "" {
		settings.SMTPPass = "********"
	}
	if settings.S3SecretKey != "" {
		settings.S3SecretKey = "********"
	}
	if settings.S3AccessKey != "" {
		settings.S3AccessKey = "********"
	}
	return settings
}

func decodeGeneral(raw json.RawMessage) (GeneralSettings, error) {
	if len(raw) == 0 {
		return defaultGeneral(), nil
	}
	var settings GeneralSettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return GeneralSettings{}, err
	}
	return settings, nil
}
