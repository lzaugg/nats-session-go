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

	subject := "example.*.position"
	slog.Info("subscribing to " + subject)
	nc.Subscribe(subject, func(m *nats.Msg) {
		fmt.Println("received message", "subject", m.Subject, "data", string(m.Data))
	})

	select {}

}
