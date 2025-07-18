package config

import (
	"log"
	"os"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	Connection *amqp091.Connection
	Channel    *amqp091.Channel
}

func InitRabbitMQ() *RabbitMQ {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	// Retry connection with backoff
	var conn *amqp091.Connection
	var err error
	for i := 0; i < 10; i++ {
		conn, err = amqp091.Dial(url)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ (attempt %d/10): %v", i+1, err)
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ after 10 attempts: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open a channel: %v", err)
	}

	_, err = ch.QueueDeclare(
		"transfer_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("failed to declare a queue: %v", err)
	}

	return &RabbitMQ{
		Connection: conn,
		Channel:    ch,
	}
}

func (r *RabbitMQ) Close() {
	if r.Channel != nil {
		r.Channel.Close()
	}

	if r.Connection != nil {
		r.Connection.Close()
	}
}
