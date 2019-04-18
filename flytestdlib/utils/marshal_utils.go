package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/struct"
	"github.com/pkg/errors"
)

var jsonPbMarshaler = jsonpb.Marshaler{}

// Unmarshals a proto struct into a proto message using jsonPb marshaler.
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

// Marshals a proto message into proto Struct using jsonPb marshaler.
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

// Marshals a proto message using jsonPb marshaler to string.
func MarshalPbToString(msg proto.Message) (string, error) {
	return jsonPbMarshaler.MarshalToString(msg)
}

// Marshals obj into a struct. Will use jsonPb if input is a proto message, otherwise, it'll use json
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
