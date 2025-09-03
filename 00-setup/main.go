package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
)

func main() {
	natsUrl := flag.String("nats-url", "nats01.mye.ch", "nats url")
	sessionToken := flag.String("session-token", "", "session token for getting initial credentials")

	flag.Parse()

	var user, server, creds, password string

	slog.Info("checking status of current setup (.env file)... ")
	err := godotenv.Load()
	if err == nil {
		user = os.Getenv("NATS_USER")
		server = os.Getenv("NATS_SERVER")
		creds = os.Getenv("NATS_CREDS_FILE")
	}

	if user == "" || server == "" {
		slog.Info("looks like we don't have any credentials or a gopher seat for this session yet, let's change that")
		err = doSetup(*sessionToken, *natsUrl)
		if err != nil {
			slog.Error("error setting up nats connection (did you use the correct session token?)", "error", err.Error())
			os.Exit(1)
		}
		// give fs sync some time to save the credentials properly
		time.Sleep(1 * time.Second)
	}
	err = doCheck()
	if err != nil {
		slog.Error("error checking nats connection", "error", err.Error())
		os.Exit(1)
	}
}

func doSetup(sessionToken, natsUrl string) error {
	nc, err := nats.Connect(natsUrl,
		//nats.Token(sessionToken),
		nats.Pass
		nats.CustomInboxPrefix("_INBOX_NEW_GOPHER."),
	)
	if err != nil {
		return fmt.Errorf("error connecting to nats: %w", err)
	}
	defer nc.Close()

	// get seat number
	slog.Info("requesting a gopher seat (credentials)...")
	nkeyCredsMsg, err := nc.Request("service.new-gopher", nil, 10*time.Second)
	if err != nil {
		return fmt.Errorf("error getting new gopher credentials (nkey) for session-token %s: %w", sessionToken, err)
	}

	// save nkey to file
	// find place to save .creds file. It is one level up from this file.
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current directory: %w", err)
	}

	user := nkeyCredsMsg.Header.Get("user")
	if user == "" {
		return fmt.Errorf("no user header found in nkey credentials message")
	}

	// save nkey to file
	credsFile := filepath.Join(dir, ".creds")
	err = os.WriteFile(credsFile, nkeyCredsMsg.Data, 0600)
	if err != nil {
		return fmt.Errorf("error saving nkey to file: %w", err)
	}

	// save env file
	envFile := filepath.Join(dir, ".env")
	err = os.WriteFile(envFile, []byte(fmt.Sprintf("NATS_USER=%s\nNATS_SERVER=%s\nNATS_CREDS_FILE=%s", user, natsUrl, credsFile)), 0600)
	if err != nil {
		return fmt.Errorf("error saving env file: %w", err)
	}
	slog.Info("---")
	slog.Info("welcome aboard", "user", user)
	slog.Info("---")
	slog.Info("credentials saved to file", "file", credsFile)
	slog.Info("env file saved to file", "file", envFile)
	slog.Info("use the credentials for all further interactions with the given NATS server")

	return nil
}

func doCheck() error {
	// read .env file
	err := godotenv.Load()
	if err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	user := os.Getenv("NATS_USER")
	server := os.Getenv("NATS_SERVER")
	creds := os.Getenv("NATS_CREDS_FILE")

	if user == "" || server == "" || creds == "" {
		return fmt.Errorf("NATS_USER, NATS_SERVER, and NATS_CREDS_FILE must be set in .env file")
	}
	nc, err := nats.Connect(server, nats.UserCredentials(creds))
	if err != nil {
		return fmt.Errorf("error connecting to nats: %w", err)
	}
	defer nc.Close()

	_, err = nc.Request("service.ping", nil, 1*time.Second)
	if err != nil {
		return fmt.Errorf("error pinging nats: %w", err)
	}
	slog.Info("====")
	slog.Info("setup finished: everything looks good")
	slog.Info("====")

	return nil
}
