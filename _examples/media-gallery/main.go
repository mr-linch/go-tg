// Package contains a photo gallery bot that demonstrates MediaMessageCallBuilder.
// It shows how to build a caption once and reuse it for both sending and editing.
package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/_examples/runner"
	"github.com/mr-linch/go-tg/tgb"
)

type galleryItem struct {
	Title       string
	Description string
	PhotoURL    string
}

var gallery = []galleryItem{
	{
		Title:       "Mountain Lake",
		Description: "A serene mountain lake surrounded by pine trees.",
		PhotoURL:    "https://picsum.photos/id/15/400/300",
	},
	{
		Title:       "City Skyline",
		Description: "Modern city skyline at golden hour.",
		PhotoURL:    "https://picsum.photos/id/274/400/300",
	},
	{
		Title:       "Forest Path",
		Description: "A winding path through an autumn forest.",
		PhotoURL:    "https://picsum.photos/id/167/400/300",
	},
	{
		Title:       "Ocean Waves",
		Description: "Waves crashing on a rocky coastline.",
		PhotoURL:    "https://picsum.photos/id/180/400/300",
	},
}

type galleryNav struct {
	Index int
}

var galleryNavFilter = tgb.NewCallbackDataFilter[galleryNav]("gallery")

func newGalleryMessage(index int) *tgb.MediaMessageCallBuilder {
	item := gallery[index]
	pm := tg.HTML

	caption := pm.Text(
		pm.Bold(item.Title),
		"",
		pm.Escape(item.Description),
		"",
		pm.Italic(fmt.Sprintf("%d / %d", index+1, len(gallery))),
	)

	prev := (index - 1 + len(gallery)) % len(gallery)
	next := (index + 1) % len(gallery)

	return tgb.NewMediaMessageCallBuilder(caption).
		ParseMode(pm).
		ShowCaptionAboveMedia(true).
		ReplyMarkup(tg.NewInlineKeyboard().
			Button(
				galleryNavFilter.MustButton("< Prev", galleryNav{Index: prev}),
				galleryNavFilter.MustButton(strconv.Itoa(index+1)+"/"+strconv.Itoa(len(gallery)), galleryNav{Index: index}),
				galleryNavFilter.MustButton("Next >", galleryNav{Index: next}),
			).Markup())
}

func main() {
	runner.Run(tgb.NewRouter().
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			b := newGalleryMessage(0)
			return msg.Update.Reply(ctx, b.AsSendPhoto(msg.Chat, tg.NewFileArgURL(gallery[0].PhotoURL)))
		}, tgb.Command("start", tgb.WithCommandAlias("gallery"))).
		CallbackQuery(galleryNavFilter.Handler(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate, nav galleryNav) error {
			_ = cbq.Answer().DoVoid(ctx)
			b := newGalleryMessage(nav.Index)
			photo := b.NewInputMediaPhoto(tg.NewFileArgURL(gallery[nav.Index].PhotoURL))
			return cbq.Update.Reply(ctx, b.AsEditMediaFromCBQ(cbq.CallbackQuery, photo))
		}), galleryNavFilter.Filter()),
	)
}
