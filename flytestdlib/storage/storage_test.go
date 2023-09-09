package storage

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/stretchr/testify/assert"
)

func TestDataReference_Split(t *testing.T) {
	input := DataReference("s3://container/path/to/file")
	scheme, container, key, err := input.Split()

	assert.NoError(t, err)
	assert.Equal(t, "s3", scheme)
	assert.Equal(t, "container", container)
	assert.Equal(t, "path/to/file", key)
}

func ExampleNewDataStore() {
	testScope := promutils.NewTestScope()
	ctx := context.Background()
	fmt.Println("Creating in memory data store.")
	store, err := NewDataStore(&Config{
		Type: TypeMemory,
	}, testScope.NewSubScope("exp_new"))

	if err != nil {
		fmt.Printf("Failed to create data store. Error: %v", err)
	}

	ref, err := store.ConstructReference(ctx, DataReference("root"), "subkey", "subkey2")
	if err != nil {
		fmt.Printf("Failed to construct data reference. Error: %v", err)
	}

	fmt.Printf("Constructed data reference [%v] and writing data to it.\n", ref)

	dataToStore := "hello world"
	err = store.WriteRaw(ctx, ref, int64(len(dataToStore)), Options{}, strings.NewReader(dataToStore))
	if err != nil {
		fmt.Printf("Failed to write data. Error: %v", err)
	}

	// Output:
	// Creating in memory data store.
	// Constructed data reference [/root/subkey/subkey2] and writing data to it.
}
