package usecase

import (
	"context"
	"fmt"

	"swipe-mgz/internal/domain"
)

type SwipeUseCase struct {
	swipeRepo    SwipeRepository
	matchRepo    MatchRepository
	locationRepo LocationRepository
	publisher    EventPublisher
	radiusKm     float64
}

func NewSwipeUseCase(
	swipeRepo SwipeRepository,
	matchRepo MatchRepository,
	locationRepo LocationRepository,
	publisher EventPublisher,
	radiusKm float64,
) *SwipeUseCase {
	return &SwipeUseCase{
		swipeRepo:    swipeRepo,
		matchRepo:    matchRepo,
		locationRepo: locationRepo,
		publisher:    publisher,
		radiusKm:     radiusKm,
	}
}

type SwipeResult struct {
	Swipe *domain.Swipe `json:"swipe"`
	Match *domain.Match `json:"match,omitempty"`
}

func (uc *SwipeUseCase) Swipe(ctx context.Context, swiperID, swipeeID string, dir domain.Direction) (*SwipeResult, error) {
	if swiperID == swipeeID {
		return nil, fmt.Errorf("cannot swipe yourself")
	}

	already, err := uc.swipeRepo.AlreadySwiped(ctx, swiperID, swipeeID)
	if err != nil {
		return nil, fmt.Errorf("check existing swipe: %w", err)
	}
	if already {
		return nil, fmt.Errorf("already swiped this user")
	}

	swipe := &domain.Swipe{
		SwiperID:  swiperID,
		SwipeeID:  swipeeID,
		Direction: dir,
	}
	if err := uc.swipeRepo.Create(ctx, swipe); err != nil {
		return nil, err
	}

	loaded, err := uc.swipeRepo.GetByUsers(ctx, swiperID, swipeeID)
	if err != nil || loaded == nil {
		return nil, fmt.Errorf("reload swipe: %w", err)
	}
	swipe = loaded

	go func(s *domain.Swipe) { _ = uc.publisher.PublishSwipe(context.Background(), s) }(swipe)

	result := &SwipeResult{Swipe: swipe}

	if dir == domain.DirectionLike {
		mutual, err := uc.swipeRepo.HasMutualLike(ctx, swiperID, swipeeID)
		if err != nil {
			return result, nil
		}
		if mutual {
			u1, u2 := swiperID, swipeeID
			if u1 > u2 {
				u1, u2 = u2, u1
			}
			match, err := uc.matchRepo.Create(ctx, u1, u2)
			if err == nil {
				result.Match = match
				go func(m *domain.Match) { _ = uc.publisher.PublishMatch(context.Background(), m) }(match)
			}
		}
	}

	return result, nil
}

func (uc *SwipeUseCase) ListMatches(ctx context.Context, userID string, limit, offset int) ([]*domain.Match, error) {
	return uc.matchRepo.ListByUser(ctx, userID, limit, offset)
}

func (uc *SwipeUseCase) UpdateLocation(ctx context.Context, userID string, lon, lat float64) error {
	return uc.locationRepo.Update(ctx, userID, lon, lat)
}

func (uc *SwipeUseCase) GetCandidates(ctx context.Context, userID string) ([]*domain.Candidate, error) {
	return uc.locationRepo.Candidates(ctx, userID, uc.radiusKm)
}
