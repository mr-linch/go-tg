package session

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/mr-linch/go-tg"
	"github.com/mr-linch/go-tg/tgb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type StoreMock struct {
	mock.Mock
}

var _ Store = (*StoreMock)(nil)

func (store *StoreMock) Set(ctx context.Context, key string, data []byte) error {
	args := store.Called(ctx, key, data)
	return args.Error(0)
}

func (store *StoreMock) Get(ctx context.Context, key string) ([]byte, error) {
	args := store.Called(ctx, key)
	data := args.Get(0)

	if data == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]byte), args.Error(1)
}

func (store *StoreMock) Del(ctx context.Context, key string) error {
	args := store.Called(ctx, key)
	return args.Error(0)
}

func TestNewManager(t *testing.T) {
	type Session struct {
		Count int
	}

	keyFunc := KeyFunc(func(update *tgb.Update) string {
		return "key"
	})

	store := NewStoreMemory()

	encodeFunc := func(v any) ([]byte, error) {
		return []byte("encode"), nil
	}

	decodeFunc := func(data []byte, v any) error {
		return errors.New("decode")
	}

	defaultSession := Session{}

	manager := NewManager(
		defaultSession,
		WithKeyFunc(keyFunc),
		WithEncoding(encodeFunc, decodeFunc),
		WithStore(store),
	)

	require.NotNil(t, manager)
	require.Equal(t, store, manager.store)
	require.Equal(t, "key", manager.keyFunc(nil))

	v, err := manager.encodeFunc(defaultSession)
	require.NoError(t, err)
	require.Equal(t, []byte("encode"), v)

	err = manager.decodeFunc([]byte("decode"), defaultSession)
	require.Error(t, err)
}

func TestManager_Init(t *testing.T) {
	type Session struct {
		Count int
	}

	keyFunc := KeyFunc(func(update *tgb.Update) string {
		return "key"
	})

	store := NewStoreMemory()

	encodeFunc := func(v any) ([]byte, error) {
		return []byte("encode"), nil
	}

	decodeFunc := func(data []byte, v any) error {
		return errors.New("decode")
	}

	defaultSession := Session{}

	manager := NewManager(defaultSession)

	manager.Init(
		WithKeyFunc(keyFunc),
		WithEncoding(encodeFunc, decodeFunc),
		WithStore(store),
	)

	require.NotNil(t, manager)
	require.Equal(t, store, manager.store)
	require.Equal(t, "key", manager.keyFunc(nil))

	v, err := manager.encodeFunc(defaultSession)
	require.NoError(t, err)
	require.Equal(t, []byte("encode"), v)

	err = manager.decodeFunc([]byte("decode"), defaultSession)
	require.Error(t, err)
}

