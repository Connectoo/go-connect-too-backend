package settings

import (
	"encoding/json"
	"time"
)

const (
	GeneralKey  = "general"
	ProviderKey = "providers"
)

// Setting is a platform configuration row.
type Setting struct {
	Key       string
	Value     json.RawMessage
	UpdatedAt time.Time
}
