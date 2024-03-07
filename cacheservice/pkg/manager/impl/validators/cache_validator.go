package validators

import (
	"github.com/flyteorg/flyte/cacheservice/pkg/errors"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/cacheservice"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func validateCacheKey(key string) error {
	if key == "" {
		return errors.NewMissingArgumentError("key")
	}

	return nil
}

func validateOutputURI(outputURI string) error {
	if outputURI == "" {
		return errors.NewMissingArgumentError("output_uri")
	}

	return nil
}

func validateOutputLiteral(literal *core.LiteralMap) error {
	if literal == nil {
		return errors.NewMissingArgumentError("literal")
	}

	return nil
}

func validateOutputMetadata(metadata *cacheservice.Metadata) error {
	if metadata.GetSourceIdentifier() == nil {
		return errors.NewMissingArgumentError("source_identifier")
	}

	return nil
}

func validateOutput(cachedOutput *cacheservice.CachedOutput) error {
	if cachedOutput == nil {
		return errors.NewMissingArgumentError("output")
	}

	if cachedOutput.Output == nil {
		return errors.NewInvalidArgumentError("output", "")
	}

	err := validateOutputMetadata(cachedOutput.Metadata)
	if err != nil {
		return err
	}

	switch output := cachedOutput.Output.(type) {
	case *cacheservice.CachedOutput_OutputLiterals:
		err = validateOutputLiteral(output.OutputLiterals)
		if err != nil {
			return err
		}
	case *cacheservice.CachedOutput_OutputUri:
		err = validateOutputURI(output.OutputUri)
		if err != nil {
			return err
		}
	default:
		return errors.NewInvalidArgumentError("output", "unknown type")
	}

	return nil
}

func validateOwnerID(ownerID string) error {
	if ownerID == "" {
		return errors.NewMissingArgumentError("owner_id")
	}

	return nil
}

func ValidatePutCacheRequest(request *cacheservice.PutCacheRequest) error {
	if request == nil {
		return errors.NewMissingArgumentError("request")
	}

	err := validateCacheKey(request.Key)
	if err != nil {
		return err
	}

	err = validateOutput(request.Output)
	if err != nil {
		return err
	}

	return nil
}

func ValidateGetCacheRequest(request *cacheservice.GetCacheRequest) error {
	if request == nil {
		return errors.NewMissingArgumentError("request")
	}

	err := validateCacheKey(request.Key)
	if err != nil {
		return err
	}

	return nil
}

func ValidateGetOrExtendReservationRequest(request *cacheservice.GetOrExtendReservationRequest) error {
	if request == nil {
		return errors.NewMissingArgumentError("request")
	}

	err := validateCacheKey(request.Key)
	if err != nil {
		return err
	}

	err = validateOwnerID(request.OwnerId)
	if err != nil {
		return err
	}

	return nil
}

func ValidateReleaseReservationRequest(request *cacheservice.ReleaseReservationRequest) error {
	if request == nil {
		return errors.NewMissingArgumentError("request")
	}

	err := validateCacheKey(request.Key)
	if err != nil {
		return err
	}

	err = validateOwnerID(request.OwnerId)
	if err != nil {
		return err
	}

	return nil
}
