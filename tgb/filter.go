package tgb

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	tg "github.com/mr-linch/go-tg"
	"golang.org/x/exp/slices"
)

// Filter is a interface for generic update filter.
type Filter interface {
	Allow(ctx context.Context, update *Update) (bool, error)
}

// The FilterFunc type is an adapter to allow the use of
// ordinary functions as filter. If f is a function
// with the appropriate signature, FilterFunc(f) is a
// Filter that calls f.
type FilterFunc func(ctx context.Context, update *Update) (bool, error)

// Allow implements Filter interface.
func (filter FilterFunc) Allow(ctx context.Context, update *Update) (bool, error) {
	return filter(ctx, update)
}

// Any pass update to handler, if any of filters allow it.
func Any(filters ...Filter) Filter {
	return FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
		for _, filter := range filters {
			if allow, err := filter.Allow(ctx, update); err != nil {
				return false, err
			} else if allow {
				return true, nil
			}
		}
		return false, nil
	})
}

// All pass update to handler, if all of filters allow it.
func All(filters ...Filter) Filter {
	return FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
		for _, filter := range filters {
			if allow, err := filter.Allow(ctx, update); err != nil {
				return false, err
			} else if !allow {
				return false, nil
			}
		}
		return true, nil
	})
}

// Not pass update to handler, if specified filter does not allow it.
func Not(filter Filter) Filter {
	return FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
		allow, err := filter.Allow(ctx, update)
		if err != nil {
			return false, err
		}
		return !allow, nil
	})
}

// commandFilter handles commands.
// Filter is registered only for Message updates.
// Custuming filter using WithCommand... options.
type commandFilter struct {
	commands      []string
	prefixies     string
	ignoreMention bool
	ignoreCase    bool
	ignoreCaption bool
}

type CommandFilterOption func(*commandFilter)

// WithCommandPrefix sets allowed command prefixies.
// By default is '/'.
func WithCommandPrefix(prefixes ...string) CommandFilterOption {
	return func(filter *commandFilter) {
		filter.prefixies = strings.Join(prefixes, "")
	}
}

// WithCommandIgnoreMention sets ignore mention in command with mention (/command@username).
// By default is false.
func WithCommandIgnoreMention(ignoreMention bool) CommandFilterOption {
	return func(filter *commandFilter) {
		filter.ignoreMention = ignoreMention
	}
}

// WithCommandIgnoreCase sets ignore case in commands. By default is true.
func WithCommandIgnoreCase(ignoreCase bool) CommandFilterOption {
	return func(filter *commandFilter) {
		filter.ignoreCase = ignoreCase
	}
}

// WithCommandIgnoreCaption sets ignore caption as text source.
// By default is true.
func WithCommandIgnoreCaption(ignoreCaption bool) CommandFilterOption {
	return func(filter *commandFilter) {
		filter.ignoreCaption = ignoreCaption
	}
}

// WithCommandAlias adds alias to command.
func WithCommandAlias(aliases ...string) CommandFilterOption {
	return func(filter *commandFilter) {
		filter.commands = append(filter.commands, aliases...)
	}
}

// Command adds filter for command with specified options.
func Command(command string, opts ...CommandFilterOption) Filter {
	filter := &commandFilter{
		commands:      []string{command},
		prefixies:     "/",
		ignoreCase:    true,
		ignoreMention: false,
		ignoreCaption: true,
	}

	for _, opt := range opts {
		opt(filter)
	}

	if filter.ignoreCase {
		for i, command := range filter.commands {
			filter.commands[i] = strings.ToLower(command)
		}
	}

	return filter
}

// getUpdateMessage returns first not nil message from update fields.
func getUpdateMessage(update *Update) *tg.Message {
	return firstNotNil(
		update.Message,
		update.EditedMessage,
		update.ChannelPost,
		update.EditedChannelPost,
	)
}

func getMessageEntities(message *tg.Message) (entities []tg.MessageEntity) {
	if len(message.Entities) > 0 {
		entities = message.Entities
	} else if len(message.CaptionEntities) > 0 {
		entities = message.CaptionEntities
	} else if message.Poll != nil {
		entities = message.Poll.ExplanationEntities
	} else if message.Game != nil {
		entities = message.Game.TextEntities
	}

	return
}

