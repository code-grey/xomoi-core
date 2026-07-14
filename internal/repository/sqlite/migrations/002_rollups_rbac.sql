-- Migration 002: Telemetry Rollups and RBAC

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
