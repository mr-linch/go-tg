package tg

// Logger defines generic interface for loggers
type Logger interface {
	Printf(format string, args ...any)
}
