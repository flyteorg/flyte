package validators

import (
	"fmt"

	"github.com/flyteorg/datacatalog/pkg/errors"
	datacatalog "github.com/flyteorg/datacatalog/protos/gen"
	"google.golang.org/grpc/codes"
)

const (
	partitionKeyName   = "partitionKey"
	partitionValueName = "partitionValue"
)

func ValidatePartitions(datasetPartitionKeys []string, artifactPartitions []*datacatalog.Partition) error {
	if len(datasetPartitionKeys) != len(artifactPartitions) {
		return errors.NewDataCatalogErrorf(codes.InvalidArgument, "Partition key mismatch, dataset keys: %+v, artifact Partitions: %+v", datasetPartitionKeys, artifactPartitions)
	}

	// Not all datasets need to be partitioned
	if len(datasetPartitionKeys) == 0 {
		return nil
	}

	// compare the contents of the datasetkeys and artifact keys
	partitionErrors := make([]error, 0)
	keyMismatch := false

	partitionKeyMatches := make(map[string]bool, len(artifactPartitions))
	for _, datasetPartitionKey := range datasetPartitionKeys {
		partitionKeyMatches[datasetPartitionKey] = false
	}

	for idx, artifactPartition := range artifactPartitions {
		if artifactPartition == nil {
			partitionErrors = append(partitionErrors, NewMissingArgumentError(fmt.Sprintf("%v[%v]", partitionKeyName, idx)))
			continue
		}

		if err := ValidateEmptyStringField(partitionKeyName, artifactPartition.Key); err != nil {
			partitionErrors = append(partitionErrors, NewMissingArgumentError(fmt.Sprintf("%v[%v]", partitionKeyName, idx)))
		} else if err := ValidateEmptyStringField(partitionValueName, artifactPartition.Value); err != nil {
			partitionErrors = append(partitionErrors, NewMissingArgumentError(fmt.Sprintf("%v[%v]", partitionValueName, idx)))
		} else {
			_, ok := partitionKeyMatches[artifactPartition.Key]

			if ok {
				partitionKeyMatches[artifactPartition.Key] = true
			} else {
				keyMismatch = true
			}
		}
	}

	if keyMismatch {
		partitionErrors = append(partitionErrors, errors.NewDataCatalogErrorf(codes.InvalidArgument, "Artifact partition assignment does not match dataset partition keys: %v", partitionKeyMatches))
	}

	if len(partitionErrors) > 0 {
		return errors.NewCollectedErrors(codes.InvalidArgument, partitionErrors)
	}

	return nil
}

// Validate that the partition keys are unique strings
func ValidateUniquePartitionKeys(partitionKeys []string) error {
	invalidPartitionKeys := false
	partitionKeySet := make(map[string]uint8, len(partitionKeys))
	for _, partitionKey := range partitionKeys {
		partitionKeySet[partitionKey]++
		if partitionKeySet[partitionKey] > 1 {
			invalidPartitionKeys = true
		}
	}

	if invalidPartitionKeys {
		return NewInvalidArgumentError(partitionKeyName, fmt.Sprintf("Keys are not unique, occurrence count: %+v", partitionKeySet))
	}

	return nil
}
