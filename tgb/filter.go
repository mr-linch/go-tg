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

// CommandFilter handles commands.
// Filter is registered only for Message updates.
// Custuming filter using WithCommand... options.
type CommandFilter struct {
	commands      []string
	prefixies     string
	ignoreMention bool
	ignoreCase    bool
	ignoreCaption bool
}

type CommandFilterOption func(*CommandFilter)

// WithCommandPrefix sets allowed command prefixies.
// By default is '/'.
func WithCommandPrefix(prefixes ...string) CommandFilterOption {
	return func(filter *CommandFilter) {
		filter.prefixies = strings.Join(prefixes, "")
	}
}

// WithCommandIgnoreMention sets ignore mention in command with mention (/command@username).
// By default is false.
func WithCommandIgnoreMention(ignoreMention bool) CommandFilterOption {
	return func(filter *CommandFilter) {
		filter.ignoreMention = ignoreMention
	}
}

// WithCommandIgnoreCase sets ignore case in commands. By default is true.
func WithCommandIgnoreCase(ignoreCase bool) CommandFilterOption {
	return func(filter *CommandFilter) {
		filter.ignoreCase = ignoreCase
	}
}

// WithCommandIgnoreCaption sets ignore caption as text source.
// By default is true.
func WithCommandIgnoreCaption(ignoreCaption bool) CommandFilterOption {
	return func(filter *CommandFilter) {
		filter.ignoreCaption = ignoreCaption
	}
}

// WithCommandAlias adds alias to command.
func WithCommandAlias(aliases ...string) CommandFilterOption {
	return func(filter *CommandFilter) {
		filter.commands = append(filter.commands, aliases...)
	}
}

// Command adds filter for command with specified options.
func Command(command string, opts ...CommandFilterOption) *CommandFilter {
	filter := &CommandFilter{
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

// Allow checks if update is allowed by filter.
func (filter *CommandFilter) Allow(ctx context.Context, update *Update) (bool, error) {
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

// RegexpFilter handles updates by regexp.
// Checks following fields:
// - Update.Message.Text
// - Update.Message.Caption
// - Update.CallbackQuery.Data
// - Update.InlineQuery.Query
// - Update.ChosenInlineResult.Query
// - Update.Poll.Question
func Regexp(re *regexp.Regexp) Filter {
	return FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
		var text string

		msg := getUpdateMessage(update)

		switch {
		case msg != nil:
			text = msg.Text

			if text == "" && msg.Caption != "" {
				text = msg.Caption
			}

			if text == "" && msg.Poll != nil {
				text = msg.Poll.Question
			}
		case update.CallbackQuery != nil && update.CallbackQuery.Data != "":
			text = update.CallbackQuery.Data
		case update.InlineQuery != nil && update.InlineQuery.Query != "":
			text = update.InlineQuery.Query
		case update.ChosenInlineResult != nil && update.ChosenInlineResult.Query != "":
			text = update.ChosenInlineResult.Query
		case update.Poll != nil && update.Poll.Question != "":
			text = update.Poll.Question
		default:
			return false, nil
		}

		return re.MatchString(text), nil
	})
}

// ChatType checks chat type in:
// - Message, EditedMessage, ChannelPost, EditedChannelPost
// - CallbackQuery.Message.Chat.Type (if not nil)
// - InlineQuery.ChatType
// - MyChatMember.Chat.Type
// - ChatMember.Chat.Type
// - ChatJoinRequest.Chat.Type
func ChatType(types ...tg.ChatType) Filter {
	return FilterFunc(func(ctx context.Context, update *Update) (bool, error) {
		var typ tg.ChatType

		msg := getUpdateMessage(update)

		switch {
		case msg != nil:
			typ = msg.Chat.Type
		case update.CallbackQuery != nil && update.CallbackQuery.Message != nil:
			typ = update.CallbackQuery.Message.Chat.Type
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
