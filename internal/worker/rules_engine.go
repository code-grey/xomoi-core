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

package worker

import (
	"context"
	"log/slog"
	"sync"

	"github.com/code-grey/xomoi-core/internal/core"
	"github.com/code-grey/xomoi-core/internal/repository"
)

// RulesEngine evaluates telemetry against user-defined thresholds in real-time.
// It uses a zero-allocation map cache to prevent hitting SQLite on every MQTT message.
type RulesEngine struct {
	repo  repository.AlertRuleRepository
	mu    sync.RWMutex
	rules map[string][]core.AlertRule // map[deviceID][]rules
}

func NewRulesEngine(repo repository.AlertRuleRepository) *RulesEngine {
	return &RulesEngine{
		repo:  repo,
		rules: make(map[string][]core.AlertRule),
	}
}

// Start loads the rules into memory on boot.
func (e *RulesEngine) Start(ctx context.Context) error {
	return e.Reload(ctx)
}

// Reload re-fetches all rules from SQLite. Call this whenever an API modifies a rule.
func (e *RulesEngine) Reload(ctx context.Context) error {
	allRules, err := e.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	newMap := make(map[string][]core.AlertRule)
	activeCount := 0
	for _, r := range allRules {
		if r.IsActive {
			newMap[r.DeviceID] = append(newMap[r.DeviceID], *r)
			activeCount++
		}
	}

	e.mu.Lock()
	e.rules = newMap
	e.mu.Unlock()

	slog.Info("Rules Engine cache populated", "total_rules", len(allRules), "active", activeCount)
	return nil
}

// Evaluate checks the parsed telemetry against the cached rules for the device.
func (e *RulesEngine) Evaluate(deviceID string, temp, hum *float64, state string) {
	e.mu.RLock()
	deviceRules, exists := e.rules[deviceID]
	e.mu.RUnlock()

	if !exists {
		return // Fast path: No rules for this device
	}

	for _, rule := range deviceRules {
		var val float64
		var stringVal string
		isNumber := true

		switch rule.TagName {
		case "temp":
			if temp == nil { continue }
			val = *temp
		case "hum":
			if hum == nil { continue }
			val = *hum
		case "state":
			if state == "" { continue }
			stringVal = state
			isNumber = false
		default:
			continue
		}

		triggered := false
		if isNumber {
			switch rule.Condition {
			case ">":
				triggered = val > rule.Threshold
			case "<":
				triggered = val < rule.Threshold
			case ">=":
				triggered = val >= rule.Threshold
			case "<=":
				triggered = val <= rule.Threshold
			case "==":
				triggered = val == rule.Threshold
			case "!=":
				triggered = val != rule.Threshold
			}
		} else {
			// For boolean/state strings, Threshold 1.0 = "ON", 0.0 = "OFF"
			boolThreshold := "OFF"
			if rule.Threshold == 1.0 {
				boolThreshold = "ON"
			}
			if rule.Condition == "==" {
				triggered = stringVal == boolThreshold
			} else if rule.Condition == "!=" {
				triggered = stringVal != boolThreshold
			}
		}

		if triggered {
			slog.Warn("🚨 RULE TRIGGERED", 
				"device", deviceID, 
				"tag", rule.TagName, 
				"condition", rule.Condition, 
				"threshold", rule.Threshold,
				"actual_num", val,
				"actual_str", stringVal)
			// TODO: Forward to WebRTC Event Bus / Push Notifications
		}
	}
}
