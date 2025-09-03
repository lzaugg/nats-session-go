package main

import (
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {

	var user, server, creds string

	err := godotenv.Load()
	if err != nil {
		slog.Error("error loading .env file (did you run 00-setup?)", "error", err.Error())
		os.Exit(1)
	}

	user = os.Getenv("NATS_USER")
	server = os.Getenv("NATS_SERVER")
	creds = os.Getenv("NATS_CREDS_FILE")

	if user == "" || server == "" || creds == "" {
		slog.Error("no credentials, server or user info found in .env file")
		os.Exit(1)
	}

	nc, err := nats.Connect(server, nats.UserCredentials(creds))
	if err != nil {
		slog.Error("error connecting to nats", "error", err.Error())
		os.Exit(1)
	}
	defer nc.Close()

	nc.Subscribe("example.vehicle-0.position", func(m *nats.Msg) {
		slog.Info("received message", "message", string(m.Data))
	})

	select {}

}
