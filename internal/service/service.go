package service

import (
	"fmt"

	"github.com/dmitrovia/collector-metrics/internal/models"
	"github.com/dmitrovia/collector-metrics/internal/storage"
)

type Service interface {
	GetMapStringsAllMetrics() *map[string]string
	AddGauge(mname string, mvalue float64)
	AddCounter(mname string, mvalue int64)
	GetStringValueGaugeMetric(mname string) (string, error)
	GetStringValueCounterMetric(mname string) (string, error)
}

type MemoryService struct {
	repository storage.Repository
}

func (s *MemoryService) GetMapStringsAllMetrics() *map[string]string {
	return s.repository.GetMapStringsAllMetrics()
}

func (s *MemoryService) AddGauge(mname string, mvalue float64) {
	s.repository.AddGauge(&models.Gauge{Name: mname, Value: mvalue})
}

func (s *MemoryService) AddCounter(mname string, mvalue int64) {
	s.repository.AddCounter(&models.Counter{Name: mname, Value: mvalue})
}

func (s *MemoryService) GetStringValueGaugeMetric(mname string) (string, error) {
	val, err := s.repository.GetStringValueGaugeMetric(mname)
	if err != nil {
		return val, fmt.Errorf("addrIsValid: %w", err)
	}

	return val, nil
}

func (s *MemoryService) GetStringValueCounterMetric(mname string) (string, error) {
	val, err := s.repository.GetStringValueCounterMetric(mname)
	if err != nil {
		return val, fmt.Errorf("addrIsValid: %w", err)
	}

	return val, nil
}

func NewMemoryService(repository storage.Repository) *MemoryService {
	return &MemoryService{repository: repository}
}
