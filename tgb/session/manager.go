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

func NewManager[T comparable](initial T, opts ...ManagerOption) *Manager[T] {
	manager := &Manager[T]{
		intial: initial,

		keyFunc:    KeyFuncChat,
		store:      NewStoreMemory(),
		encodeFunc: json.Marshal,
		decodeFunc: json.Unmarshal,
	}

	for _, opt := range opts {
		opt(manager)
	}

	return manager
}

func (manager *Manager[T]) Init(opt ManagerOption, opts ...ManagerOption) {
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

func (manager *Manager[T]) Get(ctx context.Context) *T {
	v := ctx.Value(sessionContextKey)
	if v == nil {
		return nil
	}
	return v.(*T)
}

func (manager *Manager[T]) Filter(fn func(*T) bool) tgb.Filter {
	return tgb.FilterFunc(func(ctx context.Context, update *tgb.Update) (bool, error) {
		session := manager.Get(ctx)

		if session == nil {
			return false, nil
		}

		return fn(session), nil
	})
}

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
