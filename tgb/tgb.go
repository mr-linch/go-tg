// Package tgb is a Telegram bot framework.
// It's contains high level API to easily create Telegram bots.
package tgb

func firstNotNil[T any](fields ...*T) *T {
	for _, field := range fields {
		if field != nil {
			return field
		}
	}
	return nil
}
