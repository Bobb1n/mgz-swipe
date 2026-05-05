package grpctransport

import (
	"context"
	"errors"

	"swipe-mgz/internal/domain"
	"swipe-mgz/internal/usecase"
	swipev1 "swipe-mgz/pkg/api/swipe/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SwipeService interface {
	Swipe(ctx context.Context, swiperID, swipeeID string, dir domain.Direction) (*usecase.SwipeResult, error)
	ListMatches(ctx context.Context, userID string, limit, offset int) ([]*domain.Match, error)
	UpdateLocation(ctx context.Context, userID string, lon, lat float64) error
	GetCandidates(ctx context.Context, userID string) ([]*domain.Candidate, error)
}

type Server struct {
	swipev1.UnimplementedSwipeServiceServer
	uc SwipeService
}

func NewServer(uc SwipeService) *Server {
	return &Server{uc: uc}
}

func (s *Server) Swipe(ctx context.Context, req *swipev1.SwipeRequest) (*swipev1.SwipeResponse, error) {
	dir := fromProtoDirection(req.GetDirection())
	result, err := s.uc.Swipe(ctx, req.GetSwiperId(), req.GetSwipeeId(), dir)
	if err != nil {
		return nil, mapUseCaseError(err)
	}

	resp := &swipev1.SwipeResponse{Swipe: toProtoSwipe(result.Swipe)}
	if result.Match != nil {
		resp.Match = toProtoMatch(result.Match, req.GetSwiperId())
	}
	return resp, nil
}

func (s *Server) ListMatches(ctx context.Context, req *swipev1.ListMatchesRequest) (*swipev1.ListMatchesResponse, error) {
	limit := int(req.GetLimit())
	offset := int(req.GetOffset())

	matches, err := s.uc.ListMatches(ctx, req.GetUserId(), limit, offset)
	if err != nil {
		return nil, mapUseCaseError(err)
	}

	if limit <= 0 {
		limit = usecase.DefaultListLimit
	}
	if offset < 0 {
		offset = 0
	}

	out := make([]*swipev1.Match, 0, len(matches))
	for _, m := range matches {
		out = append(out, toProtoMatch(m, req.GetUserId()))
	}
	return &swipev1.ListMatchesResponse{
		Matches: out,
		Limit:   int32(limit),
		Offset:  int32(offset),
	}, nil
}

func (s *Server) UpdateLocation(ctx context.Context, req *swipev1.UpdateLocationRequest) (*swipev1.UpdateLocationResponse, error) {
	if err := s.uc.UpdateLocation(ctx, req.GetUserId(), req.GetLongitude(), req.GetLatitude()); err != nil {
		return nil, mapUseCaseError(err)
	}
	return &swipev1.UpdateLocationResponse{}, nil
}

func (s *Server) GetCandidates(ctx context.Context, req *swipev1.GetCandidatesRequest) (*swipev1.GetCandidatesResponse, error) {
	cands, err := s.uc.GetCandidates(ctx, req.GetUserId())
	if err != nil {
		return nil, mapUseCaseError(err)
	}

	out := make([]*swipev1.Candidate, 0, len(cands))
	for _, c := range cands {
		out = append(out, &swipev1.Candidate{
			UserId:     c.UserID,
			Longitude:  c.Longitude,
			Latitude:   c.Latitude,
			DistanceKm: c.DistKm,
		})
	}
	return &swipev1.GetCandidatesResponse{Candidates: out}, nil
}

func fromProtoDirection(d swipev1.Direction) domain.Direction {
	switch d {
	case swipev1.Direction_DIRECTION_LIKE:
		return domain.DirectionLike
	case swipev1.Direction_DIRECTION_DISLIKE:
		return domain.DirectionDislike
	default:
		return ""
	}
}

func toProtoDirection(d domain.Direction) swipev1.Direction {
	switch d {
	case domain.DirectionLike:
		return swipev1.Direction_DIRECTION_LIKE
	case domain.DirectionDislike:
		return swipev1.Direction_DIRECTION_DISLIKE
	default:
		return swipev1.Direction_DIRECTION_UNSPECIFIED
	}
}

func toProtoSwipe(s *domain.Swipe) *swipev1.Swipe {
	if s == nil {
		return nil
	}
	return &swipev1.Swipe{
		Id:        s.ID,
		SwiperId:  s.SwiperID,
		SwipeeId:  s.SwipeeID,
		Direction: toProtoDirection(s.Direction),
		CreatedAt: timestamppb.New(s.CreatedAt),
	}
}

func toProtoMatch(m *domain.Match, viewerUserID string) *swipev1.Match {
	if m == nil {
		return nil
	}
	pm := &swipev1.Match{
		Id:        m.ID,
		User1Id:   m.User1ID,
		User2Id:   m.User2ID,
		CreatedAt: timestamppb.New(m.CreatedAt),
	}
	switch viewerUserID {
	case m.User1ID:
		pm.PeerUserId = m.User2ID
	case m.User2ID:
		pm.PeerUserId = m.User1ID
	}
	return pm
}

func mapUseCaseError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, usecase.ErrAlreadySwiped):
		return status.Error(codes.AlreadyExists, err.Error())
	case usecase.IsValidationError(err):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.Canceled, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
