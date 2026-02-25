package rabbit

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"logs-ingest/internal/app"
)

type Publisher struct{ ch *amqp.Channel }

func New(url string) (*Publisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	_, err = ch.QueueDeclare("log_events", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}
	return &Publisher{ch: ch}, nil
}

func (p *Publisher) Publish(ctx context.Context, evt app.LogEvent) error {
	body, _ := json.Marshal(evt)
	return p.ch.PublishWithContext(ctx, "", "log_events", false, false, amqp.Publishing{ContentType: "application/json", Body: body, Timestamp: time.Now()})
}
