package session

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

// WithStorage sets a storage for sessions.
// By default, it uses [StoreMemory].
func WithStorage(storage Store) ManagerOption {
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
