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

package alerts

import (
	"log"

	"github.com/code-grey/xomoi-core/internal/core"
	"github.com/code-grey/xomoi-core/internal/state"
)

// Evaluator checks the hot state against user-defined threshold rules.
type Evaluator struct {
	rules []core.AlertRule
}

// NewEvaluator creates a new rule evaluator engine.
func NewEvaluator(rules []core.AlertRule) *Evaluator {
	return &Evaluator{rules: rules}
}

// Evaluate runs the rules against a specific device's latest telemetry state.
// This is called instantly after the HotState is updated.
func (e *Evaluator) Evaluate(deviceState state.DeviceState) {
	log.Printf("Evaluating %d rules against device %s", len(e.rules), deviceState.DeviceID)
	
	for _, rule := range e.rules {
		if rule.DeviceID == deviceState.DeviceID && rule.IsActive {
			// Skeleton logic: 
			// 1. Extract specific field (e.g., 'temperature') from deviceState.Payload (JSON)
			// 2. Compare against rule.Condition (>, <, ==) and rule.Threshold
			// 3. If condition met -> Trigger Notification (e.g., Push to Flutter UI / Webhook)
		}
	}
}
