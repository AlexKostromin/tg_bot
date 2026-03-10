package mq

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
)

type NotificationHandler interface {
	HandleNewBooking(ctx context.Context, bookingID int) error
	HandleBookingStatusChanged(ctx context.Context, bookingID int, status string) error
}

type Consumer struct {
	ch      *amqp.Channel
	handler NotificationHandler
}

func NewConsumer(conn *amqp.Connection, handler NotificationHandler) *Consumer {
	ch, err := conn.Channel()
	if err != nil {
		panic("mq: open consumer channel: " + err.Error())
	}
	return &Consumer{ch: ch, handler: handler}
}

func (c *Consumer) Run(ctx context.Context) {
	// Объявляем очередь — idempotent, не пересоздаёт если уже есть
	_, err := c.ch.QueueDeclare("notifications", true, false, false, false, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("mq: queue declare failed")
	}

	msgs, err := c.ch.Consume("notifications", "", false, false, false, false, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("mq: consume failed")
	}

	for {
		select {
		case <-ctx.Done():
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			c.processWithRetry(ctx, d, func() error {
				switch d.RoutingKey {
				case "new_booking":
					var evt BookingEvent
					if err := json.Unmarshal(d.Body, &evt); err != nil {
						return err
					}
					return c.handler.HandleNewBooking(ctx, evt.BookingID)
				case "booking_status_changed":
					var evt BookingStatusEvent
					if err := json.Unmarshal(d.Body, &evt); err != nil {
						return err
					}
					return c.handler.HandleBookingStatusChanged(ctx, evt.BookingID, evt.Status)
				default:
					return nil
				}
			})
		}
	}
}

func (c *Consumer) processWithRetry(ctx context.Context, d amqp.Delivery, handler func() error) {
	const maxRetries = 3

	retryCount := int64(0)
	if v, ok := d.Headers["x-retry-count"]; ok {
		switch n := v.(type) {
		case int64:
			retryCount = n
		case int32:
			retryCount = int64(n)
		}
	}

	if err := handler(); err != nil {
		log.Error().Err(err).Msg("handler failed")
		if retryCount < maxRetries {
			headers := amqp.Table{"x-retry-count": retryCount + 1}
			c.ch.PublishWithContext(ctx, d.Exchange, d.RoutingKey, false, false,
				amqp.Publishing{
					ContentType: d.ContentType,
					Body:        d.Body,
					Headers:     headers,
				})
		} else {
			c.ch.PublishWithContext(ctx, "", "failed.notifications", false, false,
				amqp.Publishing{ContentType: d.ContentType, Body: d.Body})
		}
		d.Nack(false, false)
		return
	}
	d.Ack(false)
}
