package api

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func uniqueString() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func newClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}
