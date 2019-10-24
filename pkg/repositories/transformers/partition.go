package transformers

import "github.com/lyft/datacatalog/pkg/repositories/models"

func FromPartitionKeyModel(partitionKeys []models.PartitionKey) []string {
	partitionKeyStrings := make([]string, len(partitionKeys))
	for i, partitionKey := range partitionKeys {
		partitionKeyStrings[i] = partitionKey.Name
	}

	return partitionKeyStrings
}
