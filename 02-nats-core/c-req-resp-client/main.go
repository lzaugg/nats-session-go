package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

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

	// That's the most basic way to do a request/response from the client side.
	// It's basically just a publish with a reply subject set.
	msg, err := nc.Request("service.ping", nil, 10*time.Second)
	if err != nil {
		return fmt.Errorf("error requesting service.ping: %w", err)
	}

	slog.Info("received message", "message", string(msg.Data))

	return nil
}
