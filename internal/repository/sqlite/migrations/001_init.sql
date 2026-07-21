CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS devices (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    mac_address TEXT UNIQUE NOT NULL,
    secret_key TEXT NOT NULL,
    last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sensor_tags (
    device_id TEXT NOT NULL,
    field_id INTEGER NOT NULL,
    tag_name TEXT NOT NULL,
    data_type TEXT NOT NULL,
    PRIMARY KEY(device_id, field_id),
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS alert_rules (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    tag_name TEXT NOT NULL,
    condition TEXT NOT NULL,
    threshold REAL NOT NULL,
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS telemetry (
    id TEXT PRIMARY KEY,
    device_id TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    temperature REAL,
    humidity REAL,
    state TEXT,
    payload BLOB NOT NULL,
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);

-- 1. Table for aggregated telemetry rollups (Saving disk space)
CREATE TABLE IF NOT EXISTS telemetry_rollups (
    device_id TEXT NOT NULL,
    date DATE NOT NULL,
    avg_payload JSON NOT NULL,
    max_payload JSON NOT NULL,
    min_payload JSON NOT NULL,
    PRIMARY KEY(device_id, date),
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);

-- 2. Table for granular Role-Based Access Control (RBAC) sharing
CREATE TABLE IF NOT EXISTS user_device_roles (
    user_id TEXT NOT NULL,
    device_id TEXT NOT NULL,
    role TEXT NOT NULL, -- 'owner', 'editor', 'viewer'
    PRIMARY KEY(user_id, device_id),
    FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(device_id) REFERENCES devices(id) ON DELETE CASCADE
);

-- Crucial composite index for fast 30-day time-series querying
CREATE INDEX IF NOT EXISTS idx_telemetry_device_time ON telemetry(device_id, timestamp);
