package interceptorstest

import "context"

// TestUnaryHandler is an implementation of grpc.UnaryHandler for test purposes
type TestUnaryHandler struct {
	Err             error
	handleCallCount int
	capturedCtx     context.Context
	HandleFunc      func(ctx context.Context)
}

func (h *TestUnaryHandler) Handle(ctx context.Context, req interface{}) (interface{}, error) {
	h.handleCallCount++
	h.capturedCtx = ctx

	if h.HandleFunc != nil {
		h.HandleFunc(ctx)
	}

	if h.Err != nil {
		return nil, h.Err
	}

	return nil, nil
}

// GetHandleCallCount gets the number of times the handle method was called
func (h *TestUnaryHandler) GetHandleCallCount() int {
	return h.handleCallCount
}

// GetCapturedCtx gets the context captured during the last handle method call
func (h *TestUnaryHandler) GetCapturedCtx() context.Context {
	return h.capturedCtx
}
