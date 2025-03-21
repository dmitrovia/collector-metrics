// Package dbrepository provides
// working with postgre storage
package dbrepository

import (
	"context"
	"fmt"
	"sync"

	"github.com/dmitrovia/collector-metrics/internal/models/apimodels"
	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBepository - describing the storage.
type DBepository struct {
	conn        *pgxpool.Pool
	mutexG      *sync.Mutex
	mutexC      *sync.Mutex
	databaseDSN string
}

// Initiate - initialization of initial parameters.
func (m *DBepository) Initiate(
	dsn string,
	conn *pgxpool.Pool,
) {
	m.databaseDSN = dsn
	m.conn = conn
	m.mutexG = &sync.Mutex{}
	m.mutexC = &sync.Mutex{}
}

// Init - initialization of initial parameters.
func (m *DBepository) Init() {
}

// AddMetrics - adds metrics to the database.
func (m *DBepository) AddMetrics(
	ctx *context.Context,
	gauges map[string]bizmodels.Gauge,
	counters map[string]bizmodels.Counter,
) error {
	for _, gauge := range gauges {
		err := m.AddGauge(ctx, &gauge)
		if err != nil {
			return fmt.Errorf("AddMetrics->m.AddGauge: %w", err)
		}
	}

	for _, counter := range counters {
		_, err := m.AddCounter(ctx, &counter, false)
		if err != nil {
			return fmt.Errorf("AddMetrics->m.AddCounter: %w", err)
		}
	}

	return nil
}

// GetAllMetricsAPI - get all metrics in API format.
func (m *DBepository) GetAllMetricsAPI(
	ctx *context.Context,
) (*apimodels.ArrMetrics, error) {
	arr1, err := m.GetAllGaugesAPI(ctx)
	if err != nil {
		return nil, fmt.Errorf("GAllMetricsAPI->m.GAPI: %w", err)
	}

	arr2, err := m.GetAllCountersAPI(ctx)
	if err != nil {
		return nil, fmt.Errorf("GAllMetricsAPI->m.CAPI: %w", err)
	}

	result := make(apimodels.ArrMetrics, 0)
	result = append(result, arr1...)
	result = append(result, arr2...)

	return &result, nil
}

// GetAllGaugesAPI - get all gauge metrics in API format.
func (m *DBepository) GetAllGaugesAPI(
	ctx *context.Context) (
	apimodels.ArrMetrics,
	error,
) {
	var (
		name  string
		value float64
	)

	gauges := make(apimodels.ArrMetrics, 0)

	rows, err := m.conn.Query(
		*ctx,
		"select name, value from gauges")
	if err != nil {
		return nil, fmt.Errorf("GetAllGAPI->m.conn.Q: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&name, &value)
		if err != nil {
			fmt.Printf("Scan error: %v", err)
		} else {
			temp := &apimodels.Metrics{}
			temp.ID = name
			temp.Value = &value
			temp.MType = bizmodels.GaugeName

			gauges = append(gauges, *temp)
		}
	}

	return gauges, nil
}

// GetAllCountersAPI - get all
// counter metrics in API format.
func (m *DBepository) GetAllCountersAPI(
	ctx *context.Context) (
	apimodels.ArrMetrics,
	error,
) {
	var (
		name  string
		value int64
	)

	counters := make(apimodels.ArrMetrics, 0)

	rows, err := m.conn.Query(
		*ctx,
		"select name, value from counters")
	if err != nil {
		return nil, fmt.Errorf("GetAllCAPI->m.conn.Q: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&name, &value)
		if err != nil {
			fmt.Printf("Scan error: %v", err)
		} else {
			temp := &apimodels.Metrics{}
			temp.ID = name
			temp.Delta = &value
			temp.MType = bizmodels.CounterName

			counters = append(counters, *temp)
		}
	}

	return counters, nil
}

// GetAllGauges - get all gauges metrics from database.
func (m *DBepository) GetAllGauges(ctx *context.Context) (
	map[string]bizmodels.Gauge,
	error,
) {
	var (
		gauges map[string]bizmodels.Gauge
		name   string
		value  float64
	)

	gauges = make(map[string]bizmodels.Gauge)

	rows, err := m.conn.Query(
		*ctx,
		"select name, value from gauges")
	if err != nil {
		return nil, fmt.Errorf("GetAllGauges->m.conn.Q: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&name, &value)
		if err != nil {
			fmt.Printf("Scan error: %v", err)
		} else {
			temp := &bizmodels.Gauge{}
			temp.Name = name
			temp.Value = value

			gauges[name] = *temp
		}
	}

	return gauges, nil
}

// GetAllCounters - get all counter metrics from database.
func (m *DBepository) GetAllCounters(ctx *context.Context) (
	map[string]bizmodels.Counter,
	error,
) {
	var (
		counters map[string]bizmodels.Counter
		name     string
		value    int64
	)

	counters = make(map[string]bizmodels.Counter)

	rows, err := m.conn.Query(
		*ctx,
		"select name, value from counters")
	if err != nil {
		return nil, fmt.Errorf("GetAllCounters->m.CQ: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&name, &value)
		if err != nil {
			fmt.Printf("Scan error: %v", err)
		} else {
			temp := &bizmodels.Counter{}
			temp.Name = name
			temp.Value = value

			counters[name] = *temp
		}
	}

	return counters, nil
}

// GetGaugeMetric - get gauge metric by name from database.
func (m *DBepository) GetGaugeMetric(
	ctx *context.Context,
	name string,
) (*bizmodels.Gauge, error) {
	var temp *bizmodels.Gauge

	var nameMetric string

	var value float64

	temp = &bizmodels.Gauge{}

	err := m.conn.QueryRow(
		*ctx,
		"select name, value from gauges where name=$1",
		name).Scan(&nameMetric, &value)
	if err != nil {
		return nil, fmt.Errorf("GetGaugeMetric->QR: %w", err)
	}

	temp.Name = nameMetric
	temp.Value = value

	return temp, nil
}

// GetCounterMetric - get counter
// metric by name from database.
func (m *DBepository) GetCounterMetric(
	ctx *context.Context,
	name string,
) (*bizmodels.Counter, error) {
	var temp *bizmodels.Counter

	var nameMetric string

	var value int64

	temp = &bizmodels.Counter{}

	err := m.conn.QueryRow(
		*ctx,
		"select name, value from counters where name=$1",
		name).Scan(&nameMetric, &value)
	if err != nil {
		return nil,
			fmt.Errorf("GetGaugeMetric->m.conn.QueryRow: %w", err)
	}

	temp.Name = nameMetric
	temp.Value = value

	return temp, nil
}

// AddGauge - add the gauge metric to the database.
func (m *DBepository) AddGauge(
	ctx *context.Context,
	gauge *bizmodels.Gauge,
) error {
	m.mutexG.Lock()
	defer m.mutexG.Unlock()

	// comment - a transaction is needed here
	rows, err := m.conn.Exec(
		*ctx,
		"UPDATE gauges SET value = $1 where name=$2",
		gauge.Value,
		gauge.Name)
	if err != nil {
		return fmt.Errorf("AddGauge->m.conn.Exec: %w", err)
	}

	if rows.RowsAffected() == 0 {
		_, err := m.conn.Exec(
			*ctx,
			"INSERT INTO gauges (name, value) VALUES ($1, $2)",
			gauge.Name,
			gauge.Value)
		if err != nil {
			return fmt.Errorf("AddGauge->INSERT INTO error: %w", err)
		}
	}

	return nil
}

// AddCounter - add the counter metric to the database.
func (m *DBepository) AddCounter(
	ctx *context.Context,
	counter *bizmodels.Counter,
	isNew bool,
) (*bizmodels.Counter, error) {
	m.mutexC.Lock()
	defer m.mutexC.Unlock()

	var tmp string

	if isNew {
		tmp = "UPDATE counters SET value = $1 where name=$2"
	} else {
		tmp = "UPDATE counters SET value = value" +
			"+ $1 where name=$2"
	}

	rows, err := m.conn.Exec(*ctx,
		tmp,
		counter.Value,
		counter.Name)
	if err != nil {
		return nil, fmt.Errorf("AddCounter->UPD c SET: %w", err)
	}

	if rows.RowsAffected() == 0 {
		_, err = m.conn.Exec(
			*ctx,
			"INSERT INTO counters (name, value) VALUES ($1, $2)",
			counter.Name,
			counter.Value)
		if err != nil {
			return nil, fmt.Errorf("AddCounter->II: %w", err)
		}

		return counter, nil
	}

	temp, err := m.GetCounterMetric(ctx, counter.Name)
	if err != nil {
		return nil, fmt.Errorf("AddCounter->m.GetCM %w", err)
	}

	return temp, nil
}
