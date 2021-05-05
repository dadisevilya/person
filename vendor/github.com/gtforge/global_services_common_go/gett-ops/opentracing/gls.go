package opentracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/gtforge/gls"
)

const opentracingSpanGLSKey = "opentracingSpanGLSKey"

var noopTracer opentracing.NoopTracer

func SpanFromGLS() opentracing.Span {
	if span, ok := gls.Get(opentracingSpanGLSKey).(opentracing.Span); ok {
		return span
	}
	return noopTracer.StartSpan("")
}

func SpanToGLS(span opentracing.Span) {
	gls.Set(opentracingSpanGLSKey, span)
}
