# go-tg

[![Go Reference](https://pkg.go.dev/badge/github.com/mr-linch/go-tg.svg)](https://pkg.go.dev/github.com/mr-linch/go-tg)
[![go.mod](https://img.shields.io/github/go-mod/go-version/mr-linch/go-tg)](go.mod)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/mr-linch/go-tg?label=latest%20release)](https://github.com/mr-linch/go-tg/releases/latest)
<!-- auto-generated: Telegram Bot API badge -->
[![Telegram Bot API](https://img.shields.io/badge/Telegram%20Bot%20API-9.4%20%28from%20February%209%2C%202026%29-blue?logo=telegram)](https://core.telegram.org/bots/api#february-9-2026)
<!-- end: auto-generated -->
[![CI](https://github.com/mr-linch/go-tg/actions/workflows/ci.yml/badge.svg)](https://github.com/mr-linch/go-tg/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/mr-linch/go-tg/branch/main/graph/badge.svg?token=9EI5CEIYXL)](https://codecov.io/gh/mr-linch/go-tg)
[![Go Report Card](https://goreportcard.com/badge/github.com/mr-linch/go-tg)](https://goreportcard.com/report/github.com/mr-linch/go-tg)
[![[Telegram]](https://img.shields.io/badge/%20chat-@go__tg__devs-blue.svg?style=flat-square)](https://t.me/go_tg_devs)

[![beta](https://img.shields.io/badge/-beta-yellow)](https://go-faster.org/docs/projects/status)

- [Features](#features)
- [Install](#install)
- [Quick Example](#quick-example)
- [API Client](#api-client)
  - [Creating](#creating)
  - [Bot API methods](#bot-api-methods)
  - [Low-level Bot API methods call](#low-level-bot-api-methods-call)
  - [Helper methods](#helper-methods)
  - [Sending files](#sending-files)
  - [Downloading files](#downloading-files)
  - [Interceptors](#interceptors)
- [Parse Mode Formatters](#parse-mode-formatters)
- [Keyboard Builders](#keyboard-builders)
- [Updates](#updates)
  - [Handlers](#handlers)
  - [Typed Handlers](#typed-handlers)
  - [Receive updates via Polling](#receive-updates-via-polling)
  - [Receive updates via Webhook](#receive-updates-via-webhook)
  - [Routing updates](#routing-updates)
- [Message Builders](#message-builders)
  - [TextMessageCallBuilder](#textmessagecallbuilder)
  - [MediaMessageCallBuilder](#mediamessagecallbuilder)
- [Structured Callback Data](#structured-callback-data)
- [Extensions](#extensions)
  - [Sessions](#sessions)
- [Related Projects](#related-projects)
- [Projects using this package](#projects-using-this-package)
- [Thanks](#thanks)

go-tg is a Go client library for accessing [Telegram Bot API](https://core.telegram.org/bots/api), with batteries for building complex bots included.

> ⚠️ The API definitions are stable and the package is well tested and used in production. However, go-tg is still under active development and full backward compatibility is not guaranteed before reaching v1.0.0.

## Features

- :rocket: Code for Bot API types and methods is generated with embedded official documentation.
- :white_check_mark: Support [context.Context](https://pkg.go.dev/context).
- :link: API Client and bot framework are strictly separated, you can use them independently.
- :zap: No runtime reflection overhead.
- :arrows_counterclockwise: Supports Webhook and Polling natively;
- :mailbox_with_mail: [Webhook reply](https://core.telegram.org/bots/faq#how-can-i-make-requests-in-response-to-updates) for high load bots;
- :raised_hands: Handlers, filters, and middleware are supported.
- :globe_with_meridians: [WebApps](https://core.telegram.org/bots/webapps) and [Login Widget](https://core.telegram.org/widgets/login) helpers.
- :handshake: Business connections support


## Install

```bash
# go 1.21+
go get -u github.com/mr-linch/go-tg
```

## Quick Example

```go
package main

import (
  "context"
  "fmt"
  "os"
  "os/signal"
  "regexp"
  "syscall"
  "time"

  "github.com/mr-linch/go-tg"
  "github.com/mr-linch/go-tg/tgb"
)

func main() {
  ctx := context.Background()

  ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM)
  defer cancel()

  if err := run(ctx); err != nil {
    fmt.Println(err)
    defer os.Exit(1)
  }
}

func run(ctx context.Context) error {
  client := tg.New(os.Getenv("BOT_TOKEN"))

  router := tgb.NewRouter().
    // handles /start and /help
    Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
      return msg.Answer(
        tg.HTML.Text(
          tg.HTML.Bold("👋 Hi, I'm echo bot!"),
          "",
          tg.HTML.Italic("🚀 Powered by", tg.HTML.Spoiler(tg.HTML.Link("go-tg", "github.com/mr-linch/go-tg"))),
        ),
      ).ParseMode(tg.HTML).DoVoid(ctx)
    }, tgb.Command("start", tgb.WithCommandAlias("help"))).
    // handles gopher image
    Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
      if err := msg.Update.Reply(ctx, msg.AnswerChatAction(tg.ChatActionUploadPhoto)); err != nil {
        return fmt.Errorf("answer chat action: %w", err)
      }

      // emulate thinking :)
      time.Sleep(time.Second)

      return msg.AnswerPhoto(
        tg.NewFileArgURL("https://go.dev/blog/go-brand/Go-Logo/PNG/Go-Logo_Blue.png"),
      ).DoVoid(ctx)

    }, tgb.Regexp(regexp.MustCompile(`(?mi)(go|golang|gopher)[$\s+]?`))).
    // handle other messages
    Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
      return msg.Copy(msg.Chat).DoVoid(ctx)
    }).
    MessageReaction(func(ctx context.Context, reaction *tgb.MessageReactionUpdate) error {
			// sets same reaction to the message
			answer := tg.NewSetMessageReactionCall(reaction.Chat, reaction.MessageID).Reaction(reaction.NewReaction)
			return reaction.Update.Reply(ctx, answer)
		})

  return tgb.NewPoller(
    router,
    client,
    tgb.WithPollerAllowedUpdates(
			tg.UpdateTypeMessage,
      tg.UpdateTypeMessageReaction,
    )
  ).Run(ctx)
}
```

More examples can be found in [examples](https://github.com/mr-linch/go-tg/tree/main/_examples).

## API Client

### Creating

The simplest way to create a client is to call `tg.New` with a token. It uses `http.DefaultClient` and `api.telegram.org` by default:

```go
client := tg.New("<TOKEN>") // from @BotFather
```

With custom [http.Client](https://pkg.go.dev/net/http#Client):

```go
proxyURL, err := url.Parse("http://user:pass@ip:port")
if err != nil {
  return err
}

httpClient := &http.Client{
  Transport: &http.Transport{
    Proxy: http.ProxyURL(proxyURL),
  },
}

client := tg.New("<TOKEN>",
 tg.WithClientDoer(httpClient),
)
```

With self hosted Bot API server:

```go
client := tg.New("<TOKEN>",
 tg.WithClientServerURL("http://localhost:8080"),
)
```

### Bot API methods

All API methods are supported with embedded official documentation.
It's provided via Client methods.

e.g. [`getMe`](https://core.telegram.org/bots/api#getme) call:

```go
me, err := client.GetMe().Do(ctx)
if err != nil {
 return err
}

log.Printf("authorized as @%s", me.Username)
```

[`sendMessage`](https://core.telegram.org/bots/api#sendmessage) call with required and optional arguments:

```go
peer := tg.Username("MrLinch")

msg, err := client.SendMessage(peer, "<b>Hello, world!</b>").
  ParseMode(tg.HTML). // optional passed like this
  Do(ctx)

if err != nil {
  return err
}

log.Printf("sent message id %d", msg.ID)
```

Some Bot API methods do not return the object and just say `True`. So, you should use the `DoVoid` method to execute calls like that.

All calls with the returned object also have the `DoVoid` method. Use it when you do not care about the result, just ensure it's not an error (unmarshaling will also be skipped).

```go
peer := tg.Username("MrLinch")

if err := client.SendChatAction(
  peer,
  tg.ChatActionTyping
).DoVoid(ctx); err != nil {
  return err
}
```

### Low-level Bot API methods call

Client has method [`Do`](https://pkg.go.dev/github.com/mr-linch/go-tg#Client.Do) for low-level [requests](https://pkg.go.dev/github.com/mr-linch/go-tg#Request) execution:

```go
req := tg.NewRequest("sendChatAction").
  PeerID("chat_id", tg.Username("@MrLinch")).
  String("action", "typing")

if err := client.Do(ctx, req, nil); err != nil {
  return err
}
```

### Helper methods

Method [`Client.Me()`](https://pkg.go.dev/github.com/mr-linch/go-tg#Client.Me) fetches authorized bot info via [`Client.GetMe()`](https://pkg.go.dev/github.com/mr-linch/go-tg#Client.GetMe) and cache it between calls.

```go
me, err := client.Me(ctx)
if err != nil {
  return err
}
```

### Sending files

There are several ways to send files to Telegram:

- uploading a file along with a method call;
- sending a previously uploaded file by its identifier;
- sending a file using a URL from the Internet;

The [`FileArg`](https://pkg.go.dev/github.com/mr-linch/go-tg#FileArg) type is used to combine all these methods. It is an object that can be passed to client methods and depending on its contents the desired method will be chosen to send the file.

Consider each method by example.

**Uploading a file along with a method call:**

For upload a file you need to create an object [`tg.InputFile`](https://pkg.go.dev/github.com/mr-linch/go-tg#InputFile). It is a structure with two fields: file name and [`io.Reader`](https://pkg.go.dev/io#Reader) with its contents.

Type has some handy constructors, for example consider uploading a file from a local file system:

```go
inputFile, err := tg.NewInputFileLocal("/path/to/file.pdf")
if err != nil {
  return err
}
defer inputFile.Close()

peer := tg.Username("MrLinch")

if err := client.SendDocument(
  peer,
  tg.NewFileArgUpload(inputFile),
).DoVoid(ctx); err != nil {
	return err
}
```

Loading a file from a buffer in memory:

```go

buf := bytes.NewBufferString("<html>...</html>")

inputFile := tg.NewInputFile("index.html", buf)

peer := tg.Username("MrLinch")

if err := client.SendDocument(
  peer,
  tg.NewFileArgUpload(inputFile),
).DoVoid(ctx); err != nil {
	return err
}
```

**Sending a file using a URL from the Internet:**

```go
peer := tg.Username("MrLinch")

if err := client.SendPhoto(
  peer,
  tg.NewFileArgURL("https://picsum.photos/500"),
).DoVoid(ctx); err != nil {
	return err
}
```

**Sending a previously uploaded file by its identifier:**

```go
peer := tg.Username("MrLinch")

if err := client.SendPhoto(
  peer,
  tg.NewFileArgID(tg.FileID("AgACAgIAAxk...")),
).DoVoid(ctx); err != nil {
	return err
}
```

Please checkout [examples](https://github.com/mr-linch/go-tg/tree/main/_examples) with "File Upload" features for more usecases.

### Downloading files

To download a file you need to get its [`FileID`](https://pkg.go.dev/github.com/mr-linch/go-tg#FileID).
After that you need to call method [`Client.GetFile`](https://pkg.go.dev/github.com/mr-linch/go-tg#Client.GetFile) to get metadata about the file.
At the end we call method [`Client.Download`](https://pkg.go.dev/github.com/mr-linch/go-tg#Client.Download) to fetch the contents of the file.

```go

fid := tg.FileID("AgACAgIAAxk...")

file, err := client.GetFile(fid).Do(ctx)
if err != nil {
  return err
}

f, err := client.Download(ctx, file.FilePath)
if err != nil {
  return err
}
defer f.Close()

// ...
```

### Interceptors

Interceptors are used to modify or process the request before it is sent to the server and the response before it is returned to the caller. It's like a [tgb.Middleware](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Middleware), but for outgoing requests.

All interceptors should be registered on the client before the request is made.

```go
client := tg.New("<TOKEN>",
  tg.WithClientInterceptors(
    tg.Interceptor(func(ctx context.Context, req *tg.Request, dst any, invoker tg.InterceptorInvoker) error {
      started := time.Now()

      // before request
      err := invoker(ctx, req, dst)
      // after request

      log.Printf("call %s took %s", req.Method, time.Since(started))

      return err
    }),
  ),
)
```

Arguments of the interceptor are:

- `ctx` - context of the request;
- `req` - request object [tg.Request](https://pkg.go.dev/github.com/mr-linch/go-tg#Request);
- `dst` - pointer to destination for the response, can be `nil` if the request is made with `DoVoid` method;
- `invoker` - function for calling the next interceptor or the actual request.

Contrib package has some useful interceptors:

- [InterceptorRetryFloodError](https://pkg.go.dev/github.com/mr-linch/go-tg#NewInterceptorRetryFloodError) - retry request if the server returns a flood error. Parameters can be customized via options;
- [InterceptorRetryInternalServerError](https://pkg.go.dev/github.com/mr-linch/go-tg#NewInterceptorRetryInternalServerError) - retry request if the server returns an error. Parameters can be customized via options;
- [InterceptorMethodFilter](https://pkg.go.dev/github.com/mr-linch/go-tg#NewInterceptorMethodFilter) - call underlying interceptor only for specified methods;
- [InterceptorDefaultParseMethod](https://pkg.go.dev/github.com/mr-linch/go-tg#NewInterceptorDefaultParseMethod) - set default `parse_mode` for messages if not specified.

Interceptors are called in the order they are registered.

Example of using retry flood interceptor: [examples/retry-flood](https://github.com/mr-linch/go-tg/blob/main/_examples/retry-flood/main.go)

## Parse Mode Formatters

The [`tg.ParseMode`](https://pkg.go.dev/github.com/mr-linch/go-tg#ParseMode) interface provides a fluent API for formatting message text. Three modes are available: `tg.HTML`, `tg.MD2` (MarkdownV2), and `tg.MD` (legacy Markdown).

```go
pm := tg.HTML

text := pm.Text(
  pm.Bold("Order confirmed"),
  "",
  pm.Line("Item:", pm.Code("SKU-42")),
  pm.Line("Price:", pm.Bold("$9.99")),
  "",
  pm.Italic("Thank you for your purchase!"),
)

// sends:
// <b>Order confirmed</b>
//
// Item: <code>SKU-42</code>
// Price: <b>$9.99</b>
//
// <i>Thank you for your purchase!</i>
```

`Text(parts...)` joins with newlines, `Line(parts...)` joins with spaces.

**Formatting methods:** `Bold`, `Italic`, `Underline`, `Strike`, `Spoiler`, `Code`, `Pre`, `PreLanguage`, `Blockquote`, `ExpandableBlockquote`.

**Links and mentions:** `Link(title, url)`, `Mention(name, userID)`, `CustomEmoji(emoji, emojiID)`.

**Escaping:** `Escape(v)` escapes special characters for the current mode. `Escapef(format, args...)` escapes the format string while passing args through unchanged — useful for MarkdownV2 where characters like `.` `!` `|` need escaping:

```go
pm := tg.MD2

pm.Escapef("Total: %s for %s", pm.Bold("$9.99"), pm.Code("SKU-42"))
// escapes "Total:" and "for" but leaves Bold/Code output intact
```

See full example: [examples/parse-mode](https://github.com/mr-linch/go-tg/tree/main/_examples/parse-mode).

## Keyboard Builders

`tg.NewInlineKeyboard` and `tg.NewReplyKeyboard` provide a fluent API for building keyboards.

**Inline keyboard with explicit rows:**

```go
kb := tg.NewInlineKeyboard().
    Callback("📋 Orders", "orders").Callback("⚙ Settings", "settings").Row().
    URL("📖 Docs", "https://example.com/docs")

msg.Answer("Menu").ReplyMarkup(kb).DoVoid(ctx)
```

**Dynamic buttons with `Adjust`:**

`Adjust(sizes...)` redistributes buttons into rows with a repeating size pattern.

```go
kb := tg.NewInlineKeyboard()
for _, item := range items {
    kb.Button(itemFilter.MustButton(item.Name, itemData{ID: item.ID}))
}

msg.Answer("Items:").ReplyMarkup(kb.Adjust(2)).DoVoid(ctx)
// 2 buttons per row
```

**Mixing static and dynamic rows:**

```go
kb := tg.NewInlineKeyboard().
    Callback("A", "a").Callback("B", "b").Callback("C", "c").Row()
for _, item := range items {
    kb.Callback(item.Name, "item:"+item.ID)
}
kb.Adjust(4)
kb.Callback("Back", "back")
// [A] [B] [C]         ← static
// [I1] [I2] [I3] [I4] ← dynamic
// [I5] [I6]            ← remainder
// [Back]               ← static
```

**Reply keyboard with options:**

```go
kb := tg.NewReplyKeyboard().
    Text("Male").Text("Female").Text("Other").
    Resize().OneTime()

msg.Answer("Gender?").ReplyMarkup(kb).DoVoid(ctx)
```

**Methods:**

- `Button(buttons...)` — add pre-built buttons (e.g. from `CallbackFilter.MustButton`)
- `Row()` — end the current row and start a new one
- `Adjust(sizes...)` — redistribute uncommitted buttons into rows with repeating pattern
- `Markup()` — return the underlying `InlineKeyboardMarkup` / `ReplyKeyboardMarkup`

Both builders implement `ReplyMarkup` and can be passed directly to `.ReplyMarkup()`.

## Updates

Everything related to receiving and processing updates is in the [`tgb`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb) package.

### Handlers

You can create an update handler in three ways:

1. Declare the structure that implements the interface [`tgb.Handler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Handler):

```go
type MyHandler struct {}

func (h *MyHandler) Handle(ctx context.Context, update *tgb.Update) error {
  if update.Message == nil {
    return nil
  }

 log.Printf("update id: %d, message id: %d", update.ID, update.Message.ID)

 return nil
}
```

2. Wrap the function to the type [`tgb.HandlerFunc`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#HandlerFunc):

```go
var handler tgb.Handler = tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
 // skip updates of other types
 if update.Message == nil {
  return nil
 }

 log.Printf("update id: %d, message id: %d", update.ID, update.Message.ID)

 return nil
})
```

3. Wrap the function to the type `tgb.*Handler` for creating typed handlers with null pointer check:

```go
// that handler will be called only for messages
// other updates will be ignored
var handler tgb.Handler = tgb.MessageHandler(func(ctx context.Context, mu *tgb.MessageUpdate) error {
  log.Printf("update id: %d, message id: %d", mu.Update.ID, mu.ID)
  return nil
})
```

### Typed Handlers

For each subtype (field) of [`tg.Update`](https://pkg.go.dev/github.com/mr-linch/go-tg/tg#Update) you can create a typed handler.

Typed handlers it's not about routing updates but about handling them.
These handlers will only be called for updates of a certain type, the rest will be skipped. Also, they implement the [`tgb.Handler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Handler) interface.

List of typed handlers:

- [`tgb.MessageHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#MessageHandler) with [`tgb.MessageUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#MessageUpdate) for `message`, `edited_message`, `channel_post`, `edited_channel_post`, `business_message`, `edited_business_message`;
- [`tgb.InlineQueryHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#InlineQueryHandler) with [`tgb.InlineQueryUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#InlineQueryUpdate) for `inline_query`;
- [`tgb.ChosenInlineResultHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChosenInlineResultHandler) with [`tgb.ChosenInlineResultUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChosenInlineResultUpdate) for `chosen_inline_result`;
- [`tgb.CallbackQueryHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#CallbackQueryHandler) with [`tgb.CallbackQueryUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#CallbackQueryUpdate) for `callback_query`;
- [`tgb.ShippingQueryHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ShippingQueryHandler) with [`tgb.ShippingQueryUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ShippingQueryUpdate) for `shipping_query`;
- [`tgb.PreCheckoutQueryHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PreCheckoutQueryHandler) with [`tgb.PreCheckoutQueryUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PreCheckoutQueryUpdate) for `pre_checkout_query`;
- [`tgb.PollHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PollHandler) with [`tgb.PollUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PollUpdate) for `poll`;
- [`tgb.PollAnswerHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PollAnswerHandler) with [`tgb.PollAnswerUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PollAnswerUpdate) for `poll_answer`;
- [`tgb.ChatMemberUpdatedHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChatMemberUpdatedHandler) with [`tgb.ChatMemberUpdatedUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChatMemberUpdatedUpdate) for `my_chat_member`, `chat_member`;
- [`tgb.ChatJoinRequestHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChatJoinRequestHandler) with [`tgb.ChatJoinRequestUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChatJoinRequestUpdate) for `chat_join_request`;
- [`tgb.MessageReactionHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#MessageReactionHandler) with [`tgb.MessageReactionUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#MessageReactionUpdate) for `message_reaction`;
- [`tgb.MessageReactionCountHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#MessageReactionCountHandler) with [`tgb.MessageReactionCountUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#MessageReactionCountUpdate) for `message_reaction_count`;
- [`tgb.ChatBoostHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChatBoostHandler) with [`tgb.ChatBoostUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChatBoostUpdate) for `chat_boost`;
- [`tgb.RemovedChatBoostHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#RemovedChatBoostHandler) with [`tgb.RemovedChatBoostUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#RemovedChatBoostUpdate) for `removed_chat_boost`;
- [`tgb.BusinessConnectionHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#BusinessConnectionHandler) with [`tgb.BusinessConnectionUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#BusinessConnectionUpdate) for `business_connection`;
- [`tgb.DeletedBusinessMessagesHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#DeletedBusinessMessagesHandler) with [`tgb.DeletedBusinessMessagesUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#DeletedBusinessMessagesUpdate) for `deleted_business_messages`;
- [`tgb.PurchasedPaidMediaHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PurchasedPaidMediaHandler) with [`tgb.PurchasedPaidMediaUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PurchasedPaidMediaUpdate) for `purchased_paid_media`;

`tgb.*Updates` has many useful methods for "answer" the update, please checkout godoc by links above.

### Receive updates via Polling

Use [`tgb.NewPoller`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#NewPoller) to create a poller with specified [`tg.Client`](https://pkg.go.dev/github.com/mr-linch/go-tg/tg#Client) and [`tgb.Handler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Handler). Also accepts [`tgb.PollerOption`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PollerOption) for customizing the poller.

```go
handler := tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
  // ...
})

poller := tgb.NewPoller(handler, client,
  // receive max 100 updates in a batch
  tgb.WithPollerLimit(100),
)

// polling will be stopped on context cancel
if err := poller.Run(ctx); err != nil {
  return err
}

```

### Receive updates via Webhook

Webhook handler and server can be created by [`tgb.NewWebhook`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#NewWebhook).
That function has following arguments:

- `handler` - [`tgb.Handler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Handler) for handling updates;
- `client` - [`tg.Client`](https://pkg.go.dev/github.com/mr-linch/go-tg/tg#Client) for making setup requests;
- `url` - full url of the webhook server
- optional `options` - [`tgb.WebhookOption`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#WebhookOption) for customizing the webhook.

Webhook has several security checks that are enabled by default:

- Check if the IP of the sender is in the [allowed ranges](https://core.telegram.org/bots/webhooks#the-short-version).
- Check if the request has a valid security token [header](https://core.telegram.org/bots/api#setwebhook). By default, the token is the SHA256 hash of the Telegram Bot API token.

> ℹ️ These checks can be disabled by passing `tgb.WithWebhookSecurityToken(""), tgb.WithWebhookSecuritySubnets()` when creating the webhook.

> ⚠️ At the moment, the webhook does not integrate custom certificate. So, you should handle HTTPS requests on load balancer.

```go
handler := tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
   // ...
})


webhook := tgb.NewWebhook(handler, client, "https://bot.com/webhook",
  tgb.WithDropPendingUpdates(true),
)

// configure telegram webhook and start HTTP server.
// the server will be stopped on context cancel.
if err := webhook.Run(ctx, ":8080"); err != nil {
  return err
}
```

Webhook is a regular [`http.Handler`](https://pkg.go.dev/net/http#Handler) that can be used in any HTTP-compatible router. But you should call [`Webhook.Setup`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Webhook.Setup) before starting the server to configure the webhook on the Telegram side.

**e.g. integration with [chi](https://pkg.go.dev/github.com/go-chi/chi/v5) router**

```go
handler := tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
  // ...
})

webhook := tgb.NewWebhook(handler, client, "https://bot.com/webhook",
  tgb.WithDropPendingUpdates(true),
)

// get current webhook configuration and sync it if needed.
if err := webhook.Setup(ctx); err != nil {
  return err
}

r := chi.NewRouter()

r.Post("/webhook", webhook)

http.ListenAndServe(":8080", r)

```

### Routing updates

When building complex bots, routing updates is one of the most boilerplate parts of the code.
The [`tgb`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb) package contains a number of primitives to simplify this.

#### [`tgb.Router`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Router)

This is an implementation of [`tgb.Handler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Handler), which provides the ability to route updates between multiple related handlers.
It is useful for handling updates in different ways depending on the update subtype.

```go
router := tgb.NewRouter()

router.Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
  // will be called for every Update with not nil `Message` field
})

router.EditedMessage(func(ctx context.Context, mu *tgb.MessageUpdate) error {
  // will be called for every Update with not nil `EditedMessage` field
})

router.CallbackQuery(func(ctx context.Context, update *tgb.CallbackQueryUpdate) error {
  // will be called for every Update with not nil `CallbackQuery` field
})

client := tg.New(...)

// e.g. run in long polling mode
if err := tgb.NewPoller(router, client).Run(ctx); err != nil {
  return err
}
```

#### [tgb.Filter](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Filter)

Routing by update subtype is first level of the routing. Second is **filters**. Filters are needed to determine more precisely which handler to call, for which update, depending on its contents.

In essence, filters are predicates. Functions that return a boolean value.
If the value is `true`, then the given update corresponds to a handler and the handler will be called.
If the value is `false`, check the subsequent handlers.

The [`tgb`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb) package contains many built-in filters.

e.g. [command filter](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Command) (can be customized via [`CommandFilterOption`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#CommandFilterOption))

```go
router.Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
  // will be called for every Update with not nil `Message` field and if the message text contains "/start"
}, tgb.Command("start"))
```

The handler registration function accepts any number of filters.
They will be combined using the boolean operator `and`

e.g. handle /start command in private chats only

```go
router.Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
  // will be called for every Update with not nil `Message` field
  //  and
  // if the message text contains "/start"
  //  and
  // if the Message.Chat.Type is private
}, tgb.Command("start"), tgb.ChatType(tg.ChatTypePrivate))
```

Logical operator `or` also supported.

e.g. handle /start command in groups or supergroups only

```go
isGroupOrSupergroup := tgb.Any(
  tgb.ChatType(tg.ChatTypeGroup),
  tgb.ChatType(tg.ChatTypeSupergroup),
)

router.Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
  // will be called for every Update with not nil `Message` field
  //  and
  // if the message text contains "/start"
  //  and
  //    if the Message.Chat.Type is group
  //      or
  //    if the Message.Chat.Type is supergroup
}, tgb.Command("start"), isGroupOrSupergroup)
```

All filters are universal. e.g. the command filter can be used in the `Message`, `EditedMessage`, `ChannelPost`, `EditedChannelPost` handlers.
Please checkout [`tgb.Filter`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Filter) constructors for more information about built-in filters.

For define a custom filter you should implement the [`tgb.Filter`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Filter) interface. Also you can use [`tgb.FilterFunc`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#FilterFunc) wrapper to define a filter in functional way.

e.g. filter for messages with document attachments with image type

```go
// tgb.All works like boolean `and` operator.
var isDocumentPhoto = tgb.All(
  tgb.MessageType(tg.MessageTypeDocument),
  tgb.FilterFunc(func(ctx context.Context, update *tgb.Update) (bool, error) {
    return strings.HasPrefix(update.Message.Document.MIMEType, "image/"), nil
  }),
)
```

#### [tgb.Middleware](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Middleware)

Middleware is used to modify or process the Update before it is passed to the handler.
All middleware should be registered before the handlers registration.

e.g. log all updates

```go
router.Use(func(next tgb.Handler) tgb.Handler {
  return tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
    defer func(started time.Time) {
      log.Printf("%#v [%s]", update, time.Since(started))
    }(time.Now())

    return next.Handle(ctx, update)
  })
})
```

#### Error Handler

All handlers return an `error`. If any error occurs in the chain, it will be passed to the error handler. By default, errors are returned as-is. You can customize this behavior by registering a custom error handler.

e.g. log all errors

```go
router.Error(func(ctx context.Context, update *tgb.Update, err error) error {
  log.Printf("error when handling update #%d: %v", update.ID, err)
  return nil
})
```

That example is not useful and just demonstrates the error handler.
The better way to achieve this is simply to enable logging in [`Webhook`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Webhook) or [`Poller`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Poller).

### Message Builders

When building bots with inline keyboards, you often need to send the same message as a new message in one handler and edit an existing message in another (e.g., responding to a `/start` command vs. updating on a callback button press). Message builders let you define the message content once and convert it to different API calls as needed.

#### [`tgb.TextMessageCallBuilder`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#TextMessageCallBuilder)

Builds text messages that can be sent, edited, or have their reply markup updated.

```go
builder := tgb.NewTextMessageCallBuilder(
  tg.HTML.Text(
    tg.HTML.Bold("Hello!"),
    "",
    tg.HTML.Italic("Select an option:"),
  ),
).
  ParseMode(tg.HTML).
  ReplyMarkup(tg.NewInlineKeyboardMarkup(
    tg.NewButtonRow(
      tg.NewInlineKeyboardButtonCallbackData("Option 1", "opt:1"),
      tg.NewInlineKeyboardButtonCallbackData("Option 2", "opt:2"),
    ),
  ))
```

Fluent setters: `Text`, `ParseMode`, `ReplyMarkup`, `LinkPreviewOptions`, `Entities`, `BusinessConnectionID`, `Client`.

**Conversion methods:**

- `AsSend(peer)` → `sendMessage`
- `AsEditText(peer, id)` / `FromCBQ` / `FromMsg` / `Inline` → `editMessageText`
- `AsEditReplyMarkup(peer, id)` / `FromCBQ` / `FromMsg` / `Inline` → `editMessageReplyMarkup`

**Example:** reusable menu message used for both initial send and callback edits:

```go
func newMenuMessage(items []Item) *tgb.TextMessageCallBuilder {
  pm := tg.HTML
  kb := tg.NewInlineKeyboard()
  for _, item := range items {
    kb.Button(itemFilter.MustButton(item.Name, itemData{ID: item.ID}))
  }

  return tgb.NewTextMessageCallBuilder(
    pm.Text(pm.Bold("Menu"), "", pm.Italic("Select an item:")),
  ).
    ParseMode(pm).
    ReplyMarkup(kb.Adjust(2).Markup())
}

router.
  // send menu as new message on /start
  Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
    return msg.Update.Reply(ctx, newMenuMessage(items).AsSend(msg.Chat))
  }, tgb.Command("start")).
  // edit existing message on "back" callback
  CallbackQuery(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate) error {
    return cbq.Update.Reply(ctx, newMenuMessage(items).AsEditTextFromCBQ(cbq.CallbackQuery))
  }, backFilter.Filter())
```

See full example: [examples/menu](https://github.com/mr-linch/go-tg/tree/main/_examples/menu).

#### [`tgb.MediaMessageCallBuilder`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#MediaMessageCallBuilder)

Builds caption-based media messages that can be sent as different media types, or used to edit captions and media.

```go
builder := tgb.NewMediaMessageCallBuilder(
  tg.HTML.Text(tg.HTML.Bold("Mountain Lake"), "", "A serene mountain lake."),
).
  ParseMode(tg.HTML).
  ShowCaptionAboveMedia(true).
  ReplyMarkup(keyboard)
```

Fluent setters: `Caption`, `ParseMode`, `ReplyMarkup`, `CaptionEntities`, `ShowCaptionAboveMedia`, `BusinessConnectionID`, `Client`.

**Conversion methods** — each send method takes a `tg.PeerID` and a `tg.FileArg`:

- `AsSendPhoto` / `AsSendVideo` / `AsSendAudio` / `AsSendDocument` / `AsSendAnimation` / `AsSendVoice` → corresponding `send*` method
- `AsEditCaption(peer, id)` / `FromCBQ` / `FromMsg` / `Inline` → `editMessageCaption`
- `AsEditMedia(peer, id, media)` / `FromCBQ` / `FromMsg` / `Inline` → `editMessageMedia`

**InputMedia helpers** — create `tg.InputMedia` with the builder's caption settings pre-filled:

`NewInputMediaPhoto`, `NewInputMediaVideo`, `NewInputMediaAnimation`, `NewInputMediaAudio`, `NewInputMediaDocument`.

**Example:** photo gallery with navigation buttons:

```go
func newGalleryMessage(index int) *tgb.MediaMessageCallBuilder {
  item := gallery[index]
  pm := tg.HTML

  prev := (index - 1 + len(gallery)) % len(gallery)
  next := (index + 1) % len(gallery)

  return tgb.NewMediaMessageCallBuilder(
    pm.Text(pm.Bold(item.Title), "", pm.Escape(item.Description)),
  ).
    ParseMode(pm).
    ShowCaptionAboveMedia(true).
    ReplyMarkup(tg.NewInlineKeyboard().
      Button(
        navFilter.MustButton("< Prev", nav{Index: prev}),
        navFilter.MustButton("Next >", nav{Index: next}),
      ).Markup())
}

router.
  // send photo on /start
  Message(func(ctx context.Context, msg *tgb.MessageUpdate) error {
    b := newGalleryMessage(0)
    return msg.Update.Reply(ctx, b.AsSendPhoto(msg.Chat, tg.NewFileArgURL(gallery[0].PhotoURL)))
  }, tgb.Command("start")).
  // navigate gallery on button press
  CallbackQuery(navFilter.Handler(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate, n nav) error {
    b := newGalleryMessage(n.Index)
    photo := b.NewInputMediaPhoto(tg.NewFileArgURL(gallery[n.Index].PhotoURL))
    return cbq.Update.Reply(ctx, b.AsEditMediaFromCBQ(cbq.CallbackQuery, photo))
  }), navFilter.Filter())
```

See full example: [examples/media-gallery](https://github.com/mr-linch/go-tg/tree/main/_examples/media-gallery).

### Structured Callback Data

[`tgb.CallbackDataFilter[T]`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#CallbackDataFilter) provides type-safe, declarative routing for inline keyboards. Instead of manually parsing `callback_data` strings, you define Go structs for each action — the filter handles encoding (compact enough for the 64-byte Telegram limit), prefix-based routing, and automatic decoding in handlers.

**1. Define a struct and create a filter:**

```go
type PageNav struct {
  Page int
}

var pageFilter = tgb.NewCallbackDataFilter[PageNav]("page")
```

The filter encodes structs as `"page:1"` (integers use base-36 by default for compactness). Supported field types: `bool`, `int*`, `uint*`, `float*`, `string`.

**2. Create buttons with encoded data:**

```go
tg.NewInlineKeyboard().
  Button(
    pageFilter.MustButton("< Prev", PageNav{Page: page - 1}),
    pageFilter.MustButton("Next >", PageNav{Page: page + 1}),
  )
```

`MustButton(text, value)` encodes the struct into `callback_data` and returns an `InlineKeyboardButton`. Use `Button(text, value)` if you need error handling.

**3. Route and handle with automatic decoding:**

```go
router.CallbackQuery(
  pageFilter.Handler(func(ctx context.Context, cbq *tgb.CallbackQueryUpdate, nav PageNav) error {
    // nav.Page is already decoded
    return cbq.Update.Reply(ctx, newPageMessage(nav.Page).AsEditTextFromCBQ(cbq.CallbackQuery))
  }),
  pageFilter.Filter(), // matches callbacks with "page:" prefix
)
```

`Filter()` matches callback queries by prefix. `Handler()` wraps your handler and passes the decoded struct as a third argument.

**Codec options** can be passed to `NewCallbackDataFilter` to customize encoding:

```go
var filter = tgb.NewCallbackDataFilter[MyData]("prefix",
  tgb.WithCallbackDataCodecDelimiter(';'),  // field separator (default: ':')
  tgb.WithCallbackDataCodecIntBase(10),     // decimal integers (default: 36)
)
```

Per-field overrides are available via struct tags: `` `tgbase:"16"` ``, `` `tgfmt:"e"` ``, `` `tgprec:"2"` ``.

See full example: [examples/menu](https://github.com/mr-linch/go-tg/tree/main/_examples/menu).

## Extensions

### Sessions

#### What is a Session?

Session it's a simple storage for data related to the Telegram chat.
It allow you to share data between different updates from the same chat.
This data is persisted in the [session store](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb/session#Store) and will be available for the next updates from the same chat.

In fact, the session is the usual `struct` and you can define it as you wish.
One requirement is that the session must be serializable.
By default, the session is serialized using [`encoding/json`](https://pkg.go.dev/encoding/json) package, but you can use any other marshal/unmarshal funcs.

#### When not to use sessions?

- you need to store large amount of data;
- your data is not serializable;
- you need access to data from other chat sessions;
- session data should be used by other systems;

#### Where sessions store

Session store is simple key-value storage.
Where key is a string value unique for each chat and value is serialized session data.
By default, manager use [`StoreMemory`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb/session#StoreMemory) implementation.
Also package has [`StoreFile`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb/session#StoreFile) based on FS.

#### How to use sessions?

1. You should define a session struct:
   ```go
    type Session struct {
      PizzaCount int
    }
   ```
2. Create a session manager:
   ```go
    var sessionManager = session.NewManager(Session{
      PizzaCount: 0,
    })
   ```
3. Attach the session manager to the router:
   ```go
    router.Use(sessionManager)
   ```
4. Use the session manager in the handlers:
   ```go
    router.Message(func(ctx context.Context, mu *tgb.Update) error {
      count := strings.Count(strings.ToLower(mu.Message.Text), "pizza") + strings.Count(mu.Message.Text, "🍕")
      if count > 0 {
        session := sessionManager.Get(ctx)
        session.PizzaCount += count
      }
      return nil
    })
   ```

See [session](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb/session) package and [examples](https://github.com/mr-linch/go-tg/tree/main/_examples) with `Session Manager` feature for more information.

## Related Projects

- [`mr-linch/go-tg-bot`](https://github.com/mr-linch/go-tg-bot) - one click boilerplate for creating Telegram bots with PostgreSQL database and clean architecture;
- [`bots-house/docker-telegram-bot-api`](https://github.com/bots-house/docker-telegram-bot-api) - docker image for running self-hosted Telegram Bot API with automated CI build;

## Projects using this package

- [@ttkeeperbot](https://t.me/ttkeeperbot) - Automatically upload tiktoks in groups and verify users 🇺🇦

## Thanks

- [gotd/td](https://github.com/gotd/td) for inspiration for the use of codegen;
- [aiogram/aiogram](https://github.com/aiogram/aiogram) for handlers, middlewares, filters concepts;
