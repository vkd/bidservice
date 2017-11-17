package main

import (
	"net/http"
)

// HTTPSender - interface for *http.Client
type HTTPSender interface {
	Do(req *http.Request) (*http.Response, error)
}

var _ HTTPSender = (*http.Client)(nil)

// HTTPSenderFunc - func implement HTTPSender
type HTTPSenderFunc func(req *http.Request) (*http.Response, error)

var _ HTTPSender = (HTTPSenderFunc)(nil)

// Do - implement HTTPSender
func (f HTTPSenderFunc) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}