// Allow checks if update is allowed by filter.
func (filter *commandFilter) Allow(ctx context.Context, update *Update) (bool, error) {
	msg := getUpdateMessage(update)

	if msg == nil {
		return false, nil
	}

	text := msg.Text

	if text == "" && !filter.ignoreCaption {
		text = msg.Caption
	}

	if text == "" {
		return false, nil
	}

	fullCommand, _, _ := strings.Cut(text, " ")

	me, err := update.Client.Me(ctx)
	if err != nil {
		return false, fmt.Errorf("command filter: get current bot info: %w", err)
	}

	prefix := fullCommand[:1]
	command, mention, _ := strings.Cut(fullCommand[1:], "@")

	if filter.ignoreCase {
		command = strings.ToLower(command)
	}

	if !strings.Contains(filter.prefixies, prefix) {
		return false, nil
	}

	if !filter.ignoreMention && mention != "" && !strings.EqualFold(mention, string(me.Username)) {
		return false, nil
	}

	if !slices.Contains(filter.commands, command) {
		return false, nil
	}

	return true, nil
}

func extractUpdateText(update *Update) (string, bool) {
	msg := getUpdateMessage(update)

	switch {
	case msg != nil:
		switch {
		case msg.Text != "":
			return msg.Text, true
		case msg.Caption != "":
			return msg.Caption, true
		case msg.Poll != nil && msg.Poll.Question != "":
			return msg.Poll.Question, true
		}

	case update.CallbackQuery != nil && update.CallbackQuery.Data != "":
		return update.CallbackQuery.Data, true
	case update.InlineQuery != nil && update.InlineQuery.Query != "":
		return update.InlineQuery.Query, true
	case update.ChosenInlineResult != nil && update.ChosenInlineResult.Query != "":
		return update.ChosenInlineResult.Query, true
	case update.Poll != nil && update.Poll.Question != "":
		return update.Poll.Question, true
	}

	return "", false
}

// RegexpFilter handles updates by regexp.
//
// Checks following fields:
//   - Update.Message.Text
//   - Update.Message.Caption
//   - Update.CallbackQuery.Data
//   - Update.InlineQuery.Query
//   - Update.ChosenInlineResult.Query
//   - Update.Poll.Question
func Regexp(re *regexp.Regexp) Filter {
	return FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
		var text string

		text, ok := extractUpdateText(update)
		if !ok {
			return false, nil
		}

		return re.MatchString(text), nil
	})
}

// ChatType filter checks if chat type is in specified list.
//
// Check is performed in:
//   - Message, EditedMessage, ChannelPost, EditedChannelPost
//   - CallbackQuery.Message.Chat.Type (if not nil)
//   - InlineQuery.ChatType
//   - MyChatMember.Chat.Type
//   - ChatMember.Chat.Type
//   - ChatJoinRequest.Chat.Type
func ChatType(types ...tg.ChatType) Filter {
	return FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
		var typ tg.ChatType

		msg := getUpdateMessage(update)

		switch {
		case msg != nil:
			typ = msg.Chat.Type
		case update.CallbackQuery != nil && update.CallbackQuery.Message != nil:
			typ = update.CallbackQuery.Message.Chat().Type
		case update.InlineQuery != nil:
			typ = update.InlineQuery.ChatType
		case update.MyChatMember != nil:
			typ = update.MyChatMember.Chat.Type
		case update.ChatMember != nil:
			typ = update.ChatMember.Chat.Type
		case update.ChatJoinRequest != nil:
			typ = update.ChatJoinRequest.Chat.Type
		default:
			return false, nil
		}

		return slices.Contains(types, typ), nil
	})
}

// MessageType checks Message, EditedMessage, ChannelPost, EditedChannelPost
// for matching type with specified.
// If multiple types are specified, it checks if message type is one of them.
func MessageType(types ...tg.MessageType) Filter {
	return FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
		msg := getUpdateMessage(update)

		if msg != nil {
			return slices.Contains(types, msg.Type()), nil
		}

		return false, nil
	})
}

