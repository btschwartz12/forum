package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	flags "github.com/jessevdk/go-flags"
	"go.uber.org/zap"

	"github.com/btschwartz12/forum/server"
)

type arguments struct {
	Port            int    `short:"p" long:"port" description:"Port to listen on" default:"8000"`
	VarDir          string `short:"v" long:"var-dir" env:"FORUM_VAR_DIR" description:"Directory to store data"`
	Public          bool   `long:"public" description:"Allow anyone to create an account"`
	AuthToken       string `long:"auth-token" env:"FORUM_AUTH_TOKEN" description:"Token to use for API authentication"`
	SessionKey      string `long:"session-key" env:"FORUM_SESSION_KEY" description:"Session key"`
	SlackWebhookUrl string `long:"slack-webhook-url" env:"FORUM_SLACK_WEBHOOK_URL" description:"Slack webhook URL"`
}

var args arguments

func main() {
	_, err := flags.Parse(&args)
	if err != nil {
		panic(fmt.Errorf("error parsing flags: %s", err))
	}

	if args.VarDir == "" {
		panic("var dir is required")
	}

	if args.AuthToken == "" {
		panic("auth token is required")
	}

	if args.SessionKey == "" {
		panic("session key is required")
	}

	var l *zap.Logger
	l, _ = zap.NewProduction()
	logger := l.Sugar()

	s := &server.Server{}
	err = s.Init(logger, args.VarDir, args.Public, args.AuthToken, args.SessionKey, args.SlackWebhookUrl)
	if err != nil {
		logger.Fatalw("Error initializing server", "error", err)
	}

	r := chi.NewRouter()
	r.Mount("/", s.Router())

	errChan := make(chan error)
	go func() {
		logger.Infow("Starting server", "port", args.Port)
		errChan <- http.ListenAndServe(fmt.Sprintf(":%d", args.Port), r)
	}()
	err = <-errChan
	logger.Fatalw("http server failed", "error", err)
}
