package telegram

import (
	"context"
	"errors"
	"fmt"
	"html"
	"net/url"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/deylon999/radar-bot/internal/backend"
	"github.com/deylon999/radar-bot/internal/delivery"
)

var _ delivery.Sender = (*Sender)(nil)

type Sender struct {
	bot     *bot.Bot
	siteURL string
	buttons bool
}

func NewSender(b *bot.Bot, siteURL string) *Sender {
	return &Sender{bot: b, siteURL: siteURL, buttons: isPublicURL(siteURL)}
}

func (s *Sender) Send(ctx context.Context, d backend.Delivery) error {
	msg, err := s.compose(d)
	if err != nil {
		return err
	}

	params := &bot.SendMessageParams{
		ChatID:             d.ChatID,
		Text:               msg.text,
		ParseMode:          models.ParseModeHTML,
		LinkPreviewOptions: &models.LinkPreviewOptions{IsDisabled: bot.True()},
	}
	if s.buttons {
		params.ReplyMarkup = &models.InlineKeyboardMarkup{
			InlineKeyboard: [][]models.InlineKeyboardButton{{{Text: msg.linkLabel, URL: msg.linkURL}}},
		}
	} else {
		params.Text += "\n\n" + html.EscapeString(msg.linkURL)
	}

	_, err = s.bot.SendMessage(ctx, params)
	if errors.Is(err, bot.ErrorForbidden) {
		return fmt.Errorf("%w: %v", delivery.ErrBlocked, err)
	}
	return err
}

func (s *Sender) compose(d backend.Delivery) (message, error) {
	switch {
	case d.Kind == backend.DeliveryUrgent && d.Insight != nil:
		return urgentMessage(s.siteURL, *d.Insight), nil
	case d.Kind == backend.DeliveryDigest && d.Digest != nil:
		return digestMessage(s.siteURL, *d.Digest), nil
	default:
		return message{}, fmt.Errorf("delivery %s: no payload for kind %q", d.ID, d.Kind)
	}
}

func isPublicURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	switch u.Hostname() {
	case "", "localhost", "127.0.0.1", "::1":
		return false
	}
	return true
}
