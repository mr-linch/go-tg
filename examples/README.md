All examples has one required argument `-token`.
By default, bots is running in a long poll mode.

If you want to run in a webhook mode, you can use the `-webhook-url` flag.
Listen port set to `:8000` by default, you can override it with the `-webhook-listen` flag.

| Name                                                                                      | Description                                                       | Features                                                    |
| ----------------------------------------------------------------------------------------- | ----------------------------------------------------------------- | ----------------------------------------------------------- |
| [calculator](https://github.com/mr-linch/go-tg/tree/main/examples/calculator)             | Inline Keyboard Calculator                                        | Router, Regexp filter, Button Layout                        |
| [chat-type-filter](https://github.com/mr-linch/go-tg/tree/main/examples/chat-type-filter) | Handle messages from different chats, in different handlers       | ChatType filter                                             |
| [echo-bot](https://github.com/mr-linch/go-tg/tree/main/examples/echo-bot)                 | Copy original messages, with some special cases                   | ParseMode, Webhook Respond, File Uploading, Regexp filter   |
| [media-group](https://github.com/mr-linch/go-tg/tree/main/examples/media-group)           | Send Media group                                                  | File Uploading, Media Group, Regexp filter, Default handler |
| [quote-bot](https://github.com/mr-linch/go-tg/tree/main/examples/quote-bot)               | Search via Inline Mode                                            | Inline Mode, Inline Keyboard Markup                         |
| [text-filter](https://github.com/mr-linch/go-tg/tree/main/examples/text-filter)           | Text Filter usage                                                 | Text filter, reply keyboard markup                          |
| [webapps](https://github.com/mr-linch/go-tg/tree/main/examples/webapps)                   | Parse and validate Login Widget & WebApp data, host simple webapp | WebApps, Login Widget, Embed webhook to http.Mux            |
