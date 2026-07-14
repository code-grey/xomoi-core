-- name: CreateDevice :one
INSERT INTO devices (
    id, name, mac_address, secret_key
) VALUES (
    ?, ?, ?, ?
)
RETURNING *;

-- name: GetDeviceByMAC :one
SELECT * FROM devices
WHERE mac_address = ? LIMIT 1;

-- name: InsertTelemetryRollup :exec
INSERT INTO telemetry_rollups (
    device_id, date, avg_payload, max_payload, min_payload
) VALUES (
    ?, ?, ?, ?, ?
)
ON CONFLICT(device_id, date) DO UPDATE SET
    avg_payload = excluded.avg_payload,
    max_payload = excluded.max_payload,
    min_payload = excluded.min_payload;

-- name: DeleteRawTelemetryBefore :exec
DELETE FROM telemetry
WHERE date(timestamp) < ?;

-- name: UpsertUserDeviceRole :exec
INSERT INTO user_device_roles (
    user_id, device_id, role
) VALUES (
    ?, ?, ?
)
ON CONFLICT(user_id, device_id) DO UPDATE SET
    role = excluded.role;

-- name: GetDeviceRoles :many
SELECT u.username, r.role
FROM user_device_roles r
JOIN users u ON u.id = r.user_id
WHERE r.device_id = ?;
