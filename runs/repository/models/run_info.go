package models

import (
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// MarshalRunInfo serializes a RunInfo proto to bytes for storage in the DB.
func MarshalRunInfo(info *workflow.RunInfo) ([]byte, error) {
	return proto.Marshal(info)
}

// UnmarshalRunInfo deserializes bytes from the DB into a RunInfo proto.
func UnmarshalRunInfo(data []byte) (*workflow.RunInfo, error) {
	if len(data) == 0 {
		return nil, nil
	}
	info := &workflow.RunInfo{}
	if err := proto.Unmarshal(data, info); err != nil {
		return nil, err
	}
	return info, nil
}
