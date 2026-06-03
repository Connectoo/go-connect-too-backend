package settings

import (
	"encoding/json"
	"time"
)

const GeneralKey = "general"

// Setting is a platform configuration row.
type Setting struct {
	Key       string
	Value     json.RawMessage
	UpdatedAt time.Time
}
