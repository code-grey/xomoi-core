CREATE TABLE IF NOT EXISTS telemetry_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    device_id TEXT NOT NULL,
    temperature REAL,
    humidity REAL,
    state TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (device_id) REFERENCES devices(mac_address) ON DELETE CASCADE
);

-- Crucial composite index for fast 30-day time-series querying
CREATE INDEX idx_telemetry_device_time ON telemetry_history(device_id, timestamp);
