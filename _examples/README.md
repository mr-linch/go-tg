All examples has one required argument `-token`.
By default, bots is running in a long poll mode.

If you want to run in a webhook mode, you can use the `-webhook-url` flag.
Listen port set to `:8000` by default, you can override it with the `-webhook-listen` flag.

| Name                                                                                      | Description                                                       | Features                                                       |
| ----------------------------------------------------------------------------------------- | ----------------------------------------------------------------- | -------------------------------------------------------------- |
| [calculator](https://github.com/mr-linch/go-tg/tree/main/examples/calculator)             | Inline Keyboard Calculator                                        | Router, Regexp filter, Button Layout                           |
| [chat-type-filter](https://github.com/mr-linch/go-tg/tree/main/examples/chat-type-filter) | Handle messages from different chats, in different handlers       | ChatType filter                                                |
| [echo-bot](https://github.com/mr-linch/go-tg/tree/main/examples/echo-bot)                 | Copy original messages, with some special cases                   | ParseMode, Webhook Reply, File Uploading, Regexp filter        |
| [grayscale-image](https://github.com/mr-linch/go-tg/tree/main/examples/grayscale-image)   | Bot grayscale user sended image                                   | File download, file upload, Message type filter, Custom filter |
| [media-group](https://github.com/mr-linch/go-tg/tree/main/examples/media-group)           | Send Media group                                                  | File Uploading, Media Group, Regexp filter, Default handler    |
| [book-bot](https://github.com/mr-linch/go-tg/tree/main/examples/book-bot)                 | Book search via Inline Mode                                       | Inline Mode, Inline Keyboard Markup                            |
| [text-filter](https://github.com/mr-linch/go-tg/tree/main/examples/text-filter)           | Text Filter usage                                                 | Text filter, reply keyboard markup                             |
| [webapps](https://github.com/mr-linch/go-tg/tree/main/examples/webapps)                   | Parse and validate Login Widget & WebApp data, host simple webapp | WebApps, Login Widget, Embed webhook to http.Mux               |
| [session-filter](https://github.com/mr-linch/go-tg/tree/main/examples/session-filter)     | Simple form filling with persistent session                       | Router, Session Manager, Session Filters                       |
| [menu](https://github.com/mr-linch/go-tg/tree/main/examples/menu)                         | Hiearchical menu with API integration                             | ButtonLayout, TextMessageBuilder, CallbackDataFilter           |
| [retry-flood](https://github.com/mr-linch/go-tg/tree/main/examples/retry-flood)           | Retry on flood error                                              | Interceptors                                                   |
| [business-bot](https://github.com/mr-linch/go-tg/tree/main/examples/business-bot)         | Business bot with multiple handlers                               | Router                                                         |
