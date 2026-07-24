// Xomoi-Core: Sovereign Edge Node
// Copyright (C) 2026 Adrish Bora (@code-grey) & Simanjit Hujuri (@code-zephyrus)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package sqlite

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/code-grey/xomoi-core/internal/core"
	"github.com/code-grey/xomoi-core/internal/repository"
	"github.com/code-grey/xomoi-core/internal/repository/sqlite/dbgen"
)

type sqliteTelemetryRepository struct {
	db      *sql.DB
	queries *dbgen.Queries
}

func NewTelemetryRepository(db *DB) *sqliteTelemetryRepository {
	return &sqliteTelemetryRepository{
		db:      db.DB,
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

func (r *sqliteTelemetryRepository) BulkInsert(ctx context.Context, records []repository.TelemetryRecord) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// SQLite has a maximum parameter limit of 32766. With 7 columns per row,
	// the absolute maximum is 4680 rows per INSERT statement.
	// We chunk in batches of 1000 to be safe and efficient.
	const chunkSize = 1000
	const cols = 7

	for i := 0; i < len(records); i += chunkSize {
		end := i + chunkSize
		if end > len(records) {
			end = len(records)
		}
		chunk := records[i:end]

		var sb strings.Builder
		sb.WriteString("INSERT INTO telemetry (id, device_id, timestamp, temperature, humidity, state, payload) VALUES ")
		
		vals := make([]interface{}, 0, len(chunk)*cols)

		for j, rec := range chunk {
			var nTemp, nHum sql.NullFloat64
			var nState sql.NullString

			if rec.Temperature != nil {
				nTemp = sql.NullFloat64{Float64: *rec.Temperature, Valid: true}
			}
			if rec.Humidity != nil {
				nHum = sql.NullFloat64{Float64: *rec.Humidity, Valid: true}
			}
			if rec.State != nil {
				nState = sql.NullString{String: *rec.State, Valid: true}
			}

			if j > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString("(?, ?, ?, ?, ?, ?, ?)")
			vals = append(vals, rec.ID, rec.DeviceID, rec.Timestamp, nTemp, nHum, nState, rec.PayloadBlob)
		}

		// A single, massive CGO call containing up to 1000 rows
		_, err = tx.ExecContext(ctx, sb.String(), vals...)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
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
