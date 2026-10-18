package telegram

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/deylon999/radar-bot/internal/backend"
)

const helpText = `Я присылаю срочные инсайты «Радара» и сводку по расписанию.

/start — подключить уведомления
/help — эта справка`

type Handlers struct {
	backend backend.Client
	siteURL string
	log     *slog.Logger
}

func NewHandlers(client backend.Client, siteURL string, log *slog.Logger) *Handlers {
	return &Handlers{backend: client, siteURL: siteURL, log: log}
}

func (h *Handlers) Register(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeCommandStartOnly, h.start)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeCommandStartOnly, h.help)
}

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

	var username string
	if update.Message.From != nil {
		username = update.Message.From.Username
	}

	_, err := h.backend.ConfirmLink(ctx, code, chatID, username)
	if err != nil {
		h.log.Warn("confirm link", "chat_id", chatID, "err", err)
	} else {
		h.log.Info("chat linked", "chat_id", chatID)
	}
	h.reply(ctx, b, chatID, linkReply(err))
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

func startPayload(text string) string {
	_, payload, _ := strings.Cut(strings.TrimSpace(text), " ")
	return strings.TrimSpace(payload)
}

func linkReply(err error) string {
	switch {
	case err == nil:
		return "Готово. Срочные инсайты будут приходить сюда"
	case errors.Is(err, backend.ErrCodeExpired):
		return "Ссылка устарела. Нажмите «Подключить Telegram» ещё раз"
	case errors.Is(err, backend.ErrCodeUsed):
		return "Эта ссылка уже использована"
	case errors.Is(err, backend.ErrCodeNotFound):
		return "Ссылка не подходит. Откройте её заново на сайте"
	default:
		return "Не получилось подключить уведомления. Попробуйте позже"
	}
}
