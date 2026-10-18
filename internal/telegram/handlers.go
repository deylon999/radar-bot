package telegram

import (
	"context"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

const helpText = `Я присылаю срочные инсайты «Радара» и сводку по расписанию.

/start — подключить уведомления
/help — эта справка`

type Handlers struct {
	siteURL string
	log     *slog.Logger
}

func NewHandlers(siteURL string, log *slog.Logger) *Handlers {
	return &Handlers{siteURL: siteURL, log: log}
}

func (h *Handlers) Register(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeCommandStartOnly, h.start)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeCommandStartOnly, h.help)
}

// Default answers everything the bot doesn't recognise.
func (h *Handlers) Default(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}
	h.reply(ctx, b, update.Message.Chat.ID, "Не понял команду. Список команд — /help")
}

func (h *Handlers) start(ctx context.Context, b *bot.Bot, update *models.Update) {
	chatID := update.Message.Chat.ID

	code := startPayload(update.Message.Text)
	if code == "" {
		h.reply(ctx, b, chatID, "Чтобы получать уведомления, нажмите «Подключить Telegram» на сайте: "+
			h.siteURL+"/notifications")
		return
	}

	// TODO: confirm the code with the backend and bind the chat.
	h.log.Info("link requested", "chat_id", chatID, "code", code)
	h.reply(ctx, b, chatID, "Привязка скоро заработает.")
}

func (h *Handlers) help(ctx context.Context, b *bot.Bot, update *models.Update) {
	h.reply(ctx, b, update.Message.Chat.ID, helpText)
}

func (h *Handlers) reply(ctx context.Context, b *bot.Bot, chatID int64, text string) {
	_, err := b.SendMessage(ctx, &bot.SendMessageParams{ChatID: chatID, Text: text})
	if err != nil {
		h.log.Error("send message", "chat_id", chatID, "err", err)
	}
}

// startPayload returns the deep-link parameter from "/start <payload>".
func startPayload(text string) string {
	_, payload, _ := strings.Cut(strings.TrimSpace(text), " ")
	return strings.TrimSpace(payload)
}
