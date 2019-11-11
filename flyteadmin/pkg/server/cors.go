package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
)

const (
	// At time of writing, the version argument is required to be 1.
	// See https://github.com/grpc-ecosystem/grpc-gateway/blob/69669120b0e010b88291cd9d24761297c2a17582/runtime/pattern.go#L53
	GrpcGatewayPatternVersion = 1
)

// This is a function used for CORS support. It produces a Pattern object that can be attached to the grpc-gateway
// ServeMux object. It should match any and all URLs. The two op codes say, push the entire path to the stack 'OpPushM',
// and then ignore the result, 'OpNop'.
func GetGlobPattern() runtime.Pattern {
	return runtime.MustPattern(runtime.NewPattern(GrpcGatewayPatternVersion, []int{int(utilities.OpPushM), int(utilities.OpNop)},
		[]string{}, ""))
}

// This is just a type alias because the grpc-gateway library did not define one.
type ForwardResponseOptionHandler func(i context.Context, writer http.ResponseWriter, message proto.Message) error

// This returns a handler to be used with the runtime.WithForwardResponseOption() function. This is middleware that will add
// the CORS header below to all requests. Certain requests like GET will not trigger a preflight request and instead the
// browser expects this header to be present in the response.
func GetForwardResponseOptionHandler(allowedOrigins []string) ForwardResponseOptionHandler {
	return func(i context.Context, writer http.ResponseWriter, message proto.Message) error {
		writer.Header().Set("Access-Control-Allow-Origin", strings.Join(allowedOrigins, ", "))
		writer.Header().Set("Access-Control-Allow-Credentials", "true")
		return nil
	}
}
