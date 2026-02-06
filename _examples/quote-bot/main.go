// Package contains example of using tgb.ChatType filter.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/_examples/runner"
	"github.com/mr-linch/go-tg/tgb"
)

func main() {
	quotesClient := &QuotesClient{
		Doer: http.DefaultClient,
	}

	r := tgb.NewRouter()

	r.Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
		return msg.Answer("Hey! I'm inline bot, so click button below for start 👇").
			ReplyMarkup(tg.NewInlineKeyboardMarkup(tg.NewButtonRow(
				tg.NewInlineKeyboardButtonSwitchInlineQueryCurrentChat("Start", " "),
			))).
			DoVoid(ctx)
	}, tgb.ChatType(tg.ChatTypePrivate), tgb.Command("start"))

	r.InlineQuery(func(ctx context.Context, iq *tgb.InlineQueryUpdate) error {
		query := strings.TrimSpace(iq.Query)

		language := "en"
		if iq.From.LanguageCode == "ru" {
			language = iq.From.LanguageCode
		}

		query, params := parseInlineQuery(query)

		var (
			quotes []Quote
			err    error
		)

		if author, ok := params["author"]; ok {
			quotes, err = quotesClient.ListByAuthor(ctx, language, query, author, 0, 10)
			if err != nil {
				return fmt.Errorf("list quotes: %w", err)
			}
		} else {
			quotes, err = quotesClient.List(ctx, language, query, 0, 10)
			if err != nil {
				return fmt.Errorf("list quotes: %w", err)
			}
		}

		result := make([]tg.InlineQueryResult, len(quotes))

		for i, quote := range quotes {

			messageText := tg.HTML.Text(
				tg.HTML.Italic("„"+quote.Text+"“"),
				"",
				tg.HTML.Line("by", tg.HTML.Bold(quote.Author.Name)),
			)

			article := tg.NewInlineQueryResultArticle(
				quote.ID,
				quote.Author.Name,
				tg.InputTextMessageContent{
					MessageText: messageText,
					ParseMode:   tg.HTML,
				},
			)
			article.Article.Description = quoteText(quote.Text)
			article.Article.ReplyMarkup = tg.NewInlineKeyboardMarkup(
				tg.NewButtonRow(
					tg.NewInlineKeyboardButtonSwitchInlineQueryCurrentChat(
						fmt.Sprintf("More by %s", quote.Author.Name),
						fmt.Sprintf("author:%s ", quote.Author.ID),
					),
				),
			).Ptr()
			result[i] = article
		}

		return iq.Answer(result).CacheTime(0).DoVoid(ctx)
	})

	runner.Run(r)
}

func quoteText(v string) string {
	return "„" + v + "“"
}

type QuotesClient struct {
	Doer *http.Client
}

func (c *QuotesClient) getURL(lang string) string {
	return fmt.Sprintf("https://api.fisenko.net/v1/quotes/%s", lang)
}

func (c *QuotesClient) getAuthorURL(lang string, authorID string) string {
	return fmt.Sprintf("https://api.fisenko.net/v1/authors/%s/%s/quotes", lang, authorID)
}

type Quote struct {
	ID     string      `json:"id"`
	Text   string      `json:"text"`
	Author QuoteAuthor `json:"author"`
}

type QuoteAuthor struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var inlineQueryParamRegexp = regexp.MustCompile(`(\w+):[\s]?(\w+)`)

func parseInlineQuery(v string) (string, map[string]string) {
	args := inlineQueryParamRegexp.FindAllStringSubmatch(v, -1)

	params := make(map[string]string, len(args))

	for _, match := range args {
		if len(match) == 3 {
			params[match[1]] = match[2]
		}
	}

	queryWithoutParams := inlineQueryParamRegexp.ReplaceAllString(v, "")

	return strings.TrimSpace(queryWithoutParams), params
}

func (c *QuotesClient) ListByAuthor(ctx context.Context, language, query, author string, offset, limit int) ([]Quote, error) {
	u, err := url.Parse(c.getAuthorURL(language, author))
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	q := u.Query()

	if query != "" {
		q.Set("query", query)
	}

	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}

	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}

	u.RawQuery = q.Encode()

	log.Printf("GET %s", u.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	res, err := c.Doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	var result []Quote

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}

func (c *QuotesClient) List(ctx context.Context, language, query string, offset, limit int) ([]Quote, error) {
	u, err := url.Parse(c.getURL(language))
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	q := u.Query()

	if query != "" {
		q.Set("query", query)
	}

	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}

	if offset > 0 {
		q.Set("offset", strconv.Itoa(offset))
	}

	u.RawQuery = q.Encode()

	log.Printf("GET %s", u.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	res, err := c.Doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer res.Body.Close()

	var result []Quote

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}
