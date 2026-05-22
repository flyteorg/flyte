package api

import (
	"fmt"
	"net/http"
	"time"
)

func uniqueString() string { //nolint: unused
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func newClient() *http.Client { //nolint: unused
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}
