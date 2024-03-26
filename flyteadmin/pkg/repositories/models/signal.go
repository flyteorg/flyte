package models

// Signal primary key
type SignalKey struct {
	ExecutionKey
	SignalID string `gorm:"primary_key;index" valid:"length(0|255)"`
}

// Database model to encapsulate a signal.
type Signal struct {
	BaseModel
	SignalKey
	Type  []byte `gorm:"not null"`
	Value []byte
}

var SignalColumns = modelColumns(Signal{})
