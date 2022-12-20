package interfaces

// Exposes the common methods required for a subscriber.
// There is one ProcessNotification per type.
type Processor interface {

	// Starts processing messages from the underlying subscriber.
	// If the channel closes gracefully, no error will be returned.
	// If the underlying channel experiences errors,
	// an error is returned and the channel is closed.
	StartProcessing()

	// This should be invoked when the application is shutting down.
	// If StartProcessing() returned an error, StopProcessing() will return an error because
	// the channel was already closed.
	StopProcessing() error
}
