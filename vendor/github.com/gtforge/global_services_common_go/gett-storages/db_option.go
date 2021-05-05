package gettStorages

import opentracing "github.com/opentracing/opentracing-go"

type DBOption func(o *dbOptions)

func DBWithTracer(tracer opentracing.Tracer) DBOption {
	return func(o *dbOptions) {
		o.tracer = tracer
	}
}

func dbOptionsApply(opts ...DBOption) dbOptions {
	dbo := defaultDBOptions
	for _, o := range opts {
		o(&dbo)
	}
	return dbo
}

type dbOptions struct {
	tracer opentracing.Tracer
}

var defaultDBOptions = dbOptions{
	tracer: opentracing.GlobalTracer(),
}
