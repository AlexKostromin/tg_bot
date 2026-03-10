package mq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	ch *amqp.Channel
}

func NewPublisher(conn *amqp.Connection) *Publisher {
	ch, err := conn.Channel()
	if err != nil {
		panic("mq: open channel: " + err.Error())
	}
	// Объявляем exchange для уведомлений
	ch.ExchangeDeclare("notifications", "direct", true, false, false, false, nil)
	return &Publisher{ch: ch}
}

type BookingEvent struct {
	BookingID int `json:"booking_id"`
}

type ReminderEvent struct {
	BookingID int    `json:"booking_id"`
	Kind      string `json:"kind"`
}

type BookingStatusEvent struct {
	BookingID int    `json:"booking_id"`
	Status    string `json:"status"`
}

func (p *Publisher) PublishBookingCreated(ctx context.Context, bookingID int) error {
	body, _ := json.Marshal(BookingEvent{BookingID: bookingID})
	return p.ch.PublishWithContext(ctx, "notifications", "new_booking", false, false,
		amqp.Publishing{ContentType: "application/json", Body: body})
}

func (p *Publisher) PublishBookingStatusChanged(ctx context.Context, bookingID int, status string) error {
	body, _ := json.Marshal(BookingStatusEvent{BookingID: bookingID, Status: status})
	return p.ch.PublishWithContext(ctx, "notifications", "booking_status_changed", false, false,
		amqp.Publishing{ContentType: "application/json", Body: body})
}

func (p *Publisher) PublishReminder(ctx context.Context, bookingID int, kind string, delayMs int64) error {
	body, _ := json.Marshal(ReminderEvent{BookingID: bookingID, Kind: kind})
	return p.ch.PublishWithContext(ctx, "reminders", "", false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Headers:     amqp.Table{"x-delay": delayMs},
		})
}
