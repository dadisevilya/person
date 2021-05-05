package opentracing

import (
	goTransport "github.com/gtforge/go-transport/http/opentracing"
	"github.com/opentracing/opentracing-go"
	"net/http"
)

func OpentracingMiddleware(tracer opentracing.Tracer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		trace := goTransport.Middleware(goTransport.WithTracer(tracer))
		return trace(openTracingGLSMiddleware(next))
	}
}

func openTracingGLSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if span := opentracing.SpanFromContext(r.Context()); span != nil {
			SpanToGLS(span)
		}
		next.ServeHTTP(w, r)
	})
}
