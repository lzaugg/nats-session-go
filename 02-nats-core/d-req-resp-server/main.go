package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
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

	subject := fmt.Sprintf("service.ping.%s", user)

	slog.Info("listening to requests", "subject", subject)
	slog.Info("adapt the previous example (c-req-resp-client) to use the correct subject for the request")

	// Again, the most basic way to do a request/response from the server side.
	// It's basically just a subscribe and when a reply subject is set, it will be used to send the reply.

	// For higher level use cases, check NATS micro for handling queue groups, discovery, status codes, etc automatically.
	sub, err := nc.Subscribe(subject, func(m *nats.Msg) {
		slog.Info("received message", "message", string(m.Data), "subject", m.Subject)
		if m.Reply == "" {
			slog.Warn("no reply subject. skipping reply", "subject", m.Subject)
			return
		}
		err := nc.Publish(m.Reply, []byte("pong"))
		if err != nil {
			slog.Error("error publishing to reply subject", "error", err.Error())
			return
		}
	})
	if err != nil {
		return fmt.Errorf("error subscribing to service.ping for subject %s: %w", subject, err)
	}
	defer sub.Unsubscribe()

	select {}
}
