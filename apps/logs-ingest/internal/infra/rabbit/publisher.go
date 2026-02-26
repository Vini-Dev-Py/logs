package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"logs-ingest/internal/app"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct{ ch *amqp.Channel }

func New(url string) (*Publisher, error) {
	var conn *amqp.Connection
	var err error
	
	// Retries connection up to 15 times, waiting 2 seconds between attempts
	for i := 0; i < 15; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("rabbit.New: waiting for rabbitmq... %v", err)
		time.Sleep(2 * time.Second)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq after retries: %w", err)
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
