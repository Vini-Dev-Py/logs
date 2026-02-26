package rabbit

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"logs-worker/internal/app"
	"logs-worker/internal/infra/cassandra"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Consume(url string, repo cassandra.Repo) error {
	var conn *amqp.Connection
	var err error

	// Retry connection up to 15 times, waiting 2 seconds between attempts
	for i := 0; i < 15; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("rabbit.Consume: waiting for rabbitmq... %v", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to rabbitmq after retries: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	msgs, err := ch.Consume("log_events", "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	log.Println("logs-worker consuming")
	for msg := range msgs {
		var e app.Event
		if err := json.Unmarshal(msg.Body, &e); err != nil {
			_ = msg.Nack(false, false)
			continue
		}
		if err := repo.Persist(e); err != nil {
			_ = msg.Nack(false, true)
			continue
		}
		_ = msg.Ack(false)
	}
	return nil
}
