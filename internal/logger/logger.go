// Package logger provides functions
// working with server logging.
package logger

import (
	"fmt"

	"go.uber.org/zap"
)

// Initialize - initializing a logging object.
func Initialize(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("Initialize->ParseAtomicLevel: %w",
			err)
	}

	cfg := zap.NewProductionConfig()

	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("Initialize->Build: %w", err)
	}

	return zl, nil
}
