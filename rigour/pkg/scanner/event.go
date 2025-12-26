package scanner

import (
	"encoding/json"
	"time"
)

type ScanEvent struct {
	Timestamp time.Time       `json:"timestamp"`
	IP        string          `json:"ip"`
	Port      int             `json:"port"`
	Protocol  string          `json:"protocol"`
	TLS       bool            `json:"tls"`
	Transport string          `json:"transport"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
}
