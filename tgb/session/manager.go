// Package session provides a session managment.
//
// # What is the session?
//
// The session of a chat is a persistent storage with user defined structure.
// As example, you can store user input in the multi-step form, counters, etc.
//
// # Where data is stored?
//
// By default, the data is stored in memory (see [StoreMemory]).
// It's not good, because it's not persistent and you can lose data if the bot is restarted.
// You can use any provided storage, see subpackages of current package for list.
// Or you can use your own storage by implementing [Store] interface.
//
// # How it works?
//
// See [Manager.Wrapper] for details about middleware logic.
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/mr-linch/go-tg/tgb"
)

// KeyFunc is a function that returns a key for a session from update.
type KeyFunc func(update *tgb.Update) string

// KeyFuncChat generate a key from update chat id.
func KeyFuncChat(update *tgb.Update) string {
	chat := update.Chat()

	if chat != nil {
		return strconv.Itoa(int(chat.ID))
	}

	return ""
}

type managerSettings interface {
	setKeyFunc(KeyFunc)
	setStore(Store)

	setEncoding(
		func(v any) ([]byte, error),
		func(data []byte, v any) error,
	)
}

// ManagerOption is a function that sets options for a session manager.
type ManagerOption func(managerSettings)

// WithKeyFunc sets a key function for a get session.
// By default, it uses [KeyFuncChat].
func WithKeyFunc(keyFunc KeyFunc) ManagerOption {
	return func(settings managerSettings) {
		settings.setKeyFunc(keyFunc)
	}
}

// WithStore sets a storage for sessions.
// By default, it uses [StoreMemory].
func WithStore(storage Store) ManagerOption {
	return func(settings managerSettings) {
		settings.setStore(storage)
	}
}

// WithEncoding sets a encoding for sessions.
// By default, it uses [json.Marshal] and [json.Unmarshal].
func WithEncoding(
	encode func(v any) ([]byte, error),
	decode func(data []byte, v any) error,
) ManagerOption {
	return func(settings managerSettings) {
		settings.setEncoding(encode, decode)
	}
}

// Manager provides a persistent data storage for bot.
// You can use it to store chat-specific data persistently.
//

type Manager[T comparable] struct {
	intial  T
	keyFunc KeyFunc
	store   Store

	encodeFunc func(v any) ([]byte, error)
	decodeFunc func(d []byte, v any) error
}

func (manager *Manager[T]) setKeyFunc(keyFunc KeyFunc) {
	manager.keyFunc = keyFunc
}

func (manager *Manager[T]) setStore(store Store) {
	manager.store = store
}

func (manager *Manager[T]) setEncoding(
	encode func(v any) ([]byte, error),
	decode func(d []byte, v any) error,
) {
	manager.encodeFunc = encode
	manager.decodeFunc = decode
}

// NewManager creates a new session manager with session initial value.
// Settings of the manager can be changed by passing [ManagerOption] functions.
//
// Options can be passed later with [Manager.Setup] method.
// It is useful if you want to define [Manager] globally and init with e.g. store later.
func NewManager[T comparable](initial T, opts ...ManagerOption) *Manager[T] {
	manager := &Manager[T]{
		intial: initial,

		keyFunc: KeyFuncChat,
		store:   NewStoreMemory(),

		encodeFunc: json.Marshal,
		decodeFunc: json.Unmarshal,
	}

	for _, opt := range opts {
		opt(manager)
	}

	return manager
}

// Setup it is late initialization method.
func (manager *Manager[T]) Setup(opt ManagerOption, opts ...ManagerOption) {
	for _, opt := range append([]ManagerOption{opt}, opts...) {
		opt(manager)
	}
}

func (manager *Manager[T]) saveSession(ctx context.Context, key string, v *T) error {
	data, err := manager.encodeFunc(v)
	if err != nil {
		return fmt.Errorf("encode session: %w", err)
	}

	return manager.store.Set(ctx, key, data)
}

func (manager *Manager[T]) getSession(ctx context.Context, key string) (*T, error) {
	sessionData, err := manager.store.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if sessionData == nil {
		initial := manager.intial
		return &initial, nil
	}

	var session T

	if err := manager.decodeFunc(sessionData, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

// Get returns Session from [context.Context].
// If session doesn't exist, it returns nil.
func (manager *Manager[T]) Get(ctx context.Context) *T {
	v := ctx.Value(sessionContextKey)
	if v == nil {
		return nil
	}
	return v.(*T)
}

// Reset resets session value to initial value.
//
// If session value equals to initial value, it will be removed from [Store].
func (manager *Manager[T]) Reset(session *T) {
	*session = manager.intial
}

// Filter creates a [github.com/mr-linch/go-tg/tgb.Filter] based on Session data.
//
// Example:
//
//  isStepName := manager.Filter(func(session *Session) bool {
//    return session.Step == "name"
//  })
//
// This filter can be used in [github.com/mr-linch/go-tg/tgb.Router] handler registration method:
//
//  router.Message(func(ctx context.Context, mu *tgb.MessageUpdate) error {
//    // ...
//  }, isStepName)
func (manager *Manager[T]) Filter(fn func(*T) bool) tgb.Filter {
	return tgb.FilterFunc(func(ctx context.Context, update *tgb.Update) (bool, error) {
		session := manager.Get(ctx)

		if session == nil {
			return false, nil
		}

		return fn(session), nil
	})
}

// Wrap allow use manager as [github.com/mr-linch/go-tg/tgb.Middleware].
//
// This middleware do following:
//   1. fetch session data from [Store] or create new one if it doesn't exist.
//   2. put session data to [context.Context]
//   3. call handler (note: if chain returns error, return an error and do not save changes)
//   4. update session data in [Store] if it was changed (delete if session value equals initial)
func (manager *Manager[T]) Wrap(next tgb.Handler) tgb.Handler {
	return tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
		key := manager.keyFunc(update)

		if key == "" {
			return fmt.Errorf("can't get key from update")
		}

		session, err := manager.getSession(ctx, key)
		if err != nil {
			return fmt.Errorf("get session from store: %w", err)
		}

		// copy session before passing to next handler,
		// for compare in future
		sessionBeforeHandle := *session

		ctx = context.WithValue(ctx, sessionContextKey, session)

		if err := next.Handle(ctx, update); err != nil {
			return err
		}

		// check if session changed and should be updated
		if sessionBeforeHandle != *session {
			if *session == manager.intial {
				if err := manager.store.Del(ctx, key); err != nil {
					return fmt.Errorf("delete default session: %w", err)
				}
				return nil
			}
			if err := manager.saveSession(ctx, key, session); err != nil {
				return fmt.Errorf("save session to store: %w", err)
			}
		}

		return nil
	})
}

type contextKey int

const (
	sessionContextKey contextKey = iota
)
