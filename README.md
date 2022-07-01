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


## Usage

### Install

```bash
go get github.com/mr-linch/go-tg
```

