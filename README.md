# go-tg 

[![Go Reference](https://pkg.go.dev/badge/github.com/mr-linch/go-tg.svg)](https://pkg.go.dev/github.com/mr-linch/go-tg) 
[![go.mod](https://img.shields.io/github/go-mod/go-version/mr-linch/go-tg)](go.mod)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/mr-linch/go-tg?label=latest%20release)](https://github.com/mr-linch/go-tg/releases/latest)
![Telegram Bot API](https://img.shields.io/badge/Telegram%20Bot%20API-6.1-blue?logo=telegram)
[![CI](https://github.com/mr-linch/go-tg/actions/workflows/ci.yml/badge.svg)](https://github.com/mr-linch/go-tg/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/mr-linch/go-tg/branch/main/graph/badge.svg?token=9EI5CEIYXL)](https://codecov.io/gh/mr-linch/go-tg)
[![Go Report Card](https://goreportcard.com/badge/github.com/mr-linch/go-tg)](https://goreportcard.com/report/github.com/mr-linch/go-tg) 
[![beta](https://img.shields.io/badge/-beta-yellow)](https://go-faster.org/docs/projects/status)


- [Features](#features)
- [Install](#install)
- [Quick Example](#quick-example)
- [API Client](#api-client)
  * [Creating](#creating)
  * [Bot API methods](#bot-api-methods)
  * [Low-level Bot API methods call](#low-level-bot-api-methods-call)
  * [Helper methods](#helper-methods)
- [Updates](#updates)
  * [Handlers](#handlers)
  * [Typed Handlers](#typed-handlers)
  * [Recieve updates via Polling](#recieve-updates-via-polling)
  * [Recieve updates via Webhook](#recieve-updates-via-webhook)


go-tg is a Go client library for accessing [Telegram Bot API](https://core.telegram.org/bots/api), with batteries for building complex bots included.

## Features
 - Code for Bot API types and methods is generated with embedded official documentation.
 - Support [context.Context](https://pkg.go.dev/context).
 - Separte client and bot framework by packages, use only what you need.
 - API Client and bot framework are strictly separeted, you can use them independently. 
 - No runtime reflection overhead. 
 - Supports Webhook and Polling natively;
 - Handlers, filters, and middlewares are supported.



## Install

```bash
go get -u github.com/mr-linch/go-tg
```

## Quick Example

TODO

## API Client 

### Creating

The simplest way for create client it's call `tg.New` with token. That constructor use `http.DefaultClient` as default client and `api.telegram.org` as server URL:

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

All API methods is supported with embedded official documentation.
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

log.Printf("sended message id %d", msg.ID)
```

Some Bot API methods do not return the object and just say `True`. So, you should use the `DoVoid` method to execute calls like that. 

All calls with the returned object also have the `DoVoid` method. Use it when you do not care about the result, just be ensure it's not error (unmarshaling also be skipped). 


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

## Updates

Everything related to receiving and processing updates is in the [`tgb`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb) package. 

### Handlers 

You can create an update handler in three ways: 

1. Declare the structure that implements the interface [`tgb.Handler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Handler): 

```go
type MyHandler struct {}

func (h *MyHandler) Handle(ctx context.Context, update *tgb.Update) error {
  if update.Message != nil {
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
These handlers will only be called for updates of a certain type, the rest will be skipped. Also they impliment the [`tgb.Handler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Handler) interface.


List of typed handlers:
  - [`tgb.MessageHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#MessageHandler) with [`tgb.MessageUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#MessageUpdate) for `message`, `edited_message`, `channel_post`, `edited_channel_post`;
  - [`tgb.InlineQueryHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#InlineQueryHandler) with [`tgb.InlineQueryUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#InlineQueryUpdate) for `inline_query`
  - [`tgb.ChosenInlineResult`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChosenInlineResult) with [`tgb.ChosenInlineResultUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChosenInlineResultUpdate) for `chosen_inline_result`;
  - [`tgb.CallbackQueryHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#CallbackQueryHandler) with [`tgb.CallbackQueryUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#CallbackQueryUpdate) for `callback_query`;
  - [`tgb.ShippingQueryHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ShippingQueryHandler) with [`tgb.ShippingQueryUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ShippingQueryUpdate) for `shipping_query`;
  - [`tgb.PreCheckoutQueryHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PreCheckoutQueryHandler) with [`tgb.PreCheckoutQueryUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PreCheckoutQueryUpdate) for `pre_checkout_query`;
  - [`tgb.PollHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PollHandler) with [`tgb.PollUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PollUpdate) for `poll`;
  - [`tgb.PollAnswerHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PollAnswerHandler) with [`tgb.PollAnswerUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PollAnswerUpdate) for `poll_answer`;
  - [`tgb.ChatMemberUpdatedHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChatMemberUpdatedHandler) with [`tgb.ChatMemberUpdatedUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChatMemberUpdatedUpdate) for `my_chat_member`, `chat_member`;
  - [`tgb.ChatJoinRequestHandler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChatJoinRequestHandler) with [`tgb.ChatJoinRequestUpdate`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#ChatJoinRequestUpdate) for `chat_join_request`;

`tgb.*Updates` has many useful methods for "answer" the update, please checkout godoc by links above.

### Recieve updates via Polling

Use [`tgb.NewPoller`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#NewPoller) to create a poller with specified [`tg.Client`](https://pkg.go.dev/github.com/mr-linch/go-tg/tg#Client) and [`tgb.Handler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Handler). Also accepts [`tgb.PollerOption`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#PollerOption) for customizing the poller.

```go
handler := tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
  // ...
})

poller := tgb.NewPoller(handler, client,
  // recieve max 100 updates in a batch
  tgb.WithPollerLimit(100),
)

// polling will be stopped on context cancel
if err := poller.Run(ctx); err != nil {
  return err
}

```

### Recieve updates via Webhook

Webhook handler and server can be created by [`tgb.NewWebhook`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#NewWebhook).
That function has following arguments: 
 - `handler` - [`tgb.Handler`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Handler) for handling updates;
 - `client` - [`tg.Client`](https://pkg.go.dev/github.com/mr-linch/go-tg/tg#Client) for making setup requests;
 - `url` - full url of the webhook server 
 - optional `options` - [`tgb.WebhookOption`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#WebhookOption) for customizing the webhook.

Webhook has several security checks that are enabled by default: 
 - Check if the IP of the sender in the [allowed ranges](https://core.telegram.org/bots/webhooks#the-short-version).
 - Check if the request has valid security token [header](https://core.telegram.org/bots/api#setwebhook). By default the token is SHA256 hash of Telegram Bot API token. 

> That checks can be disabled by passing `tgb.WithWebhookSecurityToken(""), tgb.WithWebhookSecuritySubnets()` when creating the webhook.

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

Webhook it is regular [`http.Handler`](https://pkg.go.dev/net/http#Handler) that can be used in any HTTP-compatible router. But you should call [`Webhook.Setup`](https://pkg.go.dev/github.com/mr-linch/go-tg/tgb#Webhook.Setup) before starting the server to configure the webhook on Telegram side.

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

r.Get("/webhook", webhook)

http.ListenAndServe(":8080", r)

```