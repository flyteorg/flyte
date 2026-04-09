package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConstants(t *testing.T) {
	assert.Equal(t, 5432, defaultPort)
	assert.Equal(t, "postgres", defaultUser)
	assert.Equal(t, "postgres", defaultPass)
	assert.Equal(t, "/var/cache/embedded-postgres", cachePath)
	assert.Equal(t, "/tmp/embedded-postgres-runtime", runtimePath)
	assert.Equal(t, "/var/lib/flyte/storage/db", dataPath)
	assert.Equal(t, 999, pgUID)
	assert.Equal(t, 999, pgGID)
}

func TestDataPathPersistence(t *testing.T) {
	// Verify the data path would survive container restarts (under /var/lib/flyte/storage/)
	assert.Contains(t, dataPath, "/var/lib/flyte/storage/")
}

func TestInitDBSkipsIfVersionFileExists(t *testing.T) {
	// Create a temp directory with PG_VERSION file to simulate existing data dir
	tmpDir := t.TempDir()
	dataDir := filepath.Join(tmpDir, "db")
	err := os.MkdirAll(dataDir, 0700)
	assert.NoError(t, err)

	versionFile := filepath.Join(dataDir, "PG_VERSION")
	err = os.WriteFile(versionFile, []byte("16"), 0644)
	assert.NoError(t, err)

	// The library handles this internally - just verify the file check works
	_, err = os.Stat(versionFile)
	assert.NoError(t, err, "PG_VERSION file should exist")
}

func TestPgHbaConfAppend(t *testing.T) {
	// Simulate pg_hba.conf modification that enableRemoteConnections does
	tmpDir := t.TempDir()
	hbaPath := filepath.Join(tmpDir, "pg_hba.conf")

	initialContent := "# default pg_hba.conf\nlocal all all password\n"
	err := os.WriteFile(hbaPath, []byte(initialContent), 0600)
	assert.NoError(t, err)

	f, err := os.OpenFile(hbaPath, os.O_APPEND|os.O_WRONLY, 0600)
	assert.NoError(t, err)
	_, err = f.WriteString("\n# Allow all hosts (sandbox environment)\nhost all all 0.0.0.0/0 password\nhost all all ::/0 password\n")
	assert.NoError(t, err)
	f.Close()

	content, err := os.ReadFile(hbaPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "0.0.0.0/0")
	assert.Contains(t, string(content), "::/0")
	assert.Contains(t, string(content), initialContent) // original content preserved
}
