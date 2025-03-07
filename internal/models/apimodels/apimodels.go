// Package apimodels
// describes the data exchange model for handlers
package apimodels

type Metrics struct {
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	ID    string   `json:"id"`
	MType string   `json:"type"`
}

type ArrMetrics []Metrics

type CfgServer struct {
	PORT                 string `json:"address"`
	FileStoragePath      string `json:"storeFile"`
	DatabaseDSN          string `json:"databaseDsn"`
	Key                  string `json:"keySha"`
	CryptoPrivateKeyPath string `json:"cryptoKey"`
	StoreInterval        int    `json:"storeInterval"`
	Restore              bool   `json:"restore"`
}

type CfgAgent struct {
	PORT                string `json:"address"`
	Key                 string `json:"keySha"`
	CryptoPublicKeyPath string `json:"cryptoKey"`
	ReportInterval      int    `json:"reportInterval"`
	PollInterval        int    `json:"pollInterval"`
}