func TestManager_Wrap(t *testing.T) {
	t.Run("CantGetKey", func(t *testing.T) {
		type session struct{}

		manager := NewManager(session{},
			WithKeyFunc(func(update *tgb.Update) string { return "" }),
		)

		handler := manager.Wrap(tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
			return nil
		}))

		err := handler.Handle(context.Background(), nil)

		require.EqualError(t, err, "can't get key from update")
	})

	t.Run("CantGetSession", func(t *testing.T) {
		type session struct{}

		store := &StoreMock{}

		store.On("Get",
			mock.Anything,
			"key",
		).Return(nil, errors.New("can't get session"))

		manager := NewManager(session{},
			WithKeyFunc(func(update *tgb.Update) string { return "key" }),
			WithStore(store),
		)

		handler := manager.Wrap(tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
			return nil
		}))

		err := handler.Handle(context.Background(), &tgb.Update{})

		require.EqualError(t, err, "get session from store: can't get session")
	})

	t.Run("HandleAndSave", func(t *testing.T) {
		type Session struct {
			Counter int
		}

		store := &StoreMock{}

		store.On("Get",
			mock.Anything,
			"1",
		).Return(nil, nil)

		store.On("Set",
			mock.Anything,
			"1",
			[]byte(`{"Counter":2}`),
		).Return(nil)

		manager := NewManager(
			Session{Counter: 1},
			WithStore(store),
		)

		handler := manager.Wrap(tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
			session := manager.Get(ctx)
			session.Counter++

			require.Equal(t, 2, session.Counter)

			return nil
		}))

		err := manager.Wrap(handler).Handle(context.Background(), &tgb.Update{Update: &tg.Update{
			Message: &tg.Message{
				Chat: tg.Chat{
					ID: 1,
				},
			},
		}})

		require.NoError(t, err)

		store.AssertExpectations(t)

		require.NoError(t, err)
	})

	t.Run("HandleErrorNotSave", func(t *testing.T) {
		type Session struct {
			Counter int
		}

		store := &StoreMock{}

		store.On("Get",
			mock.Anything,
			"1",
		).Return(nil, nil)

		manager := NewManager(
			Session{Counter: 1},
			WithStore(store),
		)

		handler := manager.Wrap(tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
			session := manager.Get(ctx)
			session.Counter++

			require.Equal(t, 2, session.Counter)

			return fmt.Errorf("error")
		}))

		err := manager.Wrap(handler).Handle(context.Background(), &tgb.Update{Update: &tg.Update{
			Message: &tg.Message{
				Chat: tg.Chat{
					ID: 1,
				},
			},
		}})

		store.AssertExpectations(t)

		require.EqualError(t, err, "error")
	})

	t.Run("HandleAndSaveError", func(t *testing.T) {
		type Session struct {
			Counter int
		}

		store := &StoreMock{}

		store.On("Get",
			mock.Anything,
			"1",
		).Return(nil, nil)

		store.On("Set",
			mock.Anything,
			"1",
			[]byte(`{"Counter":2}`),
		).Return(errors.New("error"))

		manager := NewManager(
			Session{Counter: 1},
			WithStore(store),
		)

		handler := manager.Wrap(tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
			session := manager.Get(ctx)
			session.Counter++

			require.Equal(t, 2, session.Counter)

			return nil
		}))

		err := manager.Wrap(handler).Handle(context.Background(), &tgb.Update{Update: &tg.Update{
			Message: &tg.Message{
				Chat: tg.Chat{
					ID: 1,
				},
			},
		}})

		store.AssertExpectations(t)

		require.EqualError(t, err, "save session to store: error")
	})

	t.Run("HandleAndDelete", func(t *testing.T) {
		type Session struct {
			Counter int
		}

		store := &StoreMock{}

		store.On("Get",
			mock.Anything,
			"1",
		).Return([]byte(`{"Counter": 2}`), nil)

		store.On("Del",
			mock.Anything,
			"1",
		).Return(nil)

		manager := NewManager(
			Session{Counter: 1},
			WithStore(store),
		)

		handler := manager.Wrap(tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
			session := manager.Get(ctx)

			session.Counter = 1

			return nil
		}))

		err := manager.Wrap(handler).Handle(context.Background(), &tgb.Update{Update: &tg.Update{
			Message: &tg.Message{
				Chat: tg.Chat{
					ID: 1,
				},
			},
		}})

		require.NoError(t, err)

		store.AssertExpectations(t)
	})

	t.Run("HandleAndDeleteViaReset", func(t *testing.T) {
		type Session struct {
			Counter int
		}

		store := &StoreMock{}

		store.On("Get",
			mock.Anything,
			"1",
		).Return([]byte(`{"Counter": 2}`), nil)

		store.On("Del",
			mock.Anything,
			"1",
		).Return(nil)

		manager := NewManager(
			Session{Counter: 1},
			WithStore(store),
		)

		handler := manager.Wrap(tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
			session := manager.Get(ctx)

			manager.Reset(session)

			return nil
		}))

		err := manager.Wrap(handler).Handle(context.Background(), &tgb.Update{Update: &tg.Update{
			Message: &tg.Message{
				Chat: tg.Chat{
					ID: 1,
				},
			},
		}})

		require.NoError(t, err)

		store.AssertExpectations(t)
	})

	t.Run("HandleAndDeleteError", func(t *testing.T) {
		type Session struct {
			Counter int
		}

		store := &StoreMock{}

		store.On("Get",
			mock.Anything,
			"1",
		).Return([]byte(`{"Counter": 2}`), nil)

		store.On("Del",
			mock.Anything,
			"1",
		).Return(errors.New("error"))

		manager := NewManager(
			Session{Counter: 1},
			WithStore(store),
		)

		handler := manager.Wrap(tgb.HandlerFunc(func(ctx context.Context, update *tgb.Update) error {
			session := manager.Get(ctx)

			session.Counter = 1

			return nil
		}))

		err := manager.Wrap(handler).Handle(context.Background(), &tgb.Update{Update: &tg.Update{
			Message: &tg.Message{
				Chat: tg.Chat{
					ID: 1,
				},
			},
		}})

		require.EqualError(t, err, "delete default session: error")

		store.AssertExpectations(t)
	})
}

func TestManager_Get(t *testing.T) {
	ctx := context.Background()

	manager := NewManager(struct{}{})

	session := manager.Get(ctx)

	assert.Nil(t, session)
}

func TestKeyFuncChat(t *testing.T) {
	key := KeyFuncChat(&tgb.Update{Update: &tg.Update{
		Message: &tg.Message{
			Chat: tg.Chat{
				ID: 1,
			},
		},
	}})

	assert.Equal(t, "1", key)

	key = KeyFuncChat(&tgb.Update{})

	assert.Equal(t, "", key)
}
