package requests

import (
	"net/http"
	"time"
)

func NewClient() *http.Client {
	c := &http.Client{
		// TODO: Flag/config for timeout
		Timeout: time.Second * 5,
	}

	return c
}
