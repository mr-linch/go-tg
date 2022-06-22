package tgb

type Logger interface {
	Printf(format string, args ...any)
}
