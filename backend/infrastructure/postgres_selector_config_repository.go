package infrastructure

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ileego/go_browser_capture/backend/domain"
)

type PostgresSelectorConfigRepository struct {
	db *pgxpool.Pool
}

func NewPostgresSelectorConfigRepository(db *pgxpool.Pool) *PostgresSelectorConfigRepository {
	return &PostgresSelectorConfigRepository{db: db}
}

func (r *PostgresSelectorConfigRepository) Save(ctx context.Context, config *domain.SelectorConfig) error {
	if config.ID == "" {
		return r.insert(ctx, config)
	}
	return r.update(ctx, config)
}

func (r *PostgresSelectorConfigRepository) insert(ctx context.Context, config *domain.SelectorConfig) error {
	query := `
		INSERT INTO selector_configs (name, domain, selector, regex, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id string
	err := r.db.QueryRow(ctx, query, config.Name, config.Domain, config.Selector, config.Regex, config.IsActive).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert selector config: %w", err)
	}

	config.ID = domain.SelectorConfigID(id)
	return nil
}

func (r *PostgresSelectorConfigRepository) update(ctx context.Context, config *domain.SelectorConfig) error {
	query := `
		UPDATE selector_configs
		SET name = $1, domain = $2, selector = $3, regex = $4, is_active = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(ctx, query, config.Name, config.Domain, config.Selector, config.Regex, config.IsActive, config.ID)
	if err != nil {
		return fmt.Errorf("failed to update selector config: %w", err)
	}

	return nil
}

func (r *PostgresSelectorConfigRepository) FindByID(ctx context.Context, id domain.SelectorConfigID) (*domain.SelectorConfig, error) {
	query := `
		SELECT id, name, domain, selector, regex, is_active
		FROM selector_configs
		WHERE id = $1
	`

	config, err := r.querySingle(ctx, query, id)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (r *PostgresSelectorConfigRepository) FindByName(ctx context.Context, name string) (*domain.SelectorConfig, error) {
	query := `
		SELECT id, name, domain, selector, regex, is_active
		FROM selector_configs
		WHERE name = $1
	`

	config, err := r.querySingle(ctx, query, name)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (r *PostgresSelectorConfigRepository) FindByDomain(ctx context.Context, domain string) ([]*domain.SelectorConfig, error) {
	query := `
		SELECT id, name, domain, selector, regex, is_active
		FROM selector_configs
		WHERE domain = $1
		ORDER BY name
	`

	configs, err := r.queryAll(ctx, query, domain)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func (r *PostgresSelectorConfigRepository) FindAll(ctx context.Context) ([]*domain.SelectorConfig, error) {
	query := `
		SELECT id, name, domain, selector, regex, is_active
		FROM selector_configs
		ORDER BY name
	`

	configs, err := r.queryAll(ctx, query)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func (r *PostgresSelectorConfigRepository) FindActive(ctx context.Context) ([]*domain.SelectorConfig, error) {
	query := `
		SELECT id, name, domain, selector, regex, is_active
		FROM selector_configs
		WHERE is_active = true
		ORDER BY name
	`

	configs, err := r.queryAll(ctx, query)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func (r *PostgresSelectorConfigRepository) Delete(ctx context.Context, id domain.SelectorConfigID) error {
	query := `
		DELETE FROM selector_configs
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete selector config: %w", err)
	}

	return nil
}

func (r *PostgresSelectorConfigRepository) querySingle(ctx context.Context, query string, args ...interface{}) (*domain.SelectorConfig, error) {
	row := r.db.QueryRow(ctx, query, args...)

	var config domain.SelectorConfig
	err := row.Scan(&config.ID, &config.Name, &config.Domain, &config.Selector, &config.Regex, &config.IsActive)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func (r *PostgresSelectorConfigRepository) queryAll(ctx context.Context, query string, args ...interface{}) ([]*domain.SelectorConfig, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query selector configs: %w", err)
	}
	defer rows.Close()

	var configs []*domain.SelectorConfig
	for rows.Next() {
		var config domain.SelectorConfig
		if err := rows.Scan(&config.ID, &config.Name, &config.Domain, &config.Selector, &config.Regex, &config.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan selector config: %w", err)
		}
		configs = append(configs, &config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return configs, nil
}
