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

	subject := "example.*.position"
	slog.Info("subscribing to " + subject)
	// If you want to use a channel, check ChanSubscribe, for queue groups use QueueSubscribe, etc...
	sub, err := nc.Subscribe(subject, func(m *nats.Msg) {
		fmt.Println("received message", "subject", m.Subject, "data", string(m.Data))
	})
	if err != nil {
		return fmt.Errorf("error subscribing to "+subject+": %w", err)
	}
	defer sub.Unsubscribe()

	select {}

}
