package utils

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/pkg/errors"
)

var jsonPbMarshaler = jsonpb.Marshaler{}

// UnmarshalStructToPb unmarshals a proto struct into a proto message using jsonPb marshaler.
func UnmarshalStructToPb(structObj *structpb.Struct, msg proto.Message) error {
	if structObj == nil {
		return fmt.Errorf("nil Struct object passed")
	}

	if msg == nil {
		return fmt.Errorf("nil proto.Message object passed")
	}

	jsonObj, err := jsonPbMarshaler.MarshalToString(structObj)
	if err != nil {
		return errors.WithMessage(err, "Failed to marshal strcutObj input")
	}

	if err = jsonpb.UnmarshalString(jsonObj, msg); err != nil {
		return errors.WithMessage(err, "Failed to unmarshal json obj into proto")
	}

	return nil
}

// MarshalPbToStruct marshals a proto message into proto Struct using jsonPb marshaler.
func MarshalPbToStruct(in proto.Message) (out *structpb.Struct, err error) {
	if in == nil {
		return nil, fmt.Errorf("nil proto message passed")
	}

	var buf bytes.Buffer
	if err := jsonPbMarshaler.Marshal(&buf, in); err != nil {
		return nil, errors.WithMessage(err, "Failed to marshal input proto message")
	}

	out = &structpb.Struct{}
	if err = jsonpb.Unmarshal(bytes.NewReader(buf.Bytes()), out); err != nil {
		return nil, errors.WithMessage(err, "Failed to unmarshal json object into struct")
	}

	return out, nil
}

// MarshalPbToString marshals a proto message using jsonPb marshaler to string.
func MarshalPbToString(msg proto.Message) (string, error) {
	return jsonPbMarshaler.MarshalToString(msg)
}

// MarshalObjToStruct marshals obj into a struct. Will use jsonPb if input is a proto message, otherwise, it'll use json
// marshaler.
func MarshalObjToStruct(input interface{}) (*structpb.Struct, error) {
	if p, casted := input.(proto.Message); casted {
		return MarshalPbToStruct(p)
	}

	b, err := json.Marshal(input)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to marshal input proto message")
	}

	// Turn JSON into a protobuf struct
	structObj := &structpb.Struct{}
	if err := jsonpb.Unmarshal(bytes.NewReader(b), structObj); err != nil {
		return nil, errors.WithMessage(err, "Failed to unmarshal json object into struct")
	}

	return structObj, nil
}

// UnmarshalStructToObj unmarshals a struct to the passed obj. Don't use this if the unmarshalled obj is a proto message.
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
