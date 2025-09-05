package main

import (
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
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

	// List of smile emojis
	smileEmojis := []string{
		"😀", "😃", "😄", "😁", "😆", "😅", "😂", "🤣", "😊", "😇",
		"🙂", "🙃", "😉", "😌", "😍", "🥰", "😘", "😗", "😙", "😚",
		"😋", "😛", "😝", "😜", "🤪", "🤨", "🧐", "🤓", "😎", "🤩",
		"🥳", "😏", "😒", "😞", "😔", "😟", "😕", "🙁", "☹️", "😣",
		"😖", "😫", "😩", "🥺", "😢", "😭", "😤", "😡", "😠",
	}

	// Select a random emoji
	randomEmoji := smileEmojis[rand.Intn(len(smileEmojis))]

	// This example makes only sense if someone is listening to the subject and displaying something on the screen.
	subject := fmt.Sprintf("slide.root.9.%s.html", user)
	slog.Info("publishing to " + subject)
	err = nc.Publish(subject, []byte(randomEmoji))
	if err != nil {
		return fmt.Errorf("error publishing to "+subject+": %w", err)
	}

	return nil
}
