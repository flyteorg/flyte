package utils

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto" //nolint: staticcheck
	structpb "google.golang.org/protobuf/types/known/structpb"
)

var jsonPbMarshaler = protojson.MarshalOptions{}
var jsonPbUnmarshaler = protojson.UnmarshalOptions{
	DiscardUnknown: true,
}

// UnmarshalStructToPb unmarshals a proto struct into a proto message using jsonPb marshaler.
func UnmarshalStructToPb(structObj *structpb.Struct, msg proto.Message) error {
	if structObj == nil {
		return fmt.Errorf("nil Struct object passed")
	}

	if msg == nil {
		return fmt.Errorf("nil proto.Message object passed")
	}

	jsonObj, err := jsonPbMarshaler.Marshal(structObj)
	if err != nil {
		return errors.WithMessage(err, "Failed to marshal strcutObj input")
	}

	if err = UnmarshalBytesToPb(jsonObj, msg); err != nil {
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
	b, err := jsonPbMarshaler.Marshal(in)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to marshal input proto message")
	}
	buf.Write(b)

	out = &structpb.Struct{}
	if err = UnmarshalBytesToPb(buf.Bytes(), out); err != nil {
		return nil, errors.WithMessage(err, "Failed to unmarshal json object into struct")
	}

	return out, nil
}

// MarshalPbToString marshals a proto message using jsonPb marshaler to string.
func MarshalPbToString(msg proto.Message) (string, error) {
	if msg == nil {
		return "", fmt.Errorf("nil proto message passed")
	}

	b, err := jsonPbMarshaler.Marshal(msg)
	return string(b), err
}

// UnmarshalStringToPb unmarshals a string to a proto message
func UnmarshalStringToPb(s string, msg proto.Message) error {
	return jsonPbUnmarshaler.Unmarshal([]byte(s), msg)
}

// MarshalPbToBytes marshals a proto message to a byte slice
func MarshalPbToBytes(msg proto.Message) ([]byte, error) {
	if msg == nil {
		return nil, fmt.Errorf("nil proto message passed")
	}

	var buf bytes.Buffer
	b, err := jsonPbMarshaler.Marshal(msg)
	if err != nil {
		return nil, err
	}
	buf.Write(b)
	return buf.Bytes(), nil
}

// UnmarshalBytesToPb unmarshals a byte slice to a proto message
func UnmarshalBytesToPb(b []byte, msg proto.Message) error {
	return jsonPbUnmarshaler.Unmarshal(b, msg)
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
	if err := UnmarshalBytesToPb(b, structObj); err != nil {
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
