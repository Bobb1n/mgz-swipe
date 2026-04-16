package http

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"swipe-mgz/internal/domain"
	"swipe-mgz/internal/usecase"

	"github.com/labstack/echo/v4"
)

const headerUserID = "X-User-Id"

type SwipeService interface {
	Swipe(ctx context.Context, swiperID, swipeeID string, dir domain.Direction) (*usecase.SwipeResult, error)
	ListMatches(ctx context.Context, userID string, limit, offset int) ([]*domain.Match, error)
	UpdateLocation(ctx context.Context, userID string, lon, lat float64) error
	GetCandidates(ctx context.Context, userID string) ([]*domain.Candidate, error)
}

type Handler struct {
	uc SwipeService
}

func NewHandler(uc SwipeService) *Handler {
	return &Handler{uc: uc}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "service": "swipe-mgz"})
	})

	v1 := e.Group("/v1")
	v1.POST("/swipes", h.swipe)
	v1.GET("/matches", h.listMatches)
	v1.PUT("/location", h.updateLocation)
	v1.GET("/candidates", h.getCandidates)
}

func (h *Handler) swipe(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return writeErr(c, http.StatusUnauthorized, "missing user id")
	}

	var req swipeRequest
	if err := c.Bind(&req); err != nil {
		return writeErr(c, http.StatusBadRequest, "invalid json")
	}

	result, err := h.uc.Swipe(c.Request().Context(), userID, req.SwipeeID, domain.Direction(req.Direction))
	if err != nil {
		return mapUseCaseError(c, err)
	}
	return c.JSON(http.StatusOK, toSwipeResponse(result))
}

func (h *Handler) listMatches(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return writeErr(c, http.StatusUnauthorized, "missing user id")
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	matches, err := h.uc.ListMatches(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return mapUseCaseError(c, err)
	}

	out := make([]*matchDTO, 0, len(matches))
	for _, m := range matches {
		out = append(out, toMatchDTO(m))
	}
	if limit <= 0 {
		limit = usecase.DefaultListLimit
	}
	if offset < 0 {
		offset = 0
	}
	return c.JSON(http.StatusOK, matchesResponse{Matches: out, Limit: limit, Offset: offset})
}

func (h *Handler) updateLocation(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return writeErr(c, http.StatusUnauthorized, "missing user id")
	}

	var req updateLocationRequest
	if err := c.Bind(&req); err != nil {
		return writeErr(c, http.StatusBadRequest, "invalid json")
	}

	if err := h.uc.UpdateLocation(c.Request().Context(), userID, req.Longitude, req.Latitude); err != nil {
		return mapUseCaseError(c, err)
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) getCandidates(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return writeErr(c, http.StatusUnauthorized, "missing user id")
	}

	candidates, err := h.uc.GetCandidates(c.Request().Context(), userID)
	if err != nil {
		return mapUseCaseError(c, err)
	}

	out := make([]*candidateDTO, 0, len(candidates))
	for _, cand := range candidates {
		out = append(out, toCandidateDTO(cand))
	}
	return c.JSON(http.StatusOK, candidatesResponse{Candidates: out})
}

func currentUserID(c echo.Context) (string, bool) {
	id := c.Request().Header.Get(headerUserID)
	return id, id != ""
}

func writeErr(c echo.Context, code int, msg string) error {
	return c.JSON(code, errorResponse{Error: msg})
}

func mapUseCaseError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, usecase.ErrAlreadySwiped):
		return writeErr(c, http.StatusConflict, err.Error())
	case usecase.IsValidationError(err):
		return writeErr(c, http.StatusBadRequest, err.Error())
	default:
		return writeErr(c, http.StatusInternalServerError, "internal error")
	}
}
