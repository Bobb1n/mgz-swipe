package http

import (
	"net/http"
	"strconv"

	"swipe-mgz/internal/domain"
	"swipe-mgz/internal/usecase"

	"github.com/labstack/echo/v4"
)

const headerUserID = "X-User-Id"

type Handler struct {
	uc *usecase.SwipeUseCase
}

func NewHandler(uc *usecase.SwipeUseCase) *Handler {
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

func currentUserID(c echo.Context) (string, bool) {
	id := c.Request().Header.Get(headerUserID)
	return id, id != ""
}

func errJSON(c echo.Context, code int, msg string) error {
	return c.JSON(code, map[string]string{"error": msg})
}

// POST /v1/swipes
// Body: {"swipee_id": "uuid", "direction": "like"|"dislike"}
func (h *Handler) swipe(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return errJSON(c, http.StatusUnauthorized, "missing user id")
	}

	var body struct {
		SwipeeID  string `json:"swipee_id"`
		Direction string `json:"direction"`
	}
	if err := c.Bind(&body); err != nil {
		return errJSON(c, http.StatusBadRequest, "invalid json")
	}
	if body.SwipeeID == "" {
		return errJSON(c, http.StatusBadRequest, "swipee_id is required")
	}
	dir := domain.Direction(body.Direction)
	if dir != domain.DirectionLike && dir != domain.DirectionDislike {
		return errJSON(c, http.StatusBadRequest, "direction must be 'like' or 'dislike'")
	}

	result, err := h.uc.Swipe(c.Request().Context(), userID, body.SwipeeID, dir)
	if err != nil {
		return errJSON(c, http.StatusBadRequest, err.Error())
	}
	return c.JSON(http.StatusOK, result)
}

// GET /v1/matches?limit=20&offset=0
func (h *Handler) listMatches(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return errJSON(c, http.StatusUnauthorized, "missing user id")
	}

	limit, offset := 20, 0
	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := c.QueryParam("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	matches, err := h.uc.ListMatches(c.Request().Context(), userID, limit, offset)
	if err != nil {
		return errJSON(c, http.StatusInternalServerError, "internal error")
	}
	if matches == nil {
		matches = []*domain.Match{}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"matches": matches,
		"limit":   limit,
		"offset":  offset,
	})
}

// PUT /v1/location
// Body: {"longitude": 37.6, "latitude": 55.7}
func (h *Handler) updateLocation(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return errJSON(c, http.StatusUnauthorized, "missing user id")
	}

	var body struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	}
	if err := c.Bind(&body); err != nil {
		return errJSON(c, http.StatusBadRequest, "invalid json")
	}
	if body.Longitude < -180 || body.Longitude > 180 {
		return errJSON(c, http.StatusBadRequest, "longitude must be between -180 and 180")
	}
	if body.Latitude < -90 || body.Latitude > 90 {
		return errJSON(c, http.StatusBadRequest, "latitude must be between -90 and 90")
	}

	if err := h.uc.UpdateLocation(c.Request().Context(), userID, body.Longitude, body.Latitude); err != nil {
		return errJSON(c, http.StatusInternalServerError, "internal error")
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// GET /v1/candidates
func (h *Handler) getCandidates(c echo.Context) error {
	userID, ok := currentUserID(c)
	if !ok {
		return errJSON(c, http.StatusUnauthorized, "missing user id")
	}

	candidates, err := h.uc.GetCandidates(c.Request().Context(), userID)
	if err != nil {
		return errJSON(c, http.StatusInternalServerError, "internal error")
	}
	if candidates == nil {
		candidates = []*domain.Candidate{}
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"candidates": candidates,
	})
}
