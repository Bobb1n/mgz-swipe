package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"swipe-mgz/internal/domain"

	kafka "github.com/segmentio/kafka-go"
)

const (
	TopicSwipe = "swipe.events"
	TopicMatch = "match.events"
)

type swipeEvent struct {
	SwiperID  string    `json:"swiper_id"`
	SwipeeID  string    `json:"swipee_id"`
	Direction string    `json:"direction"`
	CreatedAt time.Time `json:"created_at"`
}

type matchEvent struct {
	MatchID   int64     `json:"match_id"`
	User1ID   string    `json:"user1_id"`
	User2ID   string    `json:"user2_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Publisher struct {
	swipeWriter *kafka.Writer
	matchWriter *kafka.Writer
}

func NewPublisher(brokers string) *Publisher {
	addrs := strings.Split(brokers, ",")
	return &Publisher{
		swipeWriter: &kafka.Writer{
			Addr:                   kafka.TCP(addrs...),
			Topic:                  TopicSwipe,
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
		matchWriter: &kafka.Writer{
			Addr:                   kafka.TCP(addrs...),
			Topic:                  TopicMatch,
			Balancer:               &kafka.LeastBytes{},
			AllowAutoTopicCreation: true,
		},
	}
}

func publish(w *kafka.Writer, key, value []byte, topic string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := w.WriteMessages(ctx, kafka.Message{Key: key, Value: value})
	if err != nil {
		slog.Error("kafka publish failed", "topic", topic, "error", err)
	}
	return err
}

func (p *Publisher) PublishSwipe(ctx context.Context, s *domain.Swipe) error {
	payload, err := json.Marshal(swipeEvent{
		SwiperID:  s.SwiperID,
		SwipeeID:  s.SwipeeID,
		Direction: string(s.Direction),
		CreatedAt: s.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("marshal swipe event: %w", err)
	}
	return publish(p.swipeWriter, []byte(s.SwiperID), payload, TopicSwipe)
}

func (p *Publisher) PublishMatch(ctx context.Context, m *domain.Match) error {
	payload, err := json.Marshal(matchEvent{
		MatchID:   m.ID,
		User1ID:   m.User1ID,
		User2ID:   m.User2ID,
		CreatedAt: m.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("marshal match event: %w", err)
	}
	return publish(p.matchWriter, []byte(m.User1ID), payload, TopicMatch)
}

func (p *Publisher) Close() error {
	err1 := p.swipeWriter.Close()
	err2 := p.matchWriter.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
