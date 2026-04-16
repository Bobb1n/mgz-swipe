package usecase

import (
	"context"
	"errors"
	"fmt"

	"swipe-mgz/internal/domain"
)

var (
	ErrSelfSwipe        = errors.New("cannot swipe yourself")
	ErrAlreadySwiped    = errors.New("already swiped this user")
	ErrInvalidDirection = errors.New("direction must be 'like' or 'dislike'")
	ErrInvalidUserID    = errors.New("user id is required")
	ErrInvalidLongitude = errors.New("longitude must be between -180 and 180")
	ErrInvalidLatitude  = errors.New("latitude must be between -90 and 90")
)

const (
	DefaultListLimit = 20
	MaxListLimit     = 100
)

type SwipeResult struct {
	Swipe *domain.Swipe
	Match *domain.Match
}

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

func (uc *SwipeUseCase) Swipe(ctx context.Context, swiperID, swipeeID string, dir domain.Direction) (*SwipeResult, error) {
	if swiperID == "" || swipeeID == "" {
		return nil, ErrInvalidUserID
	}
	if dir != domain.DirectionLike && dir != domain.DirectionDislike {
		return nil, ErrInvalidDirection
	}
	if swiperID == swipeeID {
		return nil, ErrSelfSwipe
	}

	already, err := uc.swipeRepo.AlreadySwiped(ctx, swiperID, swipeeID)
	if err != nil {
		return nil, fmt.Errorf("check existing swipe: %w", err)
	}
	if already {
		return nil, ErrAlreadySwiped
	}

	swipe := &domain.Swipe{
		SwiperID:  swiperID,
		SwipeeID:  swipeeID,
		Direction: dir,
	}
	if err := uc.swipeRepo.Create(ctx, swipe); err != nil {
		return nil, fmt.Errorf("create swipe: %w", err)
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
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if limit <= 0 {
		limit = DefaultListLimit
	}
	if limit > MaxListLimit {
		limit = MaxListLimit
	}
	if offset < 0 {
		offset = 0
	}
	return uc.matchRepo.ListByUser(ctx, userID, limit, offset)
}

func (uc *SwipeUseCase) UpdateLocation(ctx context.Context, userID string, lon, lat float64) error {
	if userID == "" {
		return ErrInvalidUserID
	}
	if lon < -180 || lon > 180 {
		return ErrInvalidLongitude
	}
	if lat < -90 || lat > 90 {
		return ErrInvalidLatitude
	}
	return uc.locationRepo.Update(ctx, userID, lon, lat)
}

func (uc *SwipeUseCase) GetCandidates(ctx context.Context, userID string) ([]*domain.Candidate, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	return uc.locationRepo.Candidates(ctx, userID, uc.radiusKm)
}

func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidUserID) ||
		errors.Is(err, ErrInvalidDirection) ||
		errors.Is(err, ErrInvalidLongitude) ||
		errors.Is(err, ErrInvalidLatitude) ||
		errors.Is(err, ErrSelfSwipe) ||
		errors.Is(err, ErrAlreadySwiped)
}
