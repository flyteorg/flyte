package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
)

var jsonPbMarshaler = jsonpb.Marshaler{}
var jsonPbUnmarshaler = &jsonpb.Unmarshaler{
	AllowUnknownFields: true,
}

func UnmarshalStruct(structObj *structpb.Struct, msg proto.Message) error {
	if structObj == nil {
		return fmt.Errorf("nil Struct Object passed")
	}

	jsonObj, err := jsonPbMarshaler.MarshalToString(structObj)
	if err != nil {
		return err
	}

	if err = jsonPbUnmarshaler.Unmarshal(strings.NewReader(jsonObj), msg); err != nil {
		return err
	}

	return nil
}

func MarshalStruct(in proto.Message, out *structpb.Struct) error {
	if out == nil {
		return fmt.Errorf("nil Struct Object passed")
	}

	jsonObj, err := jsonPbMarshaler.MarshalToString(in)
	if err != nil {
		return err
	}

	if err = jsonpb.UnmarshalString(jsonObj, out); err != nil {
		return err
	}

	return nil
}

func MarshalToString(msg proto.Message) (string, error) {
	return jsonPbMarshaler.MarshalToString(msg)
}

// TODO: Use the stdlib version in the future, or move there if not there.
// Don't use this if input is a proto Message.
func MarshalObjToStruct(input interface{}) (*structpb.Struct, error) {
	b, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	// Turn JSON into a protobuf struct
	structObj := &structpb.Struct{}
	if err := jsonpb.UnmarshalString(string(b), structObj); err != nil {
		return nil, err
	}
	return structObj, nil
}

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
