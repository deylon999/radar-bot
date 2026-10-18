package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-telegram/bot"

	"github.com/deylon999/radar-bot/internal/backend"
	"github.com/deylon999/radar-bot/internal/config"
	"github.com/deylon999/radar-bot/internal/telegram"
)

func main() {
	if err := run(); err != nil {
		slog.Error("bot stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := backend.NewFake(time.Now)
	handlers := telegram.NewHandlers(client, cfg.SiteURL, log)

	b, err := bot.New(cfg.BotToken, bot.WithDefaultHandler(handlers.Default))
	if err != nil {
		return err
	}
	handlers.Register(b)

	log.Info("bot started")
	b.Start(ctx)
	log.Info("bot stopped")

	return nil
}
