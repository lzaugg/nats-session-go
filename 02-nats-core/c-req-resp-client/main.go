package main

import (
	"flag"
	"log/slog"
	"os"
	"time"

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

	// That's the most basic way to do a request/response from the client side.
	// It's basically just a publish with a reply subject set.
	msg, err := nc.Request("service.ping", nil, 10*time.Second)
	if err != nil {
		slog.Error("error requesting service.ping", "error", err.Error())
		os.Exit(1)
	}

	slog.Info("received message", "message", string(msg.Data))
}
