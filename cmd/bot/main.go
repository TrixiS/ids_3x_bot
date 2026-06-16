package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/TrixiS/goram"
	"github.com/TrixiS/goram/flood"
	"github.com/TrixiS/ids_3x_bot/internal/config"
	"github.com/TrixiS/ids_3x_bot/internal/phrases"
)

type Config struct {
	BotToken      string `env:"BOT_TOKEN,required"`
	Secret        string `env:"SECRET,required"`
	WebhookURL    string `env:"WEBHOOK_URL,required"`
	ListenAddress string `env:"LISTEN_ADDRESS,required"`
}

func main() {
	cfg := config.Load(&Config{})

	bot := goram.NewBot(goram.BotOptions{
		Token: cfg.BotToken,
		FloodHandler: flood.NewCondHandler(
			func(ctx context.Context, method string, request any, duration time.Duration) {
				slog.Warn("waiting for flood", "method", method, "dur", duration)
			},
		),
	})

	err := bot.SetWebhookVoid(context.Background(), &goram.SetWebhookRequest{
		URL:            cfg.WebhookURL,
		SecretToken:    cfg.Secret,
		AllowedUpdates: []goram.UpdateType{goram.UpdateMessage},
	})

	if err != nil {
		panic(err)
	}

	http.HandleFunc("/updates", func(w http.ResponseWriter, r *http.Request) {
		const secretHeaderKey = "X-Telegram-Bot-Api-Secret-Token"

		if cfg.Secret != r.Header.Get(secretHeaderKey) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		update := goram.Update{}
		err := json.NewDecoder(r.Body).Decode(&update)
		r.Body.Close()

		if err != nil {
			slog.Error("decode update", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		go func() {
			err := handleUpdate(context.Background(), bot, &update)

			if err != nil {
				slog.Error("handle update", "err", err)
			}
		}()

		w.WriteHeader(http.StatusOK)
	})

	slog.Info("listen", "address", cfg.ListenAddress)

	if err := http.ListenAndServe(cfg.ListenAddress, nil); err != nil {
		panic(err)
	}
}

func handleUpdate(ctx context.Context, bot *goram.Bot, update *goram.Update) error {
	if update.Message == nil {
		return nil
	}

	message := update.Message

	if message.From == nil {
		return nil
	}

	if message.Text == "/start" {
		return bot.SendMessageVoid(ctx, &goram.SendMessageRequest{
			ChatID: message.ChatID(),
			Text: fmt.Sprintf(
				phrases.StartMessageTextFmt,
				message.From.ID,
				message.From.Username,
			),
			ReplyMarkup: phrases.StartMarkup,
		})
	}

	if message.ChatShared != nil {
		return bot.SendMessageVoid(ctx, &goram.SendMessageRequest{
			ChatID:    message.ChatID(),
			Text:      fmt.Sprintf(phrases.IDFmt, message.ChatShared.ChatID),
			ParseMode: goram.ParseModeHTML,
		})
	}

	if message.UsersShared != nil && len(message.UsersShared.Users) > 0 {
		return bot.SendMessageVoid(ctx, &goram.SendMessageRequest{
			ChatID:    message.ChatID(),
			Text:      fmt.Sprintf(phrases.IDFmt, message.UsersShared.Users[0].UserID),
			ParseMode: goram.ParseModeHTML,
		})
	}

	if message.ForwardOrigin != nil {
		origin := message.ForwardOrigin
		builder := strings.Builder{}

		if origin.Chat != nil {
			fmt.Fprintf(&builder, phrases.ChatIDRowFmt, origin.Chat.ID)
		} else if origin.SenderChat != nil {
			fmt.Fprintf(&builder, phrases.ChatIDRowFmt, origin.SenderChat.ID)
		}

		if origin.SenderUser != nil {
			fmt.Fprintf(&builder, phrases.UserIDRowFmt, origin.SenderUser.ID)
		}

		return bot.SendMessageVoid(ctx, &goram.SendMessageRequest{
			ChatID:    message.ChatID(),
			Text:      builder.String(),
			ParseMode: goram.ParseModeHTML,
		})
	}

	return nil
}
