package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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
	natsUrl := flag.String("nats-url", "nats01.mye.ch", "nats url (without user:pass)")
	sessionToken := flag.String("session-token", "", "session token for getting initial credentials")
	envFile := flag.String("env-file", ".env", "env file to save credentials to (default: .env)")

	flag.Parse()

	if *sessionToken == "" {
		return fmt.Errorf("session token is required")
	}

	var server string

	slog.Info("checking status of current setup (.env file)... ")
	err := godotenv.Load(*envFile)
	if err == nil {
		server = os.Getenv("NATS_SERVER")
	}

	if server == "" {
		slog.Info("looks like we don't have any credentials or a gopher seat for this session yet, let's change that")
		err = doSetup(*sessionToken, *natsUrl)
		if err != nil {
			return fmt.Errorf("error setting up nats connection (did you use the correct session token?): %w", err)
		}
		// give fs sync some time to save the credentials properly
		time.Sleep(1 * time.Second)
	}
	err = doCheck(*envFile)
	if err != nil {
		return fmt.Errorf("error checking nats connection: %w", err)
	}
	return nil
}

func doSetup(sessionToken, natsUrl string) error {
	// check if natsUrl already contains the schema nats://
	if strings.HasPrefix(natsUrl, "nats://") {
		// URL already has schema, add user:pass after the schema
		natsUrl = fmt.Sprintf("nats://gopher:%s@%s", sessionToken, strings.TrimPrefix(natsUrl, "nats://"))
	} else {
		// URL doesn't have schema, add it along with user:pass
		natsUrl = fmt.Sprintf("nats://gopher:%s@%s", sessionToken, natsUrl)
	}

	nc, err := nats.Connect(natsUrl) //nats.CustomInboxPrefix("_INBOX_NEW_GOPHER."),

	if err != nil {
		return fmt.Errorf("error connecting to nats: %w", err)
	}
	defer nc.Close()

	// get seat number
	slog.Info("requesting a gopher seat (credentials)...")
	nextGopherMsg, err := nc.Request("service.next-gopher", nil, 10*time.Second)
	if err != nil {
		return fmt.Errorf("error getting new gopher user name for session %s: %w", sessionToken, err)
	}

	// save nkey to file
	// find place to save .creds file. It is one level up from this file.
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	user := string(nextGopherMsg.Data)

	// save env file
	envFile := filepath.Join(dir, ".env")
	err = os.WriteFile(envFile, []byte(fmt.Sprintf("NATS_USER=%s\nNATS_SERVER=%s", user, natsUrl)), 0600)
	if err != nil {
		return fmt.Errorf("error saving env file: %w", err)
	}
	slog.Info("---")
	slog.Info("welcome aboard", "user", user)
	slog.Info("---")
	slog.Info("env file saved to file", "file", envFile)
	slog.Info("use the .env file for all further interactions with the given NATS server")

	return nil
}

func doCheck(envFile string) error {
	// read .env file
	err := godotenv.Load(envFile)
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	user := os.Getenv("NATS_USER")
	server := os.Getenv("NATS_SERVER")

	if user == "" || server == "" {
		return fmt.Errorf("NATS_USER and NATS_SERVER must be set in .env file")
	}
	nc, err := nats.Connect(server)
	if err != nil {
		return fmt.Errorf("error connecting to nats: %w", err)
	}
	defer nc.Close()

	_, err = nc.Request("service.ping", nil, 1*time.Second)
	if err != nil {
		return fmt.Errorf("error pinging nats: %w", err)
	}
	slog.Info("====")
	slog.Info("you're all set up!", "user", user)
	slog.Info("====")

	return nil
}
