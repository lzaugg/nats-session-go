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

	envFile := flag.String("env-file", ".env", "env file to load credentials from (default: .env)")
	flag.Parse()

	var user, server string

	err := godotenv.Load(*envFile)
	if err != nil {
		slog.Error("error loading .env file (did you run 00-setup?)", "error", err.Error())
		os.Exit(1)
	}

	user = os.Getenv("NATS_USER")
	server = os.Getenv("NATS_SERVER")

	if user == "" || server == "" {
		slog.Error("no credentials, server or user info found in .env file")
		os.Exit(1)
	}

	nc, err := nats.Connect(server)
	if err != nil {
		slog.Error("error connecting to nats", "error", err.Error())
		os.Exit(1)
	}
	defer nc.Close()

	/*
		// old way to get jetstream context (kind of deprecated)
		_, err = nc.JetStream()
		if err != nil {
			slog.Error("error getting jetstream context", "error", err.Error())
			os.Exit(1)
		}
	*/

	// new way
	js, err := jetstream.New(nc)
	if err != nil {
		slog.Error("error getting jetstream context", "error", err.Error())
		os.Exit(1)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	streamName := fmt.Sprintf("POSITION_%s", user)
	stream, err := js.CreateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{"example.*.position"},
	})
	if err != nil {
		slog.Error("error creating stream", "error", err.Error())
		os.Exit(1)
	}

	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		DeliverPolicy: jetstream.DeliverAllPolicy,
		Durable:       "position-consumer-" + user,
	})
	if err != nil {
		slog.Error("error creating consumer", "error", err.Error())
		os.Exit(1)
	}

	keysEvents, err := keyboard.GetKeys(10)
	if err != nil {
		slog.Error("error getting keys", "error", err.Error())
		os.Exit(1)
	}
	defer keyboard.Close()

	paused := false
	for {

		select {
		case key := <-keysEvents:
			if key.Key == keyboard.KeyCtrlC {
				slog.Info("ctrl+c pressed, exiting")
				os.Exit(0)
			}
			if key.Rune == 'p' || key.Rune == 'P' {
				slog.Info("handle key event", "paused", paused)
				paused = !paused
			}
		default:
			if paused {
				continue
			}
			msgs, _ := cons.Fetch(1)
			var i int
			for msg := range msgs.Messages() {
				msg.Ack()
				metadata, _ := msg.Metadata()
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
