package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/eiannone/keyboard"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	err := run()
	if err != nil {
		slog.Error("error running", "error", err.Error())
		os.Exit(1)
	}
}

func run() error {
	envFile := flag.String("env-file", ".env", "env file to load credentials from (default: .env)")
	flag.Parse()

	var user, server string

	err := godotenv.Load(*envFile)
	if err != nil {
		return fmt.Errorf("error loading .env file (did you run 00-setup?): %w", err)
	}

	user = os.Getenv("NATS_USER")
	server = os.Getenv("NATS_SERVER")

	if user == "" || server == "" {
		return fmt.Errorf("no credentials, server or user info found in .env file")
	}

	nc, err := nats.Connect(server)
	if err != nil {
		return fmt.Errorf("error connecting to nats: %w", err)
	}
	defer nc.Close()

	// old way to get jetstream context (kind of deprecated)
	/*
		_, err = nc.JetStream()
		if err != nil {
			return fmt.Errorf("error getting jetstream context: %w", err)
		}
	*/

	// new way to get jetstream context
	js, err := jetstream.New(nc)
	if err != nil {
		return fmt.Errorf("error getting jetstream context: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	streamName := fmt.Sprintf("POSITION_%s", user)

	// Creating a stream with all position subjects.
	// There are a lot of options to configure the stream, like acks/guarantees, storage policy, etc.
	stream, err := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{"example.*.position"},
	})
	if err != nil {
		return fmt.Errorf("error creating stream: %w", err)
	}

	// Create a durable consumer for the stream. All the consumer preferences
	// are stored on the consumer context (server side).
	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		DeliverPolicy: jetstream.DeliverAllPolicy,
		Durable:       "position-consumer-" + user,
	})
	if err != nil {
		return fmt.Errorf("error creating consumer: %w", err)
	}

	// Handle keyboard events for pausing and exiting.
	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		return fmt.Errorf("error getting keys: %w", err)
	}
	defer keyboard.Close()

	paused := false
	for {

		select {
		case key := <-keysEvents:
			if key.Key == keyboard.KeyCtrlC {
				slog.Info("ctrl+c pressed, exiting")
				return nil
			}
			if key.Rune == 'p' || key.Rune == 'P' {
				slog.Info("handle key event", "paused", paused)
				paused = !paused
			}
		default:
			if paused {
				continue
			}
			// Fetch 1 message from the consumer.
			msgs, _ := cons.Fetch(1)
			var i int
			for msg := range msgs.Messages() {
				msg.Ack()
				metadata, err := msg.Metadata()
				if err != nil {
					slog.Error("error getting metadata", "error", err.Error())
					continue
				}
				fmt.Println("received message", "subject", msg.Subject(), "seq", metadata.Sequence, "data", string(msg.Data()))
				i++
			}
		}
	}

}
