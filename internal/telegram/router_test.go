package telegram

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/deylon999/radar-bot/internal/backend"
)

type fakeTelegram struct {
	mu   sync.Mutex
	sent []string
}

func (f *fakeTelegram) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/sendMessage") {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.mu.Lock()
		f.sent = append(f.sent, r.FormValue("text"))
		f.mu.Unlock()
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "result": map[string]any{}})
}

func (f *fakeTelegram) last(t *testing.T) string {
	t.Helper()
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.sent) == 0 {
		t.Fatal("bot sent nothing")
	}
	return f.sent[len(f.sent)-1]
}

func newTestBot(t *testing.T) (*bot.Bot, *fakeTelegram) {
	t.Helper()

	tg := &fakeTelegram{}
	srv := httptest.NewServer(tg)
	t.Cleanup(srv.Close)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := NewHandlers(backend.NewFake(time.Now), "https://radar.example", log)

	b, err := bot.New("test:token",
		bot.WithServerURL(srv.URL),
		bot.WithSkipGetMe(),
		bot.WithNotAsyncHandlers(),
		bot.WithDefaultHandler(h.Default),
	)
	if err != nil {
		t.Fatal(err)
	}
	h.Register(b)
	return b, tg
}

func command(text string) *models.Update {
	cmd, _, _ := strings.Cut(text, " ")
	return &models.Update{Message: &models.Message{
		Chat:     models.Chat{ID: 42},
		From:     &models.User{ID: 42, Username: "demo_director"},
		Text:     text,
		Entities: []models.MessageEntity{{Type: models.MessageEntityTypeBotCommand, Offset: 0, Length: len(cmd)}},
	}}
}

func TestCommandsReachTheirHandlers(t *testing.T) {
	b, tg := newTestBot(t)
	ctx := context.Background()

	steps := []struct {
		text string
		want string
	}{
		{"/help", "Я присылаю срочные инсайты"},
		{"/status", "Чтобы получать уведомления"},
		{"/start valid-code", "Готово."},
		{"/status", "Уведомления подключены: АО «Полевея Агро»"},
		{"/stop", "Уведомления отключены."},
		{"/start valid-code", "Эта ссылка уже использована"},
		{"/unknown", "Не понял команду."},
	}
	for _, s := range steps {
		b.ProcessUpdate(ctx, command(s.text))
		if got := tg.last(t); !strings.HasPrefix(got, s.want) {
			t.Errorf("%s: got %q, want prefix %q", s.text, got, s.want)
		}
	}
}
