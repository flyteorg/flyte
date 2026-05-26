package utils

import (
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

var jsonPbMarshaler = protojson.MarshalOptions{}
var jsonPbUnmarshaler = protojson.UnmarshalOptions{
	DiscardUnknown: true,
}

// Deprecated: Use flytestdlib/utils.UnmarshalStructToPb instead.
func UnmarshalStruct(structObj *structpb.Struct, msg proto.Message) error {
	if structObj == nil {
		return fmt.Errorf("nil Struct Object passed")
	}

	jsonObj, err := jsonPbMarshaler.Marshal(structObj)
	if err != nil {
		return err
	}

	if err = jsonPbUnmarshaler.Unmarshal(jsonObj, msg); err != nil {
		return err
	}

	return nil
}

// Deprecated: Use flytestdlib/utils.MarshalPbToStruct instead.
func MarshalStruct(in proto.Message, out *structpb.Struct) error {
	if out == nil {
		return fmt.Errorf("nil Struct Object passed")
	}

	jsonObj, err := jsonPbMarshaler.Marshal(in)
	if err != nil {
		return err
	}

	if err = jsonPbUnmarshaler.Unmarshal(jsonObj, out); err != nil {
		return err
	}

	return nil
}

// Deprecated: Use flytestdlib/utils.MarshalToString instead.
func MarshalToString(msg proto.Message) (string, error) {
	b, err := jsonPbMarshaler.Marshal(msg)
	return string(b), err
}

// Deprecated: Use flytestdlib/utils.MarshalObjToStruct instead.
// Don't use this if input is a proto Message.
func MarshalObjToStruct(input interface{}) (*structpb.Struct, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	// Turn JSON into a protobuf struct
	structObj := &structpb.Struct{}
	if err := jsonPbUnmarshaler.Unmarshal(b, structObj); err != nil {
		return nil, err
	}
	return structObj, nil
}

// Deprecated: Use flytestdlib/utils.UnmarshalStructToObj instead.
// Don't use this if the unmarshalled obj is a proto message.
func UnmarshalStructToObj(structObj *structpb.Struct, obj interface{}) error {
	if structObj == nil {
		return fmt.Errorf("nil Struct Object passed")
	}

	jsonObj, err := json.Marshal(structObj)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(jsonObj, obj); err != nil {
		return err
	}

	return nil
}
