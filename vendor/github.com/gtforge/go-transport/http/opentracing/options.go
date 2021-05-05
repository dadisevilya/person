package opentracing

import (
	"net/http"

	opentracing "github.com/opentracing/opentracing-go"
)

var defaultOptions = options{
	filterFunc:          defaultFilterFunc,
	statusCodeErrorFunc: defaultStatusCodeIsError,
	tracer:              opentracing.GlobalTracer(),
}

// FilterFunc allows users to provide a function that filters out certain methods from being traced.
// If it returns false, the given request will not be traced.
type FilterFunc func(req *http.Request) bool

// StatusCodeIsError allows the customization of which requests are considered errors in the tracing system.
type StatusCodeIsError func(statusCode int) bool

type options struct {
	filterFunc          FilterFunc
	statusCodeErrorFunc StatusCodeIsError
	tracer              opentracing.Tracer
}

func evaluateOptions(opts []Option) *options {
	evaluatedOptions := defaultOptions
	for _, o := range opts {
		o(&evaluatedOptions)
	}
	return &evaluatedOptions
}

type Option func(*options)

// WithFilterFunc customizes the function used for deciding whether a given call is traced or not.
func WithFilterFunc(f FilterFunc) Option {
	return func(o *options) {
		if f != nil {
			o.filterFunc = f
		}
	}
}

// WithStatusCodeIsError customizes the function used for deciding whether a given call was an error
func WithStatusCodeIsError(f StatusCodeIsError) Option {
	return func(o *options) {
		if f != nil {
			o.statusCodeErrorFunc = f
		}
	}
}

// WithTracer sets a custom tracer to be used for this middleware, otherwise the opentracing.GlobalTracer is used.
func WithTracer(tracer opentracing.Tracer) Option {
	return func(o *options) {
		if tracer != nil {
			o.tracer = tracer
		}
	}
}

func defaultStatusCodeIsError(statusCode int) bool {
	switch statusCode / 100 {
	case 1, 2, 3:
		return false
	default: // 4, 5
		return true
	}
}

func defaultFilterFunc(_ *http.Request) bool {
	return true // trace all requests by default
}
