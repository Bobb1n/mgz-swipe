package domain

import "time"

type Direction string

const (
	DirectionLike    Direction = "like"
	DirectionDislike Direction = "dislike"
)

type Swipe struct {
	ID        int64     `json:"id"`
	SwiperID  string    `json:"swiper_id"`
	SwipeeID  string    `json:"swipee_id"`
	Direction Direction `json:"direction"`
	CreatedAt time.Time `json:"created_at"`
}

type Match struct {
	ID        int64     `json:"id"`
	User1ID   string    `json:"user1_id"`
	User2ID   string    `json:"user2_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Candidate struct {
	UserID    string  `json:"user_id"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
	DistKm    float64 `json:"distance_km"`
}
