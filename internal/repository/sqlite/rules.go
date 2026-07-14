package sqlite

import (
	"context"

	"github.com/code-grey/xomoi-core/internal/core"
)

type RuleRepository struct {
	db *DB
}

func NewRuleRepository(db *DB) *RuleRepository {
	return &RuleRepository{db: db}
}

func (r *RuleRepository) Create(ctx context.Context, rule *core.AlertRule) error {
	query := `
		INSERT INTO alert_rules (id, device_id, tag_name, condition, threshold, is_active)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query, rule.ID, rule.DeviceID, rule.TagName, rule.Condition, rule.Threshold, rule.IsActive)
	return err
}

func (r *RuleRepository) GetByDevice(ctx context.Context, deviceID string) ([]*core.AlertRule, error) {
	query := `SELECT id, device_id, tag_name, condition, threshold, is_active, created_at FROM alert_rules WHERE device_id = ?`
	rows, err := r.db.QueryContext(ctx, query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*core.AlertRule
	for rows.Next() {
		var rule core.AlertRule
		if err := rows.Scan(&rule.ID, &rule.DeviceID, &rule.TagName, &rule.Condition, &rule.Threshold, &rule.IsActive, &rule.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, &rule)
	}
	return rules, nil
}

func (r *RuleRepository) GetAll(ctx context.Context) ([]*core.AlertRule, error) {
	query := `SELECT id, device_id, tag_name, condition, threshold, is_active, created_at FROM alert_rules`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*core.AlertRule
	for rows.Next() {
		var rule core.AlertRule
		if err := rows.Scan(&rule.ID, &rule.DeviceID, &rule.TagName, &rule.Condition, &rule.Threshold, &rule.IsActive, &rule.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, &rule)
	}
	return rules, nil
}

func (r *RuleRepository) Update(ctx context.Context, rule *core.AlertRule) error {
	query := `
		UPDATE alert_rules 
		SET tag_name = ?, condition = ?, threshold = ?, is_active = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, rule.TagName, rule.Condition, rule.Threshold, rule.IsActive, rule.ID)
	return err
}

func (r *RuleRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM alert_rules WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
