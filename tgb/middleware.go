package tgb

type Middleware interface {
	Wrap(Handler) Handler
}

type MiddlewareFunc func(Handler) Handler

func (m MiddlewareFunc) Wrap(h Handler) Handler {
	return m(h)
}

type chain []Middleware

// Append extends a chain, adding the specified middleware
// as the last ones in the request flow.
func (c chain) Append(mws ...Middleware) chain {
	result := make(chain, 0, len(c)+len(mws))
	result = append(result, c...)
	result = append(result, mws...)
	return result
}

// Then wraps handler with middleware chain.
func (c chain) Then(handler Handler) Handler {
	for i := range c {
		handler = c[len(c)-1-i].Wrap(handler)
	}

	return handler
}
