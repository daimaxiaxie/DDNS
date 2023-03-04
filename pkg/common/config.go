package common

import "encoding/json"

type Config struct {
	Cloud    string          `json:"cloud"`
	Version  int             `json:"version"`
	Duration int             `json:"duration"`
	Extra    json.RawMessage `json:"extra"`
}
