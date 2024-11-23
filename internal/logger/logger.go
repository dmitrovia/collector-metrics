package logger

import (
	"fmt"

	"go.uber.org/zap"
)

func Initialize(level string) (*zap.Logger, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("loggerInitializeSetLevel: %w", err)
	}

	cfg := zap.NewProductionConfig()

	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("loggerInitializeBuildCfg: %w", err)
	}

	return zl, nil
}
