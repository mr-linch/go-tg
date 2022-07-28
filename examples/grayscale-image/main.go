package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/jpeg"
	"io"
	"strings"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/examples"
	"github.com/mr-linch/go-tg/tgb"
	"golang.org/x/exp/slices"
)

func main() {
	examples.Run(tgb.NewRouter().
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			// handle /start command
			return mu.Answer("Hey, send me a photo for flip it!").DoVoid(ctx)
		}, tgb.Command("start")).
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			// handle photo
			sizes := append([]tg.PhotoSize{}, mu.Message.Photo...)

			// find the biggest photo
			slices.SortFunc(sizes, func(a, b tg.PhotoSize) bool {
				return a.Width*a.Height > b.Width*b.Height
			})

			photo := sizes[0]

			if err := mu.Update.Respond(ctx,
				tg.NewSendChatActionCall(mu.Message.Chat, tg.ChatActionUploadPhoto),
			); err != nil {
				return fmt.Errorf("send chat action: %w", err)
			}

			// download photo
			fileInfo, err := mu.Client.GetFile(photo.FileID).Do(ctx)
			if err != nil {
				return fmt.Errorf("get file: %w", err)
			}

			file, err := mu.Client.Download(ctx, fileInfo.FilePath)
			if err != nil {
				return fmt.Errorf("download file: %w", err)
			}
			defer file.Close()

			// convert to grayscale
			grayscaledImage, err := grayscaleImage(file)
			if err != nil {
				return fmt.Errorf("process image: %w", err)
			}

			return mu.AnswerPhoto(tg.NewFileArgUpload(
				tg.NewInputFile("image.jpg", grayscaledImage),
			)).DoVoid(ctx)
		}, tgb.MessageType(tg.MessageTypePhoto)).
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			// handle send as document
			return mu.Answer("Please, send me photo as image, not as document").DoVoid(ctx)
		}, isDocumentPhoto).
		Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
			// handle other messages
			return mu.Answer("Send me a photo").DoVoid(ctx)
		}),
	)
}

func grayscaleImage(in io.Reader) (io.Reader, error) {
	img, _, err := image.Decode(in)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	bounds := img.Bounds()

	out := image.NewGray(bounds)

	for x := 0; x < bounds.Max.X; x++ {
		for y := 0; y < bounds.Max.Y; y++ {
			rgba := img.At(x, y)
			out.Set(x, y, rgba)
		}
	}

	buf := &bytes.Buffer{}

	if err := jpeg.Encode(buf, out, nil); err != nil {
		return nil, fmt.Errorf("encode image: %w", err)
	}

	return buf, nil
}

// isDocumnentPhoto it's composite filter.
// It's used to filter messages with type document and photo.
var isDocumentPhoto = tgb.All(
	tgb.MessageType(tg.MessageTypeDocument),
	tgb.FilterFunc(func(ctx context.Context, update *tgb.Update) (bool, error) {
		return strings.HasPrefix(update.Message.Document.MIMEType, "image/"), nil
	}),
)
