package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	_ "github.com/lib/pq"
)

const (
	defaultPort = 5432
	defaultUser = "postgres"
	defaultPass = "postgres"
	cachePath   = "/var/cache/embedded-postgres"
	runtimePath = "/tmp/embedded-postgres-runtime"
	dataPath    = "/var/lib/flyte/storage/db"
	pgUID       = 999
	pgGID       = 999
)

func main() {
	log.Println("Starting embedded PostgreSQL...")

	// Prepare directories as root before dropping privileges.
	// The library calls os.RemoveAll(dataPath) on start, which requires
	// write permission on the parent directory, so we chown parents too.
	for _, dir := range []string{dataPath, runtimePath, cachePath} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		// Chown both the directory and its parent (needed for RemoveAll)
		for _, d := range []string{dir, filepath.Dir(dir)} {
			if err := os.Chown(d, pgUID, pgGID); err != nil {
				log.Fatalf("Failed to chown directory %s: %v", d, err)
			}
		}
	}

	// Drop privileges to postgres user (PostgreSQL refuses to run as root)
	if os.Getuid() == 0 {
		if err := syscall.Setgid(pgGID); err != nil {
			log.Fatalf("Failed to setgid: %v", err)
		}
		if err := syscall.Setuid(pgUID); err != nil {
			log.Fatalf("Failed to setuid: %v", err)
		}
		log.Printf("Dropped privileges to uid=%d gid=%d", os.Getuid(), os.Getgid())
	}

	db := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(uint32(defaultPort)).
			Username(defaultUser).
			Password(defaultPass).
			Database("postgres").
			DataPath(dataPath).
			RuntimePath(runtimePath).
			CachePath(cachePath).
			StartParameters(map[string]string{
				"max_connections":  "200",
				"listen_addresses": "*",
			}).
			Version(embeddedpostgres.V16),
	)

	if err := db.Start(); err != nil {
		log.Fatalf("Failed to start embedded PostgreSQL: %v", err)
	}
	log.Printf("Embedded PostgreSQL started on port %d", defaultPort)

	// Allow connections from all hosts (needed for K8s pods to connect)
	if err := enableRemoteConnections(); err != nil {
		log.Fatalf("Failed to configure pg_hba.conf: %v", err)
	}

	// Create additional databases needed by Flyte
	if err := createDatabases(); err != nil {
		log.Fatalf("Failed to create databases: %v", err)
	}

	// Signal readiness
	if err := os.WriteFile("/tmp/embedded-postgres-ready", []byte("ready"), 0644); err != nil {
		log.Printf("Warning: failed to write ready file: %v", err)
	}
	log.Println("Embedded PostgreSQL is ready")

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh

	log.Printf("Received signal %v, shutting down embedded PostgreSQL...", sig)
	if err := db.Stop(); err != nil {
		log.Printf("Warning: failed to stop embedded PostgreSQL cleanly: %v", err)
	}
	log.Println("Embedded PostgreSQL stopped")
}

func enableRemoteConnections() error {
	hbaPath := filepath.Join(dataPath, "pg_hba.conf")
	f, err := os.OpenFile(hbaPath, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open pg_hba.conf: %w", err)
	}
	defer f.Close()
	if _, err := f.WriteString("\n# Allow all hosts (sandbox environment)\nhost all all 0.0.0.0/0 password\nhost all all ::/0 password\n"); err != nil {
		return fmt.Errorf("failed to write pg_hba.conf: %w", err)
	}

	// Reload PostgreSQL configuration
	connStr := fmt.Sprintf(
		"host=127.0.0.1 port=%d user=%s password=%s dbname=postgres sslmode=disable",
		defaultPort, defaultUser, defaultPass,
	)
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect for reload: %w", err)
	}
	defer conn.Close()
	if _, err := conn.Exec("SELECT pg_reload_conf()"); err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}
	log.Println("Configured pg_hba.conf for remote connections")
	return nil
}

func createDatabases() error {
	connStr := fmt.Sprintf(
		"host=127.0.0.1 port=%d user=%s password=%s dbname=postgres sslmode=disable",
		defaultPort, defaultUser, defaultPass,
	)
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer conn.Close()

	for _, dbName := range []string{"flyte", "runs"} {
		var exists bool
		err := conn.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check database %s: %w", dbName, err)
		}
		if !exists {
			if _, err := conn.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)); err != nil {
				return fmt.Errorf("failed to create database %s: %w", dbName, err)
			}
			log.Printf("Created database: %s", dbName)
		} else {
			log.Printf("Database already exists: %s", dbName)
		}
	}
	return nil
}
