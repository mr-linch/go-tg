package tgb

import (
	"context"
	"fmt"
	"strings"

	tg "github.com/mr-linch/go-tg"
	"golang.org/x/exp/slices"
)

type Filter interface {
	Allow(ctx context.Context, update *tg.Update) (bool, error)
}

type FilterFunc func(ctx context.Context, update *tg.Update) (bool, error)

func (filter FilterFunc) Allow(ctx context.Context, update *tg.Update) (bool, error) {
	return filter(ctx, update)
}

// Any pass update to handler, if any of filters allow it.
func Any(filters ...Filter) Filter {
	return FilterFunc(func(ctx context.Context, update *tg.Update) (bool, error) {
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
	return FilterFunc(func(ctx context.Context, update *tg.Update) (bool, error) {
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

type CommandFilter struct {
	commands      []string
	prefixies     string
	ignoreMention bool
	ignoreCase    bool
	ignoreCaption bool
}

type CommandFilterOption func(*CommandFilter)

func WithCommandPrefix(prefixes ...string) CommandFilterOption {
	return func(filter *CommandFilter) {
		filter.prefixies = strings.Join(prefixes, "")
	}
}

func WithCommandIgnoreMention(ignoreMention bool) CommandFilterOption {
	return func(filter *CommandFilter) {
		filter.ignoreMention = ignoreMention
	}
}

func WithCommandIgnoreCase(ignoreCase bool) CommandFilterOption {
	return func(filter *CommandFilter) {
		filter.ignoreCase = ignoreCase
	}
}

func WithCommandIgnoreCaption(ignoreCaption bool) CommandFilterOption {
	return func(filter *CommandFilter) {
		filter.ignoreCaption = ignoreCaption
	}
}

func WithCommandAlias(aliases ...string) CommandFilterOption {
	return func(filter *CommandFilter) {
		filter.commands = append(filter.commands, aliases...)
	}
}

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

func (filter *CommandFilter) Allow(ctx context.Context, update *tg.Update) (bool, error) {
	if update.Message == nil {
		return false, nil
	}

	text := update.Message.Text

	if text == "" && !filter.ignoreCaption {
		text = update.Message.Caption
	}

	if text == "" {
		return false, nil
	}

	fullCommand, _, _ := strings.Cut(text, " ")

	me, err := update.Client().Me(ctx)
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
