package client

import (
	"context"
	"strings"

	"github.com/flyteorg/flytestdlib/logger"
)

// This type is meant only to encapsulate the response coming from Qubole as a type, it is
// not meant to be stored locally.
type QuboleStatus string

const (
	QuboleStatusUnknown   QuboleStatus = "UNKNOWN"
	QuboleStatusWaiting   QuboleStatus = "WAITING"
	QuboleStatusRunning   QuboleStatus = "RUNNING"
	QuboleStatusDone      QuboleStatus = "DONE"
	QuboleStatusError     QuboleStatus = "ERROR"
	QuboleStatusCancelled QuboleStatus = "CANCELLED"
)

var QuboleStatuses = map[QuboleStatus]struct{}{
	QuboleStatusUnknown:   {},
	QuboleStatusWaiting:   {},
	QuboleStatusRunning:   {},
	QuboleStatusDone:      {},
	QuboleStatusError:     {},
	QuboleStatusCancelled: {},
}

func NewQuboleStatus(ctx context.Context, status string) QuboleStatus {
	upperCased := strings.ToUpper(status)
	if _, ok := QuboleStatuses[QuboleStatus(upperCased)]; ok {
		return QuboleStatus(upperCased)
	}
	logger.Warnf(ctx, "Invalid Qubole Status found: %v", status)
	return QuboleStatusUnknown
}