// MessageEntity checks Message, EditedMessage, ChannelPost, EditedChannelPost .Entities, .CaptionEntities, .Poll.ExplanationEntities or .Game.TextEntities
// for matching type with specified.
// If multiple types are specified, it checks if message entity type is one of them.
func MessageEntity(types ...tg.MessageEntityType) Filter {
	return FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
		var entities []tg.MessageEntity

		if update.Poll != nil {
			entities = update.Poll.ExplanationEntities
		} else {
			msg := getUpdateMessage(update)
			if msg == nil {
				return false, nil
			}

			entities = getMessageEntities(msg)
		}

		if len(entities) == 0 {
			return false, nil
		}

		for _, entity := range entities {
			if slices.Contains(types, entity.Type) {
				return true, nil
			}
		}

		return false, nil
	})
}

// TextFuncFilterOption is a filter option for TextFuncFilter.
type TextFuncFilterOption func(*textFuncFilter)

func WithTextFuncIgnoreCase(v bool) TextFuncFilterOption {
	return func(filter *textFuncFilter) {
		filter.ignoreCase = v
	}
}

// textFuncFilter it's base filter strings comparing filter.
// Checkout constructors for more info.
// All constructors can be customized by TextFuncFilterOption.
type textFuncFilter struct {
	ignoreCase bool
	fn         func(text string, ignoreCase bool) bool
}

func (filter *textFuncFilter) Allow(ctx context.Context, update *Update) (bool, error) {
	text, ok := extractUpdateText(update)
	if !ok || text == "" {
		return false, nil
	}

	return filter.fn(text, filter.ignoreCase), nil
}

// TextFunc creates a generic TextFuncFilter with specified function.
func TextFunc(fn func(text string, ignoreCase bool) bool, opts ...TextFuncFilterOption) Filter {
	filter := &textFuncFilter{
		ignoreCase: false,
		fn:         fn,
	}

	for _, opt := range opts {
		opt(filter)
	}

	return filter
}

// TextEqual creates a TextFuncFilter that checks if text of update equals to specified.
func TextEqual(v string, opts ...TextFuncFilterOption) Filter {
	return TextFunc(func(text string, ignoreCase bool) bool {
		if ignoreCase {
			return strings.EqualFold(text, v)
		}

		return text == v
	}, opts...)
}

// TextHasPrefix creates a TextFuncFilter that checks if text of update has prefix.
func TextHasPrefix(v string, opts ...TextFuncFilterOption) Filter {
	return TextFunc(func(text string, ignoreCase bool) bool {
		if ignoreCase {
			text = strings.ToLower(text)
			v = strings.ToLower(v)
		}

		return strings.HasPrefix(text, v)
	}, opts...)
}

// TextHasSuffix creates a TextFuncFilter that checks if text of update has suffix.
func TextHasSuffix(v string, opts ...TextFuncFilterOption) Filter {
	return TextFunc(func(text string, ignoreCase bool) bool {
		if ignoreCase {
			text = strings.ToLower(text)
			v = strings.ToLower(v)
		}

		return strings.HasSuffix(text, v)
	}, opts...)
}

// TextContains creates a TextFuncFilter that checks if text of update contains specified.
func TextContains(v string, opts ...TextFuncFilterOption) Filter {
	return TextFunc(func(text string, ignoreCase bool) bool {
		if ignoreCase {
			text = strings.ToLower(text)
			v = strings.ToLower(v)
		}

		return strings.Contains(text, v)
	}, opts...)
}

// TextIn creates a TextFuncFilter that checks if text of update is in specified slice.
func TextIn(vs []string, opts ...TextFuncFilterOption) Filter {
	return TextFunc(func(text string, ignoreCase bool) bool {
		if ignoreCase {
			for _, v := range vs {
				if strings.EqualFold(text, v) {
					return true
				}
			}

			return false
		}

		return slices.Contains(vs, text)
	}, opts...)
}
