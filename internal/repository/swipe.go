package repository

import (
	"context"
	"errors"
	"fmt"

	"swipe-mgz/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAlreadySwiped = errors.New("already swiped")

type SwipeRepo struct {
	db *pgxpool.Pool
}

func NewSwipeRepo(db *pgxpool.Pool) *SwipeRepo {
	return &SwipeRepo{db: db}
}

func (r *SwipeRepo) Create(ctx context.Context, swipe *domain.Swipe) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO swipes (swiper_id, swipee_id, direction)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (swiper_id, swipee_id) DO NOTHING`,
		swipe.SwiperID, swipe.SwipeeID, string(swipe.Direction),
	)
	if err != nil {
		return fmt.Errorf("create swipe: %w", err)
	}
	return nil
}

// HasMutualLike checks whether swipee_id has already liked swiper_id.
func (r *SwipeRepo) HasMutualLike(ctx context.Context, swiperID, swipeeID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM swipes
			WHERE swiper_id = $1 AND swipee_id = $2 AND direction = 'like'
		)`,
		swipeeID, swiperID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check mutual like: %w", err)
	}
	return exists, nil
}

func (r *SwipeRepo) GetByUsers(ctx context.Context, swiperID, swipeeID string) (*domain.Swipe, error) {
	row := r.db.QueryRow(ctx,
		`SELECT id, swiper_id, swipee_id, direction, created_at
		 FROM swipes WHERE swiper_id = $1 AND swipee_id = $2`,
		swiperID, swipeeID,
	)
	s := &domain.Swipe{}
	err := row.Scan(&s.ID, &s.SwiperID, &s.SwipeeID, &s.Direction, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get swipe: %w", err)
	}
	return s, nil
}

// AlreadySwiped reports whether swiperID has already swiped swipeeID.
func (r *SwipeRepo) AlreadySwiped(ctx context.Context, swiperID, swipeeID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM swipes WHERE swiper_id = $1 AND swipee_id = $2)`,
		swiperID, swipeeID,
	).Scan(&exists)
	return exists, err
}
