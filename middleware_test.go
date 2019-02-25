package httptrace

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asecurityteam/logevent"
	"github.com/golang/mock/gomock"
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
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var w = httptest.NewRecorder()
	var r, _ = http.NewRequest("GET", "/", nil)

	var logger = NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any()).Do(func(event interface{}) {
		var evt frame
		var ok bool
		if evt, ok = event.(frame); !ok {
			t.Error("did not log a zipkin frame")
		}
		if evt.Zipkin.ParentID != "" {
			t.Errorf("unexpected parent span %s", evt.Zipkin.ParentID)
		}
	})
	var wrapped = fixtureHandler{}
	var handler = NewMiddleware(
		MiddlewareOptionServiceName("testservice"),
		MiddlewareOptionHostPort("localhost:8080"),
	)(&wrapped)
	handler.ServeHTTP(w, r.WithContext(logevent.NewContext(r.Context(), logger)))

	if !wrapped.called {
		t.Error("middleware did not call the wrapped handler")
	}
}

func TestMiddlewareGeneratesAdoptsHeaders(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var w = httptest.NewRecorder()
	var r, _ = http.NewRequest("GET", "/", nil)

	var logger = NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any()).Do(func(event interface{}) {
		var evt frame
		var ok bool
		if evt, ok = event.(frame); !ok {
			t.Error("did not log a zipkin frame")
		}
		if evt.Zipkin.ParentID != "0000000000000002" {
			t.Errorf("expected parent 0000000000000002 but found %s", evt.Zipkin.ParentID)
		}
		if evt.Zipkin.TraceID != "0000000000000001" {
			t.Errorf("expected trace 0000000000000001 but found %s", evt.Zipkin.TraceID)
		}
	})
	var wrapped = fixtureHandler{}
	var handler = NewMiddleware(
		MiddlewareOptionServiceName("testservice"),
		MiddlewareOptionHostPort("localhost:8080"),
	)(&wrapped)
	r.Header.Set("X-B3-TraceId", "0000000000000001")
	r.Header.Set("X-B3-SpanId", "0000000000000002")
	r.Header.Set("X-B3-ParentSpanId", "0000000000000003")
	r.Header.Set("X-B3-Sampled", "1")
	handler.ServeHTTP(w, r.WithContext(logevent.NewContext(r.Context(), logger)))

	if !wrapped.called {
		t.Error("middleware did not call the wrapped handler")
	}
}
