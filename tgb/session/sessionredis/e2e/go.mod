module github.com/mr-linch/go-tg/tgb/session/sessionredis/e2e

go 1.18

replace github.com/mr-linch/go-tg => ../../../..

replace github.com/mr-linch/go-tg/tgb/session/sessionredis => ../

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/go-redis/redis/v9 v9.0.0-beta.2
	github.com/mr-linch/go-tg v0.0.0-00010101000000-000000000000
	github.com/mr-linch/go-tg/tgb/session/sessionredis v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.8.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tomasen/realip v0.0.0-20180522021738-f0c99a92ddce // indirect
	golang.org/x/exp v0.0.0-20220613132600-b0d781184e0d // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
