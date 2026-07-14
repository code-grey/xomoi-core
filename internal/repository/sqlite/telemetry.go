package sqlite

import (
	"context"
	"database/sql"
	"time"

	"github.com/code-grey/xomoi-core/internal/core"
	"github.com/code-grey/xomoi-core/internal/repository/sqlite/dbgen"
)

type sqliteTelemetryRepository struct {
	queries *dbgen.Queries
}

func NewTelemetryRepository(db *DB) *sqliteTelemetryRepository {
	return &sqliteTelemetryRepository{
		queries: dbgen.New(db.DB),
	}
}

func (r *sqliteTelemetryRepository) InsertTelemetry(ctx context.Context, deviceID string, temp, hum *float64, state string) error {
	var nTemp, nHum sql.NullFloat64
	var nState sql.NullString

	if temp != nil {
		nTemp = sql.NullFloat64{Float64: *temp, Valid: true}
	}
	if hum != nil {
		nHum = sql.NullFloat64{Float64: *hum, Valid: true}
	}
	if state != "" {
		nState = sql.NullString{String: state, Valid: true}
	}

	return r.queries.InsertTelemetry(ctx, dbgen.InsertTelemetryParams{
		DeviceID:    deviceID,
		Temperature: nTemp,
		Humidity:    nHum,
		State:       nState,
	})
}

func (r *sqliteTelemetryRepository) GetDeviceHistory(ctx context.Context, deviceID string, since time.Time) ([]core.TelemetryPoint, error) {
	history, err := r.queries.GetTelemetryHistory(ctx, dbgen.GetTelemetryHistoryParams{
		DeviceID:  deviceID,
		Timestamp: sql.NullTime{Time: since, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	
	points := make([]core.TelemetryPoint, 0, len(history))
	for _, row := range history {
		p := core.TelemetryPoint{
			Timestamp: row.Timestamp.Time,
		}
		if row.Temperature.Valid {
			v := row.Temperature.Float64
			p.Temp = &v
		}
		if row.Humidity.Valid {
			v := row.Humidity.Float64
			p.Hum = &v
		}
		if row.State.Valid {
			p.State = row.State.String
		}
		points = append(points, p)
	}
	
	return points, nil
}
