package http

import (
	"time"

	"swipe-mgz/internal/domain"
	"swipe-mgz/internal/usecase"
)

type swipeRequest struct {
	SwipeeID  string `json:"swipee_id"`
	Direction string `json:"direction"`
}

type updateLocationRequest struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

type swipeDTO struct {
	ID        int64     `json:"id"`
	SwiperID  string    `json:"swiper_id"`
	SwipeeID  string    `json:"swipee_id"`
	Direction string    `json:"direction"`
	CreatedAt time.Time `json:"created_at"`
}

type matchDTO struct {
	ID         int64     `json:"id"`
	User1ID    string    `json:"user1_id"`
	User2ID    string    `json:"user2_id"`
	PeerUserID string    `json:"peer_user_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type candidateDTO struct {
	UserID    string  `json:"user_id"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	DistKm    float64 `json:"distance_km"`
}

type swipeResponse struct {
	Swipe *swipeDTO `json:"swipe"`
	Match *matchDTO `json:"match,omitempty"`
}

type matchesResponse struct {
	Matches []*matchDTO `json:"matches"`
	Limit   int         `json:"limit"`
	Offset  int         `json:"offset"`
}

type candidatesResponse struct {
	Candidates []*candidateDTO `json:"candidates"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func toSwipeDTO(s *domain.Swipe) *swipeDTO {
	if s == nil {
		return nil
	}
	return &swipeDTO{
		ID:        s.ID,
		SwiperID:  s.SwiperID,
		SwipeeID:  s.SwipeeID,
		Direction: string(s.Direction),
		CreatedAt: s.CreatedAt,
	}
}

func toMatchDTO(m *domain.Match) *matchDTO {
	if m == nil {
		return nil
	}
	return &matchDTO{
		ID:        m.ID,
		User1ID:   m.User1ID,
		User2ID:   m.User2ID,
		CreatedAt: m.CreatedAt,
	}
}

// toMatchDTOForViewer заполняет peer_user_id — второй участник матча относительно текущего пользователя.
func toMatchDTOForViewer(m *domain.Match, viewerUserID string) *matchDTO {
	dto := toMatchDTO(m)
	if dto == nil || viewerUserID == "" {
		return dto
	}
	switch viewerUserID {
	case m.User1ID:
		dto.PeerUserID = m.User2ID
	case m.User2ID:
		dto.PeerUserID = m.User1ID
	}
	return dto
}

func toCandidateDTO(c *domain.Candidate) *candidateDTO {
	if c == nil {
		return nil
	}
	return &candidateDTO{
		UserID:    c.UserID,
		Longitude: c.Longitude,
		Latitude:  c.Latitude,
		DistKm:    c.DistKm,
	}
}

func toSwipeResponse(r *usecase.SwipeResult, viewerUserID string) *swipeResponse {
	if r == nil {
		return nil
	}
	var match *matchDTO
	if r.Match != nil {
		match = toMatchDTOForViewer(r.Match, viewerUserID)
	}
	return &swipeResponse{
		Swipe: toSwipeDTO(r.Swipe),
		Match: match,
	}
}
