//go:build e2e
// +build e2e

// End-to-end tests for the multi-scheme DataStore. Excluded from normal builds by the `e2e` build
// tag (matching the convention in executor/test/e2e).
//
// REQUIRES A RUNNING DEVBOX: these tests dial the devbox's RustFS (S3) backend on
// http://localhost:30002. Start it first from the repo root with `make devbox-run`, then run them
// with `make test-e2e` (from flytestdlib/) or:
//
//	go test -tags=e2e ./storage/... -run RustFS -v
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

// TestRoutingStore_RustFSLazyDial is a devbox e2e test for the multi-scheme DataStore. It
// builds a DataStore whose PRIMARY scheme is the in-memory store and whose SECONDARY s3:// scheme is
// lazily dialed against the devbox's RustFS via the new Schemes override. This is the one path a
// normal devbox run can't exercise (everything in a run is the primary s3:// scheme), so we drive it
// directly here.
//
// The RustFS endpoint defaults to the devbox NodePort (http://localhost:30002); override it (and the
// bucket/credentials) via the FLYTE_RUSTFS_* env vars if your devbox differs.
func TestRoutingStore_RustFSLazyDial(t *testing.T) {
	endpoint := getenv("FLYTE_RUSTFS_ENDPOINT", "http://localhost:30002")
	bucket := getenv("FLYTE_RUSTFS_BUCKET", "flyte-data")
	s3Config := map[string]string{
		"auth_type":     "accesskey",
		"access_key_id": getenv("FLYTE_AWS_ACCESS_KEY_ID", "rustfs"),
		"secret_key":    getenv("FLYTE_AWS_SECRET_ACCESS_KEY", "rustfsstorage"),
		"region":        getenv("FLYTE_RUSTFS_REGION", "us-east-1"),
		"endpoint":      endpoint,
		"disable_ssl":   "true",
	}

	// Primary scheme is "mem"; s3:// is a SECONDARY scheme dialed lazily from the Schemes override.
	cfg := &Config{
		Type:          TypeMemory,
		InitContainer: "primary-mem",
		Schemes: map[string]SchemeConfig{
			"s3": {Kind: "s3", Config: s3Config},
		},
	}

	ds, err := NewDataStore(cfg, promutils.NewTestScope())
	require.NoError(t, err)

	ctx := context.Background()
	payload := []byte("multi-scheme routing store hello")
	s3Ref := DataReference(fmt.Sprintf("s3://%s/routing-test/obj.bin", bucket))

	// 1. Write through the SECONDARY s3 scheme. This forces the routing store to lazily dial RustFS
	//    via stowFactory on first reference — the new code path under test.
	require.NoError(t, ds.WriteRaw(ctx, s3Ref, int64(len(payload)), Options{}, bytes.NewReader(payload)))

	// 2. Read it back from RustFS to confirm the lazily-dialed backend round-trips real bytes.
	rc, err := ds.ReadRaw(ctx, s3Ref)
	require.NoError(t, err)
	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.NoError(t, rc.Close())
	assert.Equal(t, payload, got)

	// 3. Cross-scheme copy: s3 (RustFS) -> mem (primary). Exercises copyImpl routing through the
	//    store itself across two different backends.
	memRef := DataReference("mem://primary-mem/routing-test/copy.bin")
	require.NoError(t, ds.CopyRaw(ctx, s3Ref, memRef, Options{}))

	rc, err = ds.ReadRaw(ctx, memRef)
	require.NoError(t, err)
	got, err = io.ReadAll(rc)
	require.NoError(t, err)
	require.NoError(t, rc.Close())
	assert.Equal(t, payload, got, "cross-scheme copy s3->mem should preserve content")

	// 4. Clean up the RustFS object so reruns start fresh.
	assert.NoError(t, ds.Delete(ctx, s3Ref))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
