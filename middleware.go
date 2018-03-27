package stridetrace

import (
	"context"
	"fmt"
	"net/http"

	"bitbucket.org/atlassian/logevent"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

type key string

var (
	traceCtxKey = key("stridetrace-trace")
	spanCtxKey  = key("stridetrace-span")
)

// Middleware adds zipkin style request tracing.
type Middleware struct {
	wrapped     http.Handler
	serviceName string
	hostPort    string
	fromContext func(context.Context) logevent.Logger
}

func (h *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var collector = &collector{h.fromContext(r.Context())}
	var recorder = zipkin.NewRecorder(collector, false, h.hostPort, h.serviceName)
	var tracer, err = zipkin.NewTracer(recorder)
	if err != nil {
		h.wrapped.ServeHTTP(w, r)
		return
	}
	var ctx = r.Context()
	var wireContext, er = tracer.Extract(opentracing.TextMap, opentracing.HTTPHeadersCarrier(r.Header))
	if er != nil {
		var span = tracer.StartSpan(h.serviceName)
		defer span.Finish()
		ctx = context.WithValue(ctx, traceCtxKey, span.Context().(zipkin.SpanContext).TraceID.Low)
		ctx = context.WithValue(ctx, spanCtxKey, span.Context().(zipkin.SpanContext).SpanID)
		ctx = opentracing.ContextWithSpan(ctx, span)
		h.wrapped.ServeHTTP(w, r.WithContext(ctx))
		return
	}
	var span = tracer.StartSpan(h.serviceName, opentracing.ChildOf(wireContext))
	defer span.Finish()
	ctx = context.WithValue(ctx, traceCtxKey, span.Context().(zipkin.SpanContext).TraceID.Low)
	ctx = context.WithValue(ctx, spanCtxKey, span.Context().(zipkin.SpanContext).SpanID)
	ctx = opentracing.ContextWithSpan(ctx, span)
	h.wrapped.ServeHTTP(w, r.WithContext(ctx))
}

// MiddlewareOption is a configuration setting for the HTTP middleware.
type MiddlewareOption func(*Middleware) *Middleware

// MiddlewareOptionServiceName sets the service name annotation of the
// spans associated with the incoming request. The default value of this
// option is HTTPService.
func MiddlewareOptionServiceName(name string) MiddlewareOption {
	return func(m *Middleware) *Middleware {
		m.serviceName = name
		return m
	}
}

// MiddlewareOptionHostPort sets host:port annotation used to represent the
// service in spans associated with the incoming request. The default value of
// this option is 0.0.0.0:80.
func MiddlewareOptionHostPort(hostPort string) MiddlewareOption {
	return func(m *Middleware) *Middleware {
		m.hostPort = hostPort
		return m
	}
}

// MiddlewareOptionLogFetcher installs a custom function that will be used to
// extract the logevent.Logger from the context. The default value of this
// option is the logevent.FromContext method which leverages the default
// logging implementation. This setting must given if using a custom logging
// implementation with logevent.
func MiddlewareOptionLogFetcher(f func(context.Context) logevent.Logger) MiddlewareOption {
	return func(m *Middleware) *Middleware {
		m.fromContext = f
		return m
	}
}

// NewMiddleware creates a middleware.
func NewMiddleware(options ...MiddlewareOption) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		var middleware = &Middleware{
			serviceName: "HTTPService",
			fromContext: logevent.FromContext,
			wrapped:     next,
		}
		for _, option := range options {
			middleware = option(middleware)
		}
		return middleware
	}
}

// TraceIDFromContext returns the active TraceID value as a string.
func TraceIDFromContext(ctx context.Context) string {
	return fmt.Sprintf("%016x", ctx.Value(traceCtxKey))
}

// SpanIDFromContext returns the active TraceID value as a string.
func SpanIDFromContext(ctx context.Context) string {
	return fmt.Sprintf("%016x", ctx.Value(spanCtxKey))
}

// OutOfBand returns a context with all of the configuration provided to the
// middleware. This is provided with the primary intent of allowing for trace
// emissions during runtime setup (such as main.go) and background routines that
// are not attached to a request or request context.
func OutOfBand(ctx context.Context, middleware func(http.Handler) http.Handler) context.Context {
	var m *Middleware
	var ok bool
	if m, ok = middleware(nil).(*Middleware); !ok {
		return opentracing.ContextWithSpan(ctx, opentracing.GlobalTracer().StartSpan("background"))
	}
	var collector = &collector{m.fromContext(ctx)}
	var recorder = zipkin.NewRecorder(collector, false, "80", m.serviceName)
	var tracer, err = zipkin.NewTracer(recorder)
	if err != nil {
		return opentracing.ContextWithSpan(ctx, opentracing.GlobalTracer().StartSpan("background"))
	}
	return opentracing.ContextWithSpan(ctx, tracer.StartSpan("background"))
}
