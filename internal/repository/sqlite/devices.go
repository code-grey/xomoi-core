package sqlite

import (
	"context"
	"database/sql"
	"errors"

	"github.com/code-grey/xomoi-core/internal/core"
	"github.com/code-grey/xomoi-core/internal/repository/sqlite/dbgen"
)

type deviceRepo struct {
	db *DB
	q  *dbgen.Queries
}

func NewDeviceRepository(db *DB) *deviceRepo {
	return &deviceRepo{
		db: db,
		q:  dbgen.New(db.DB),
	}
}

func (r *deviceRepo) Create(ctx context.Context, device *core.Device) error {
	res, err := r.q.CreateDevice(ctx, dbgen.CreateDeviceParams{
		ID:         device.ID,
		Name:       device.Name,
		MacAddress: device.MACAddress,
		SecretKey:  device.SecretKey,
	})
	if err != nil {
		return err
	}
	device.CreatedAt = res.CreatedAt.Time
	return nil
}

func (r *deviceRepo) GetByMAC(ctx context.Context, macAddress string) (*core.Device, error) {
	row, err := r.q.GetDeviceByMAC(ctx, macAddress)
	if err != nil {
		return nil, err
	}
	return &core.Device{
		ID:         row.ID,
		Name:       row.Name,
		MACAddress: row.MacAddress,
		SecretKey:  row.SecretKey,
		LastSeen:   row.LastSeen.Time,
		CreatedAt:  row.CreatedAt.Time,
	}, nil
}

// ClaimDevice updates the device's name and rotates its secret key.
func (r *deviceRepo) ClaimDevice(ctx context.Context, macAddress, newName, newSecret string) error {
	res, err := r.db.ExecContext(ctx, "UPDATE devices SET name = ?, secret_key = ? WHERE mac_address = ?", newName, newSecret, macAddress)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("device not found")
	}
	return nil
}

func (r *deviceRepo) GetAll(ctx context.Context) ([]*core.Device, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, mac_address, secret_key, last_seen, created_at FROM devices")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*core.Device
	for rows.Next() {
		var dev core.Device
		var lastSeen, createdAt sql.NullTime
		if err := rows.Scan(&dev.ID, &dev.Name, &dev.MACAddress, &dev.SecretKey, &lastSeen, &createdAt); err != nil {
			return nil, err
		}
		dev.LastSeen = lastSeen.Time
		dev.CreatedAt = createdAt.Time
		devices = append(devices, &dev)
	}
	return devices, nil
}

func (r *deviceRepo) GetByID(ctx context.Context, id string) (*core.Device, error) { return nil, errors.New("unimplemented") }
func (r *deviceRepo) UpdateLastSeen(ctx context.Context, id string) error { return errors.New("unimplemented") }
func (r *deviceRepo) Delete(ctx context.Context, id string) error { return errors.New("unimplemented") }
