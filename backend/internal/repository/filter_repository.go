package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// FilterRule represents a single filter rule
type FilterRule struct {
	ID       string `json:"id"`
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value,omitempty"`
}

// SavedFilter represents a user's saved filter
type SavedFilter struct {
	ID        string       `json:"id" db:"id"`
	UserID    string       `json:"user_id" db:"user_id"`
	Name      string       `json:"name" db:"name"`
	Rules     []FilterRule `json:"rules" db:"rules"`
	Position  int          `json:"position" db:"position"`
	CreatedAt time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt time.Time    `json:"updated_at" db:"updated_at"`
}

// FilterRepository handles saved filters database operations
type FilterRepository struct {
	db *sql.DB
}

// NewFilterRepository creates a new filter repository
func NewFilterRepository(db *sql.DB) *FilterRepository {
	return &FilterRepository{db: db}
}

// List returns all saved filters for a user
func (r *FilterRepository) List(ctx context.Context, userID string) ([]*SavedFilter, error) {
	query := `
		SELECT id, user_id, name, rules, position, created_at, updated_at
		FROM saved_filters
		WHERE user_id = $1
		ORDER BY position ASC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var filters []*SavedFilter
	for rows.Next() {
		var filter SavedFilter
		var rulesJSON []byte
		
		err := rows.Scan(
			&filter.ID, &filter.UserID, &filter.Name, &rulesJSON,
			&filter.Position, &filter.CreatedAt, &filter.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(rulesJSON, &filter.Rules); err != nil {
			filter.Rules = []FilterRule{}
		}

		filters = append(filters, &filter)
	}

	return filters, nil
}

// GetByID returns a filter by ID
func (r *FilterRepository) GetByID(ctx context.Context, userID, filterID string) (*SavedFilter, error) {
	query := `
		SELECT id, user_id, name, rules, position, created_at, updated_at
		FROM saved_filters
		WHERE id = $1 AND user_id = $2
	`

	var filter SavedFilter
	var rulesJSON []byte

	err := r.db.QueryRowContext(ctx, query, filterID, userID).Scan(
		&filter.ID, &filter.UserID, &filter.Name, &rulesJSON,
		&filter.Position, &filter.CreatedAt, &filter.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(rulesJSON, &filter.Rules); err != nil {
		filter.Rules = []FilterRule{}
	}

	return &filter, nil
}

// Create creates a new saved filter
func (r *FilterRepository) Create(ctx context.Context, filter *SavedFilter) error {
	// Get next position
	var maxPos int
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(position), -1) FROM saved_filters WHERE user_id = $1
	`, filter.UserID).Scan(&maxPos)
	if err != nil {
		return err
	}
	filter.Position = maxPos + 1

	// Serialize rules
	rulesJSON, err := json.Marshal(filter.Rules)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO saved_filters (user_id, name, rules, position)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRowContext(ctx, query,
		filter.UserID, filter.Name, rulesJSON, filter.Position,
	).Scan(&filter.ID, &filter.CreatedAt, &filter.UpdatedAt)
}

// Update updates a saved filter
func (r *FilterRepository) Update(ctx context.Context, filter *SavedFilter) error {
	rulesJSON, err := json.Marshal(filter.Rules)
	if err != nil {
		return err
	}

	query := `
		UPDATE saved_filters
		SET name = $1, rules = $2, updated_at = NOW()
		WHERE id = $3 AND user_id = $4
	`

	result, err := r.db.ExecContext(ctx, query, filter.Name, rulesJSON, filter.ID, filter.UserID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete deletes a saved filter
func (r *FilterRepository) Delete(ctx context.Context, userID, filterID string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM saved_filters WHERE id = $1 AND user_id = $2
	`, filterID, userID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Reorder updates the position of all filters
func (r *FilterRepository) Reorder(ctx context.Context, userID string, filterIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i, id := range filterIDs {
		_, err := tx.ExecContext(ctx, `
			UPDATE saved_filters
			SET position = $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3
		`, i, id, userID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
