package stridetrace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"bitbucket.org/atlassian/logevent"
	opentracing "github.com/opentracing/opentracing-go"
)

type fixtureHandler struct {
	called bool
	ctx    context.Context
	span   opentracing.Span
}

func (h *fixtureHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.called = true
	h.ctx = r.Context()
}

func TestMiddlewareGeneratesNewSpans(t *testing.T) {
	var w = httptest.NewRecorder()
	var r, _ = http.NewRequest("GET", "/", nil)
	var emitted bool
	var logFunc = func(ctx context.Context, level logevent.LogLevel, message string, annotations map[string]interface{}) {
		emitted = true
		var traceData, ok = annotations["zipkin"].(jsonSpan)
		if !ok {
			t.Fatal("could not fetch trace data")
		}
		if traceData.ParentID != "" {
			t.Fatalf("expected no parent span but found %s", traceData.ParentID)
		}
	}
	var fromContext = logevent.NewFromContextFunc(logFunc)
	var wrapped = fixtureHandler{}
	var handler = NewMiddleware(
		MiddlewareOptionLogFetcher(fromContext),
		MiddlewareOptionServiceName("testservice"),
		MiddlewareOptionHostPort("localhost:8080"),
	)(&wrapped)
	handler.ServeHTTP(w, r)

	if !wrapped.called {
		t.Fatal("middleware did not call the wrapped handler")
	}
	if !emitted {
		t.Fatal("no spans emitted")
	}
}

func TestMiddlewareGeneratesAdoptsHeaders(t *testing.T) {
	var w = httptest.NewRecorder()
	var r, _ = http.NewRequest("GET", "/", nil)
	var emitted bool
	var logFunc = func(ctx context.Context, level logevent.LogLevel, message string, annotations map[string]interface{}) {
		emitted = true
		var traceData, ok = annotations["zipkin"].(jsonSpan)
		if !ok {
			t.Fatal("could not fetch trace data")
		}
		if traceData.ParentID != "0000000000000002" {
			t.Fatalf("expected parent 0000000000000002 but found %s", traceData.ParentID)
		}
		if traceData.TraceID != "0000000000000001" {
			t.Fatalf("expected trace 0000000000000001 but found %s", traceData.TraceID)
		}
	}
	var fromContext = logevent.NewFromContextFunc(logFunc)
	var wrapped = fixtureHandler{}
	var handler = NewMiddleware(
		MiddlewareOptionLogFetcher(fromContext),
		MiddlewareOptionServiceName("testservice"),
		MiddlewareOptionHostPort("localhost:8080"),
	)(&wrapped)
	r.Header.Set("X-B3-TraceId", "0000000000000001")
	r.Header.Set("X-B3-SpanId", "0000000000000002")
	r.Header.Set("X-B3-ParentSpanId", "0000000000000003")
	r.Header.Set("X-B3-Sampled", "1")
	handler.ServeHTTP(w, r)

	if !wrapped.called {
		t.Fatal("middleware did not call the wrapped handler")
	}
	if !emitted {
		t.Fatal("no spans emitted")
	}
}
