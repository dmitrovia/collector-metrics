// Package apimodels
// describes the data exchange model for handlers
package apimodels

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}

type ArrMetrics []Metrics
