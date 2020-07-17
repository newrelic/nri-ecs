package ecs

import (
	"net/http"
	"time"
)

func ClientWithTimeout(timeout time.Duration) *http.Client {
	client := &http.Client{
		Timeout: timeout,
	}
	return client
}
