package client

// This type is meant only to encapsulate the response coming from Presto as a type, it is
// not meant to be stored locally.

//go:generate enumer --type=PrestoStatus

type PrestoStatus int

const (
	PrestoStatusUnknown PrestoStatus = iota
	PrestoStatusWaiting
	PrestoStatusRunning
	PrestoStatusFinished
	PrestoStatusFailed
	PrestoStatusCancelled
)
