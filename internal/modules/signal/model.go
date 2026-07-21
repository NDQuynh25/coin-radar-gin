package signal

import "time"

type Signal struct {
	Time      time.Time              `json:"time"`
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // funding_spike, oi_delta, volume_spike...
	Exchange  string                 `json:"exchange,omitempty"`
	Symbol    string                 `json:"symbol"`
	Severity  string                 `json:"severity"` // info, warning, critical
	Value     float64                `json:"value"`
	Threshold float64                `json:"threshold"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
}
