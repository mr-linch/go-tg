// Package contains example of using tgb.ChatType filter.
package main

import (
	"context"
	"embed"
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
)

var (
	flagToken   string
	flagBaseURL string
	flagListen  string

	//go:embed site
	webAppFS embed.FS
)

func main() {
	flag.StringVar(&flagToken, "token", "", "Telegram Bot API token")
	flag.StringVar(&flagBaseURL, "base-url", "", "Base URL for incoming http requests")
	flag.StringVar(&flagListen, "listen", ":8080", "Listen address")
	flag.Parse()

	if flagToken == "" {
		log.Fatal("-token is required")
	}

	if flagBaseURL == "" {
		log.Fatal("-base-url is required")
	}

	flagBaseURL = strings.TrimSuffix(flagBaseURL, "/")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	if flagToken == "" {
		return fmt.Errorf("token is required")
	}

	client := tg.New(flagToken)

	me, err := client.Me(ctx)
	if err != nil {
		return fmt.Errorf("get me: %w", err)
	}
	log.Printf("auth as https://t.me/%s", me.Username)

	if err := client.SetChatMenuButton().MenuButton(tg.NewMenuButtonWebApp(tg.MenuButtonWebApp{
		Text: "Open Web App",
		WebApp: tg.WebAppInfo{
			URL: flagBaseURL + "/webapp",
		},
	})).DoVoid(ctx); err != nil {
		return fmt.Errorf("set menu button: %w", err)
	}

	menuButton, err := client.GetChatMenuButton().Do(ctx)
	if err != nil {
		return fmt.Errorf("get menu button: %w", err)
	}

	log.Printf("webapp menu button: %#v", menuButton.WebApp)

	router := newRouter(flagBaseURL)

	webhook := tgb.NewWebhook(router, client, flagBaseURL+"/webhook",
		tgb.WithWebhookLogger(log.Default()),
	)

	if err := webhook.Setup(ctx); err != nil {
		return fmt.Errorf("setup webhook: %w", err)
	}

	mux := http.NewServeMux()

	mux.Handle("/webhook", webhook)

	mux.Handle("/login-url", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authWidget, err := tg.ParseAuthWidgetQuery(r.URL.Query())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if authWidget.Valid(flagToken) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "✅ You are authorized as Telegram User #%d\n", authWidget.ID)
			fmt.Fprintf(w, "🧪 You can change some URL parameters to see how signature checking works.")
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "🛑 You are not authorized, because of invalid signature")
		}
	}))

	stripped, err := fs.Sub(webAppFS, "site")
	if err != nil {
		return fmt.Errorf("sub static: %w", err)
	}

	mux.Handle("/webapp/check", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := tg.ParseWebAppInitData(r.URL.Query())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if data.Valid(flagToken) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			encoder := json.NewEncoder(w)
			encoder.SetIndent("", " ")
			encoder.Encode(data)
		}

	}))

	mux.Handle("/webapp/", http.StripPrefix(
		"/webapp/",
		http.FileServer(http.FS(stripped)),
	))

	return runServer(ctx, mux, flagListen)
}

func newRouter(baseURL string) *tgb.Router {
	return tgb.NewRouter().
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			err := msg.Answer("hey, this is buttons demo").ReplyMarkup(tg.NewInlineKeyboardMarkup(
				tg.NewButtonColumn(
					tg.NewInlineKeyboardButtonLoginURL("Login URL", tg.LoginURL{
						URL: baseURL + "/login-url",
					}),
					tg.NewInlineKeyboardButtonWebApp("Web App", tg.WebAppInfo{
						URL: baseURL + "/webapp",
					}),
				)...,
			)).DoVoid(ctx)

			var tgErr *tg.Error
			if errors.As(err, &tgErr) && tgErr.Contains("bot_domain_invalid") {
				return msg.Answer(tg.HTML.Text(
					"⚠️ Bot is not configured properly. Follow the instruction:",
					"",
					"1. Go to @BotFather",
					"2. Send /setdomain",
					"3. Choise your bot",
					"4. Enter your URL "+baseURL,
				)).DoVoid(ctx)
			}

			return err

		}, tgb.Command("start"))
}

func runServer(ctx context.Context, handler http.Handler, listen string) error {
	server := &http.Server{
		Addr:    flagListen,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown: %v", err)
		}
	}()

	log.Printf("listening on %s", flagListen)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("listen and serve: %w", err)
	}

	return nil
}
