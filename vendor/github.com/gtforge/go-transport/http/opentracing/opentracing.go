package opentracing

import (
	"net/http"

	"github.com/gtforge/go-transport/http/httpwares"
	opentracing "github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

var httpTag = opentracing.Tag{Key: string(otext.Component), Value: "http"}

// Middleware wraps http handler with opentracing middleware
var Middleware = func(opts ...Option) httpwares.Middleware {
	// join slices in this order to make possible options rewriting
	o := evaluateOptions(append([]Option{WithFilterFunc(skipAliveTracing)}, opts...))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !o.filterFunc(r) {
				next.ServeHTTP(w, r)
				return
			}

			r, span := newServerSpanFromRequest(r, o.tracer)
			defer span.Finish()

			rw := responseWriter{ResponseWriter: w}

			next.ServeHTTP(&rw, r)

			otext.HTTPStatusCode.Set(span, uint16(rw.StatusCode()))
			if o.statusCodeErrorFunc(rw.StatusCode()) {
				otext.Error.Set(span, true)
			}
		})
	}
}

// Tripperware wraps http transport with url sanitizing service name detector and opentracing middleware
var Tripperware = func(opts ...Option) httpwares.RoundTripperMiddleware {
	o := evaluateOptions(opts)

	return func(next http.RoundTripper) http.RoundTripper {
		if next == nil {
			next = http.DefaultTransport
		}

		return roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			if !o.filterFunc(r) {
				return next.RoundTrip(r)
			}

			r, span := newClientSpanFromRequest(r, o.tracer)
			defer span.Finish()

			res, err := next.RoundTrip(r)
			if err != nil {
				otext.Error.Set(span, true)
				span.LogFields(otlog.String("event", "error"), otlog.String("message", err.Error()))
			} else {
				otext.HTTPStatusCode.Set(span, uint16(res.StatusCode))
				if o.statusCodeErrorFunc(res.StatusCode) {
					otext.Error.Set(span, true)
				}
			}

			return res, err
		})
	}
}

// roundTripperFunc http.RoundTripper adapter for using in a functional manner
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// responseWriter http.ResponseWriter wrapper fore storing status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseWriter) StatusCode() int {
	if r.statusCode == 0 {
		return http.StatusOK
	}
	return r.statusCode
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func newClientSpanFromRequest(r *http.Request, tracer opentracing.Tracer) (*http.Request, opentracing.Span) {
	var parentSpanContext opentracing.SpanContext
	if parent := opentracing.SpanFromContext(r.Context()); parent != nil {
		parentSpanContext = parent.Context()
	}

	span := tracer.StartSpan(
		clientName(r),
		opentracing.ChildOf(parentSpanContext),
		otext.SpanKindRPCClient,
		httpTag,
	)
	otext.HTTPUrl.Set(span, r.URL.String())
	otext.HTTPMethod.Set(span, r.Method)
	otext.PeerHostname.Set(span, hostName(r))

	r = r.WithContext(opentracing.ContextWithSpan(r.Context(), span))
	_ = tracer.Inject(span.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))

	return r, span
}

func newServerSpanFromRequest(r *http.Request, tracer opentracing.Tracer) (*http.Request, opentracing.Span) {
	spanContext, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
	span := tracer.StartSpan(serverName(r), otext.RPCServerOption(spanContext), httpTag)
	otext.HTTPMethod.Set(span, r.Method)
	otext.HTTPUrl.Set(span, r.URL.String())
	otext.PeerHostname.Set(span, hostName(r))
	return r.WithContext(opentracing.ContextWithSpan(r.Context(), span)), span
}

// skipAliveTracing is a default filterFunc for server
func skipAliveTracing(r *http.Request) bool {
	return r.URL.Path != "/alive"
}

// serverName creates span name for server
// i.e. http.server GET service.gett.com
func serverName(r *http.Request) string {
	return spanName(serverSpanPrefix, r.Method, hostName(r))
}

// clientName creates span name for client
// i.e. http.client GET service.gett.com
func clientName(r *http.Request) string {
	return spanName(clientSpanPrefix, r.Method, hostName(r))
}

// hostName common function that returns hostname from request
func hostName(r *http.Request) string {
	if r.URL.Host != "" {
		return r.URL.Host
	}
	return r.Host
}

// spanName common function for span name creation
func spanName(prefix, method, host string) string {
	return prefix + " " + method + " " + host
}

const (
	serverSpanPrefix = "http.server"
	clientSpanPrefix = "http.client"
)
