# go-tg 

[![Go Reference](https://pkg.go.dev/badge/github.com/mr-linch/go-tg.svg)](https://pkg.go.dev/github.com/mr-linch/go-tg) 
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/mr-linch/go-tg?label=latest%20release)](https://github.com/mr-linch/go-tg/releases/latest)
![Telegram Bot API](https://img.shields.io/badge/Telegram%20Bot%20API-6.1-blue?logo=telegram)
[![CI](https://github.com/mr-linch/go-tg/actions/workflows/ci.yml/badge.svg)](https://github.com/mr-linch/go-tg/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/mr-linch/go-tg/branch/main/graph/badge.svg?token=9EI5CEIYXL)](https://codecov.io/gh/mr-linch/go-tg)
[![Go Report Card](https://goreportcard.com/badge/github.com/mr-linch/go-tg)](https://goreportcard.com/report/github.com/mr-linch/go-tg) 

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

## Usage

### Quick Example

TODO

### API Client 

#### Creating

-Simple way:

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

#### Bot API methods

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
    ParseMode(tg.HTML). // optional
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

#### Low-level Bot API methods call

Client has method [`Do`](https://pkg.go.dev/github.com/mr-linch/go-tg#Client.Do) for low-level [requests](https://pkg.go.dev/github.com/mr-linch/go-tg#Request) execution: 

```go
req := tg.NewRequest("sendChatAction").
    PeerID("chat_id", tg.Username("@MrLinch")).
    String("action", "typing")

if err := client.Do(ctx, req, nil); err != nil {
    return err
}

```

#### Helper methods

Method [`Client.Me()`](https://pkg.go.dev/github.com/mr-linch/go-tg#Client.Me) fetches authorized bot info via [`Client.GetMe()`](https://pkg.go.dev/github.com/mr-linch/go-tg#Client.GetMe) and cache it between calls. 

```go 
me, err := client.Me(ctx)
if err != nil {
    return err
}
```

