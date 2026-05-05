package domain

import (
	"errors"
	"time"
)

// ErrSwipeDuplicate — INSERT свайпа не создал строку (гонка или UNIQUE).
var ErrSwipeDuplicate = errors.New("swipe already exists for this pair")

type Direction string

const (
	DirectionLike    Direction = "like"
	DirectionDislike Direction = "dislike"
)

type Swipe struct {
	ID        int64
	SwiperID  string
	SwipeeID  string
	Direction Direction
	CreatedAt time.Time
}

type Match struct {
	ID        int64
	User1ID   string
	User2ID   string
	CreatedAt time.Time
}

type Candidate struct {
	UserID    string
	Longitude float64
	Latitude  float64
	DistKm    float64
}
