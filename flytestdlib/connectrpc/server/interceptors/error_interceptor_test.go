package interceptors

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect"
)

func TestErrorInterceptorWrapUnary(t *testing.T) {
	t.Parallel()

	ordinaryErr := errors.New("database unavailable")
	connectErr := connect.NewError(connect.CodeNotFound, errors.New("not found"))
	wrappedConnectErr := fmt.Errorf("lookup failed: %w", connectErr)

	tests := []struct {
		name         string
		handlerErr   error
		wantErr      error
		wantCode     connect.Code
		wantResponse bool
	}{
		{
			name:         "success",
			wantCode:     connect.CodeUnknown,
			wantResponse: true,
		},
		{
			name:       "ordinary error becomes internal",
			handlerErr: ordinaryErr,
			wantCode:   connect.CodeInternal,
		},
		{
			name:       "Connect error is preserved",
			handlerErr: connectErr,
			wantErr:    connectErr,
			wantCode:   connect.CodeNotFound,
		},
		{
			name:       "wrapped Connect error is preserved",
			handlerErr: wrappedConnectErr,
			wantErr:    wrappedConnectErr,
			wantCode:   connect.CodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			response := connect.NewResponse(new(string))
			handler := NewErrorInterceptor().WrapUnary(
				func(context.Context, connect.AnyRequest) (connect.AnyResponse, error) {
					if tt.handlerErr != nil {
						return nil, tt.handlerErr
					}
					return response, nil
				},
			)

			gotResponse, gotErr := handler(context.Background(), connect.NewRequest(new(string)))
			if tt.wantResponse && gotResponse != response {
				t.Fatalf("response = %v, want original response %v", gotResponse, response)
			}
			if tt.handlerErr == nil {
				if gotErr != nil {
					t.Fatalf("error = %v, want nil", gotErr)
				}
				return
			}
			if tt.wantErr != nil && !errors.Is(gotErr, tt.wantErr) {
				t.Fatalf("error = %v, want it to contain %v", gotErr, tt.wantErr)
			}
			if tt.wantErr == nil && errors.Is(gotErr, tt.handlerErr) {
				t.Fatalf("error = %v, must not expose original error %v", gotErr, tt.handlerErr)
			}
			if gotCode := connect.CodeOf(gotErr); gotCode != tt.wantCode {
				t.Errorf("error code = %v, want %v", gotCode, tt.wantCode)
			}
			if tt.wantCode == connect.CodeInternal {
				var connectErr *connect.Error
				if !errors.As(gotErr, &connectErr) {
					t.Fatalf("error type = %T, want *connect.Error", gotErr)
				}
				if message := connectErr.Message(); message != "" {
					t.Errorf("error message = %q, want empty", message)
				}
			}
		})
	}
}

func TestErrorInterceptorWrapStreamingHandler(t *testing.T) {
	t.Parallel()

	ordinaryErr := errors.New("stream failed")
	connectErr := connect.NewError(connect.CodeInvalidArgument, errors.New("invalid message"))
	wrappedConnectErr := fmt.Errorf("receive failed: %w", connectErr)

	tests := []struct {
		name       string
		handlerErr error
		wantErr    error
		wantCode   connect.Code
	}{
		{name: "success"},
		{
			name:       "ordinary error becomes internal",
			handlerErr: ordinaryErr,
			wantCode:   connect.CodeInternal,
		},
		{
			name:       "Connect error is preserved",
			handlerErr: connectErr,
			wantErr:    connectErr,
			wantCode:   connect.CodeInvalidArgument,
		},
		{
			name:       "wrapped Connect error is preserved",
			handlerErr: wrappedConnectErr,
			wantErr:    wrappedConnectErr,
			wantCode:   connect.CodeInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			handler := NewErrorInterceptor().WrapStreamingHandler(
				func(context.Context, connect.StreamingHandlerConn) error {
					return tt.handlerErr
				},
			)

			gotErr := handler(context.Background(), nil)
			if tt.handlerErr == nil {
				if gotErr != nil {
					t.Fatalf("error = %v, want nil", gotErr)
				}
				return
			}
			if tt.wantErr != nil && !errors.Is(gotErr, tt.wantErr) {
				t.Fatalf("error = %v, want it to contain %v", gotErr, tt.wantErr)
			}
			if tt.wantErr == nil && errors.Is(gotErr, tt.handlerErr) {
				t.Fatalf("error = %v, must not expose original error %v", gotErr, tt.handlerErr)
			}
			if gotCode := connect.CodeOf(gotErr); gotCode != tt.wantCode {
				t.Errorf("error code = %v, want %v", gotCode, tt.wantCode)
			}
			if tt.wantCode == connect.CodeInternal {
				var connectErr *connect.Error
				if !errors.As(gotErr, &connectErr) {
					t.Fatalf("error type = %T, want *connect.Error", gotErr)
				}
				if message := connectErr.Message(); message != "" {
					t.Errorf("error message = %q, want empty", message)
				}
			}
		})
	}
}
