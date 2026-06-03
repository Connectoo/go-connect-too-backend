package settings

import "encoding/json"

// GeneralSettings holds editable platform configuration.
type GeneralSettings struct {
	SiteName        string `json:"site_name"`
	SupportEmail    string `json:"support_email"`
	MaintenanceMode bool   `json:"maintenance_mode"`
}

// SettingsResponse is the admin settings payload.
type SettingsResponse struct {
	General GeneralSettings `json:"general"`
}

// UpdateSettingsRequest replaces platform settings.
type UpdateSettingsRequest struct {
	General GeneralSettings `json:"general"`
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
