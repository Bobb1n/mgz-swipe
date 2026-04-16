package repository

import (
	"context"
	"fmt"

	"swipe-mgz/internal/domain"

	"github.com/redis/go-redis/v9"
)

const geoKey = "geo:users"

type LocationRepo struct {
	rdb *redis.Client
}

func NewLocationRepo(rdb *redis.Client) *LocationRepo {
	return &LocationRepo{rdb: rdb}
}

func (r *LocationRepo) Update(ctx context.Context, userID string, lon, lat float64) error {
	err := r.rdb.GeoAdd(ctx, geoKey, &redis.GeoLocation{
		Name:      userID,
		Longitude: lon,
		Latitude:  lat,
	}).Err()
	if err != nil {
		return fmt.Errorf("geo add: %w", err)
	}
	return nil
}

func (r *LocationRepo) Candidates(ctx context.Context, userID string, radiusKm float64) ([]*domain.Candidate, error) {
	results, err := r.rdb.GeoSearchLocation(ctx, geoKey, &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Member:     userID,
			Radius:     radiusKm,
			RadiusUnit: "km",
			Sort:       "ASC",
			Count:      200,
		},
		WithDist:  true,
		WithCoord: true,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("geo search: %w", err)
	}

	candidates := make([]*domain.Candidate, 0, len(results))
	for _, loc := range results {
		if loc.Name == userID {
			continue
		}
		candidates = append(candidates, &domain.Candidate{
			UserID:    loc.Name,
			Longitude: loc.Longitude,
			Latitude:  loc.Latitude,
			DistKm:    loc.Dist,
		})
	}
	return candidates, nil
}
