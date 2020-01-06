package audit

import "time"

type Principal struct {
	// Identifies authenticated end-user
	Subject string

	// The client that initiated the auth flow.
	ClientID string

	TokenIssuedAt time.Time
}

type Client struct {
	ClientIP string
}

type AccessMode int

const (
	ReadOnly AccessMode = iota
	ReadWrite
)

// Details about a specific request issued by a user.
type Request struct {
	// Service method endpoint e.g. GetWorkflowExecution
	Method string

	// Includes parameters submitted in the request.
	Parameters map[string]string

	Mode AccessMode

	ReceivedAt time.Time
}

// Summary of service response details.
type Response struct {
	// e.g. gRPC status code
	ResponseCode string

	SentAt time.Time
}

type Message struct {
	Principal Principal

	Client Client

	Request Request

	Response Response
}
