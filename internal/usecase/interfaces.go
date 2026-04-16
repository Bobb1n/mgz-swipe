package usecase

import (
	"context"

	"swipe-mgz/internal/domain"
)

type SwipeRepository interface {
	Create(ctx context.Context, swipe *domain.Swipe) error
	GetByUsers(ctx context.Context, swiperID, swipeeID string) (*domain.Swipe, error)
	HasMutualLike(ctx context.Context, swiperID, swipeeID string) (bool, error)
	AlreadySwiped(ctx context.Context, swiperID, swipeeID string) (bool, error)
}

type MatchRepository interface {
	Create(ctx context.Context, user1ID, user2ID string) (*domain.Match, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Match, error)
}

type LocationRepository interface {
	Update(ctx context.Context, userID string, lon, lat float64) error
	Candidates(ctx context.Context, userID string, radiusKm float64) ([]*domain.Candidate, error)
}

type EventPublisher interface {
	PublishSwipe(ctx context.Context, s *domain.Swipe) error
	PublishMatch(ctx context.Context, m *domain.Match) error
}
