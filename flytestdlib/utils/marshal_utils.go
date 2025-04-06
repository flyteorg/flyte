package utils

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

// UnmarshalStructToPb unmarshals a proto struct into a proto message using jsonPb marshaler.
func UnmarshalStructToPb(structObj *structpb.Struct, msg proto.Message) error {
	if structObj == nil {
		return fmt.Errorf("nil Struct object passed")
	}

	if msg == nil {
		return fmt.Errorf("nil proto.Message object passed")
	}

	jsonObj, err := protojson.Marshal(structObj)
	if err != nil {
		return errors.WithMessage(err, "Failed to marshal strcutObj input")
	}

	if err = protojson.Unmarshal(jsonObj, msg); err != nil {
		return errors.WithMessage(err, "Failed to unmarshal json obj into proto")
	}

	return nil
}

// MarshalPbToStruct marshals a proto message into proto Struct using jsonPb marshaler.
func MarshalPbToStruct(in proto.Message) (out *structpb.Struct, err error) {
	if in == nil {
		return nil, fmt.Errorf("nil proto message passed")
	}

	b, err := protojson.Marshal(in)
	if err != nil {
		return nil, errors.WithMessage(err, "Failed to marshal input proto message")
	}

	out = &structpb.Struct{}
	if err = protojson.Unmarshal(b, out); err != nil {
		return nil, errors.WithMessage(err, "Failed to unmarshal json object into struct")
	}

	return out, nil
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
	if err := protojson.Unmarshal(b, structObj); err != nil {
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
