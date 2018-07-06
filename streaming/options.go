package streaming

type DialOption func(*connection)

// WithUpdateCallback returns an options which sets a callback function for
// streaming updates. Each update will first be applied to the order book, and
// then passed to the callback function.
func WithUpdateCallback(fn UpdateCallback) DialOption {
	return func(c *connection) {
		c.MessageProcessor.updateCallback = fn
	}
}
