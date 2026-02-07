package main

import (
	"context"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/_examples/runner"
	"github.com/mr-linch/go-tg/tgb"
)

var menu = struct {
	Profile  string
	Settings string
	About    string
}{
	Profile:  "👤 Profile",
	Settings: "⚙️ Settings",
	About:    "📖 About",
}

func main() {
	runner.Run(tgb.NewRouter().
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			kb := tg.NewReplyKeyboard().
				Text(menu.Profile).
				Text(menu.Settings).
				Text(menu.About).
				Adjust(1).
				Resize()

			return mu.Answer("Hey, please click a button above to see how text filter works").
				ReplyMarkup(kb).
				DoVoid(ctx)
		}, tgb.Command("start")).
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			return mu.Answer("this is profile section").DoVoid(ctx)
		}, tgb.TextEqual(menu.Profile)).
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			return mu.Answer("this is settings section").DoVoid(ctx)
		}, tgb.TextEqual(menu.Settings)).
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			return mu.Answer("this is about section").DoVoid(ctx)
		}, tgb.TextEqual(menu.About)),
	)
}
