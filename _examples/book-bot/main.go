// Package contains example of book search bot via Inline Mode using Open Library API.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/_examples/runner"
	"github.com/mr-linch/go-tg/tgb"
)

const pageSize = 10

var (
	inlineQueryParamRegexp = regexp.MustCompile(`(\w+):([^\s]+)`)

	// notViaBot filters out messages sent via inline bots.
	notViaBot = tgb.FilterFunc(func(ctx context.Context, update *tgb.Update) (bool, error) {
		return update.Message != nil && update.Message.ViaBot == nil, nil
	})
)

func main() {
	books := &BooksClient{Doer: http.DefaultClient}

	runner.Run(tgb.NewRouter().
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			// handles /start and /help
			return msg.Answer(
				tg.HTML.Text(
					tg.HTML.Bold("Hi, I'm a book search bot!"),
					"",
					"Use me in inline mode to search for books.",
					tg.HTML.Line("Type", tg.HTML.Code("@botname tolkien"), "in any chat."),
					tg.HTML.Line("Filter by author:", tg.HTML.Code("author:Tolkien")),
				),
			).ReplyMarkup(tg.NewInlineKeyboardMarkup(tg.NewButtonRow(
				tg.NewInlineKeyboardButtonSwitchInlineQueryCurrentChat("Try it", " "),
			))).DoVoid(ctx)
		}, tgb.Command("start", tgb.WithCommandAlias("help")), tgb.ChatType(tg.ChatTypePrivate)).
		Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
			// handles any other text in private chat
			return msg.Answer("Use me in inline mode! Click the button below.").
				ReplyMarkup(tg.NewInlineKeyboardMarkup(tg.NewButtonRow(
					tg.NewInlineKeyboardButtonSwitchInlineQueryCurrentChat("Search books", " "),
				))).DoVoid(ctx)
		}, tgb.ChatType(tg.ChatTypePrivate), notViaBot).
		InlineQuery(func(ctx context.Context, iq *tgb.InlineQueryUpdate) error {
			// search books and return results
			query := strings.TrimSpace(iq.Query)
			query, params := parseInlineQuery(query)

			offset := 0
			if iq.Offset != "" {
				var err error
				offset, err = strconv.Atoi(iq.Offset)
				if err != nil {
					return fmt.Errorf("parse offset: %w", err)
				}
			}

			var (
				results []Book
				err     error
			)

			if author, ok := params["author"]; ok {
				results, err = books.SearchByAuthor(ctx, query, author, offset, pageSize)
			} else {
				if query == "" {
					query = "bestseller"
				}
				results, err = books.Search(ctx, query, offset, pageSize)
			}

			if err != nil {
				return fmt.Errorf("search books: %w", err)
			}

			items := make([]tg.InlineQueryResultClass, len(results))
			for i, book := range results {
				items[i] = newBookResult(book)
			}

			nextOffset := ""
			if len(results) == pageSize {
				nextOffset = strconv.Itoa(offset + pageSize)
			}

			return iq.Answer(items...).
				CacheTime(0).
				NextOffset(nextOffset).
				DoVoid(ctx)
		}).
		ChosenInlineResult(func(ctx context.Context, chosen *tgb.ChosenInlineResultUpdate) error {
			// log chosen result
			log.Printf("user %d chose result %s (query: %s)",
				chosen.From.ID, chosen.ResultID, chosen.Query,
			)
			return nil
		}),

		tg.WithClientInterceptors(
			tg.NewInterceptorMethodFilter(
				tg.NewInterceptorDefaultParseMethod(tg.HTML),
				"sendMessage",
			),
		),
	)
}

func newBookResult(book Book) tg.InlineQueryResultClass {
	caption := newBookCaption(book)
	keyboard := newBookKeyboard(book)

	description := book.AuthorName
	if book.FirstPublishYear > 0 {
		description += fmt.Sprintf(" (%d)", book.FirstPublishYear)
	}

	content := tg.InputTextMessageContent{
		MessageText: caption,
		ParseMode:   tg.HTML,
	}

	// show cover as large link preview above text
	if book.CoverID > 0 {
		content.LinkPreviewOptions = &tg.LinkPreviewOptions{
			URL:              book.CoverURL("L"),
			PreferLargeMedia: true,
			ShowAboveText:    true,
		}
	}

	article := tg.NewInlineQueryResultArticle(book.Key, book.Title, content).
		WithDescription(description).
		WithReplyMarkup(keyboard)

	if book.CoverID > 0 {
		article = article.WithThumbnailURL(book.CoverURL("S"))
	}

	return article
}

func newBookCaption(book Book) string {
	lines := []string{
		tg.HTML.Link(tg.HTML.Bold(book.Title), book.OpenLibraryURL()),
		tg.HTML.Line("by", tg.HTML.Italic(book.AuthorName)),
	}

	if book.FirstPublishYear > 0 {
		lines = append(lines, fmt.Sprintf("First published: %d", book.FirstPublishYear))
	}

	if book.PageCount > 0 {
		lines = append(lines, fmt.Sprintf("Pages: %d", book.PageCount))
	}

	if len(book.Subjects) > 0 {
		lines = append(lines, tg.HTML.Line("Subjects:", tg.HTML.Italic(strings.Join(book.Subjects, ", "))))
	}

	if book.EditionCount > 0 {
		lines = append(lines, fmt.Sprintf("Editions: %d", book.EditionCount))
	}

	return tg.HTML.Text(lines...)
}

func newBookKeyboard(book Book) tg.InlineKeyboardMarkup {
	return tg.NewInlineKeyboardMarkup(
		tg.NewButtonRow(
			tg.NewInlineKeyboardButtonURL("Open Library", book.OpenLibraryURL()),
			tg.NewInlineKeyboardButtonSwitchInlineQueryCurrentChat(
				fmt.Sprintf("More by %s", book.AuthorName),
				fmt.Sprintf("author:%s ", book.AuthorName),
			),
		),
	)
}

func parseInlineQuery(v string) (query string, params map[string]string) {
	args := inlineQueryParamRegexp.FindAllStringSubmatch(v, -1)

	params = make(map[string]string, len(args))

	for _, match := range args {
		if len(match) == 3 {
			params[match[1]] = match[2]
		}
	}

	queryWithoutParams := inlineQueryParamRegexp.ReplaceAllString(v, "")

	return strings.TrimSpace(queryWithoutParams), params
}
