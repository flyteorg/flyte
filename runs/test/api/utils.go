package api

import (
	"fmt"
	"net/http"
	"time"
)

func uniqueString() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func newClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}
