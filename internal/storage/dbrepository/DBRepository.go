package dbrepository

import (
	"context"
	"fmt"

	"github.com/dmitrovia/collector-metrics/internal/models/bizmodels"
	"github.com/jackc/pgx/v5"
)

type DBepository struct {
	databaseDSN string
	conn        *pgx.Conn
}

func (m *DBepository) Initiate(
	dsn string,
	conn *pgx.Conn,
) {
	m.databaseDSN = dsn
	m.conn = conn
}

func (m *DBepository) Init() {
}

func (m *DBepository) AddMetrics(
	ctx *context.Context,
	gauges map[string]bizmodels.Gauge,
	counters map[string]bizmodels.Counter,
) error {
	tranz, err := m.conn.Begin(*ctx)
	if err != nil {
		return fmt.Errorf("AddMetrics->m.conn.Begin %w", err)
	}

	for _, gauge := range gauges {
		err = m.AddGauge(ctx, &gauge)
		if err != nil {
			return fmt.Errorf("AddMetrics->m.AddGauge %w", err)
		}
	}

	for _, counter := range counters {
		_, err = m.AddCounter(ctx, &counter)
		if err != nil {
			return fmt.Errorf("AddMetrics->m.AddCounter %w", err)
		}
	}

	err = tranz.Commit(*ctx)
	if err != nil {
		return fmt.Errorf("AddMetrics->tranz.Commit %w", err)
	}

	return nil
}

func (m *DBepository) GetAllGauges(ctx *context.Context) (
	*map[string]bizmodels.Gauge,
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
		return nil, fmt.Errorf("GetAllGauges->m.conn.Query %w",
			err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&name, &value)
		if err != nil {
			fmt.Printf("Scan error: %v", err)
		} else {
			temp := new(bizmodels.Gauge)
			temp.Name = name
			temp.Value = value

			gauges[name] = *temp
		}
	}

	return &gauges, nil
}

func (m *DBepository) GetAllCounters(ctx *context.Context) (
	*map[string]bizmodels.Counter,
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
		return nil, fmt.Errorf("GetAllCounters->m.conn.Query %w",
			err)
	}

	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&name, &value)
		if err != nil {
			fmt.Printf("Scan error: %v", err)
		} else {
			temp := new(bizmodels.Counter)
			temp.Name = name
			temp.Value = value

			counters[name] = *temp
		}
	}

	return &counters, nil
}

func (m *DBepository) GetGaugeMetric(
	ctx *context.Context,
	name string,
) (*bizmodels.Gauge, error) {
	var temp *bizmodels.Gauge

	var nameMetric string

	var value float64

	temp = new(bizmodels.Gauge)

	err := m.conn.QueryRow(
		*ctx,
		"select name, value from gauges where name=$1",
		name).Scan(&nameMetric, &value)
	if err != nil {
		return nil, fmt.Errorf("GetGaugeMetric->QueryRow %w",
			err)
	}

	temp.Name = nameMetric
	temp.Value = value

	return temp, nil
}

func (m *DBepository) GetCounterMetric(
	ctx *context.Context,
	name string,
) (*bizmodels.Counter, error) {
	var temp *bizmodels.Counter

	var nameMetric string

	var value int64

	temp = new(bizmodels.Counter)

	err := m.conn.QueryRow(
		*ctx,
		"select name, value from counters where name=$1",
		name).Scan(&nameMetric, &value)
	if err != nil {
		return nil,
			fmt.Errorf("GetGaugeMetric->m.conn.QueryRow %w",
				err)
	}

	temp.Name = nameMetric
	temp.Value = value

	return temp, nil
}

func (m *DBepository) AddGauge(
	ctx *context.Context,
	gauge *bizmodels.Gauge,
) error {
	rows, err := m.conn.Exec(
		*ctx,
		"UPDATE gauges SET value = $1 where name=$2",
		gauge.Value,
		gauge.Name)
	if err != nil {
		return fmt.Errorf("AddGauge->m.conn.Exec( %w", err)
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

func (m *DBepository) AddCounter(
	ctx *context.Context,
	counter *bizmodels.Counter,
) (*bizmodels.Counter, error) {
	rows, err := m.conn.Exec(*ctx,
		"UPDATE counters SET value = value + $1 where name=$2",
		counter.Value,
		counter.Name)
	if err != nil {
		return nil,
			fmt.Errorf("AddCounter->UPDATE counters SET: %w",
				err)
	}

	if rows.RowsAffected() == 0 {
		_, err := m.conn.Exec(
			*ctx,
			"INSERT INTO counters (name, value) VALUES ($1, $2)",
			counter.Name,
			counter.Value)
		if err != nil {
			return nil,
				fmt.Errorf("AddCounter->INSERT INTO error: %w",
					err)
		}

		return counter, nil
	}

	temp, err := m.GetCounterMetric(ctx, counter.Name)
	if err != nil {
		return nil,
			fmt.Errorf("AddCounter->m.GetCounterMetric %w",
				err)
	}

	return temp, nil
}
