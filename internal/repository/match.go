package repository

import (
	"context"
	"fmt"

	"swipe-mgz/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MatchRepo struct {
	db *pgxpool.Pool
}

func NewMatchRepo(db *pgxpool.Pool) *MatchRepo {
	return &MatchRepo{db: db}
}

func (r *MatchRepo) Create(ctx context.Context, user1ID, user2ID string) (*domain.Match, error) {
	m := &domain.Match{}
	err := r.db.QueryRow(ctx,
		`WITH ins AS (
			INSERT INTO matches (user1_id, user2_id)
			VALUES ($1, $2)
			ON CONFLICT (user1_id, user2_id) DO NOTHING
			RETURNING id, user1_id, user2_id, created_at
		)
		SELECT id, user1_id, user2_id, created_at FROM ins
		UNION ALL
		SELECT id, user1_id, user2_id, created_at FROM matches WHERE user1_id = $1 AND user2_id = $2
		LIMIT 1`,
		user1ID, user2ID,
	).Scan(&m.ID, &m.User1ID, &m.User2ID, &m.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create match: %w", err)
	}
	return m, nil
}

func (r *MatchRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Match, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user1_id, user2_id, created_at
		 FROM matches
		 WHERE user1_id = $1 OR user2_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list matches: %w", err)
	}
	defer rows.Close()

	var matches []*domain.Match
	for rows.Next() {
		m := &domain.Match{}
		if err := rows.Scan(&m.ID, &m.User1ID, &m.User2ID, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan match: %w", err)
		}
		matches = append(matches, m)
	}
	return matches, rows.Err()
}
