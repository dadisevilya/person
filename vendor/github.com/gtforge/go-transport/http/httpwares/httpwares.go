package httpwares

import "net/http"

// This package defines types for wrapping and decorating http.RoundTripper and http.Handler.
// This types is suggested to use in github.com/gtforge/go-transport/http

// Middleware http.Handler wrapper
type Middleware func(handler http.Handler) http.Handler

// RoundTripperFunc implements http.RoundTripper interface for working with round tripper in functional style
type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// RoundTripperMiddleware wraps http client transport
type RoundTripperMiddleware func(http.RoundTripper) http.RoundTripper
