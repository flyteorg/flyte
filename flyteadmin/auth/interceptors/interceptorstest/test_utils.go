package interceptorstest

import "context"

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

func (h *TestUnaryHandler) GetHandleCallCount() int {
	return h.handleCallCount
}

func (h *TestUnaryHandler) GetCapturedCtx() context.Context {
	return h.capturedCtx
}
