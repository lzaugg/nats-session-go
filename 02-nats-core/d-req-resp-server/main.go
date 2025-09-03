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

	subject := fmt.Sprintf("service.ping.%s", user)

	slog.Info("listening to requests on " + subject)
	sub, err := nc.Subscribe(subject, func(m *nats.Msg) {
		slog.Info("received message", "message", string(m.Data), "subject", m.Subject)
		if m.Reply == "" {
			slog.Warn("no reply subject. skipping reply", "subject", m.Subject)
			return
		}
		nc.Publish(m.Reply, []byte("pong"))
	})
	if err != nil {
		slog.Error("error subscribing to service.ping for subject "+subject, "error", err.Error())
		os.Exit(1)
	}
	defer sub.Unsubscribe()

	select {}
}
