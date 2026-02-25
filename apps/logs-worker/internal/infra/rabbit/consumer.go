package rabbit

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"logs-worker/internal/app"
	"logs-worker/internal/infra/cassandra"
)

func Consume(url string, repo cassandra.Repo) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
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
