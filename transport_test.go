package httptrace

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type fixtureTransport struct {
	Response *http.Response
	Err      error
	Request  *http.Request
}

func (c *fixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	c.Request = r
	return c.Response, c.Err
}

func TestTraceNoopIfNoParent(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var tracer = newMockTracer(ctrl)
	var originalTracer = opentracing.GlobalTracer()
	opentracing.InitGlobalTracer(tracer)
	defer opentracing.InitGlobalTracer(originalTracer)

	var resp = http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
	}
	var wrapped = NewTransport(
		TransportOptionPeerName("TESTPATH"),
		TransportOptionSpanName("TESTSPAN"),
	)(&fixtureTransport{Response: &resp, Err: nil})
	var req, _ = http.NewRequest(http.MethodGet, "/", nil)
	wrapped.RoundTrip(req)
}

func TestTraceNoopIfNoParentFailure(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var tracer = newMockTracer(ctrl)
	var originalTracer = opentracing.GlobalTracer()
	opentracing.InitGlobalTracer(tracer)
	defer opentracing.InitGlobalTracer(originalTracer)

	var resp = http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
	}
	var wrapped = NewTransport(
		TransportOptionPeerName("TESTPATH"),
		TransportOptionSpanName("TESTSPAN"),
	)(&fixtureTransport{Response: &resp, Err: fmt.Errorf("TESTERROR")})
	var req, _ = http.NewRequest(http.MethodGet, "/", nil)

	wrapped.RoundTrip(req)
}

func TestTraceAdoptsSpanIfExists(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var tracer = newMockTracer(ctrl)
	var originalTracer = opentracing.GlobalTracer()
	opentracing.InitGlobalTracer(tracer)
	defer opentracing.InitGlobalTracer(originalTracer)

	var resp = http.Response{
		Status:     "200 OK",
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
	}
	var parentSpan = newMockSpan(ctrl)
	var parentSpanContext = newMockSpanContext(ctrl)
	var childSpan = newMockSpan(ctrl)
	var childSpanContext = newMockSpanContext(ctrl)
	var ctx = opentracing.ContextWithSpan(context.Background(), parentSpan)
	var wrapped = NewTransport(
		TransportOptionPeerName("TESTPATH"),
		TransportOptionSpanName("TESTSPAN"),
	)(&fixtureTransport{Response: &resp, Err: nil}).(*Transport)
	var req, _ = http.NewRequest(http.MethodGet, "/", nil)

	parentSpan.EXPECT().Tracer().Return(tracer)
	parentSpan.EXPECT().Context().Return(parentSpanContext)
	tracer.EXPECT().StartSpan(wrapped.spanName, opentracing.ChildOf(parentSpanContext)).Return(childSpan)
	childSpan.EXPECT().SetTag(ext.SpanKindRPCClient.Key, ext.SpanKindRPCClient.Value)
	childSpan.EXPECT().SetTag(string(ext.HTTPMethod), http.MethodGet)
	childSpan.EXPECT().SetTag(string(ext.HTTPUrl), "/")
	childSpan.EXPECT().SetTag(string(ext.PeerService), "TESTPATH")
	childSpan.EXPECT().Tracer().Return(tracer)
	childSpan.EXPECT().Context().Return(childSpanContext)
	tracer.EXPECT().Inject(childSpanContext, opentracing.TextMap, gomock.Any())
	childSpan.EXPECT().SetTag(string(ext.HTTPStatusCode), uint16(http.StatusOK))
	childSpan.EXPECT().Finish()
	wrapped.RoundTrip(req.WithContext(ctx))
}

func TestTraceAdoptsSpanIfExistsFailure(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var tracer = newMockTracer(ctrl)
	var originalTracer = opentracing.GlobalTracer()
	opentracing.InitGlobalTracer(tracer)
	defer opentracing.InitGlobalTracer(originalTracer)

	var parentSpan = newMockSpan(ctrl)
	var parentSpanContext = newMockSpanContext(ctrl)
	var childSpan = newMockSpan(ctrl)
	var childSpanContext = newMockSpanContext(ctrl)
	var ctx = opentracing.ContextWithSpan(context.Background(), parentSpan)
	var wrapped = NewTransport(
		TransportOptionPeerName("TESTPATH"),
		TransportOptionSpanName("TESTSPAN"),
	)(&fixtureTransport{Response: nil, Err: fmt.Errorf("")}).(*Transport)
	var req, _ = http.NewRequest(http.MethodGet, "/", nil)

	parentSpan.EXPECT().Tracer().Return(tracer)
	parentSpan.EXPECT().Context().Return(parentSpanContext)
	tracer.EXPECT().StartSpan(wrapped.spanName, opentracing.ChildOf(parentSpanContext)).Return(childSpan)
	childSpan.EXPECT().SetTag(ext.SpanKindRPCClient.Key, ext.SpanKindRPCClient.Value)
	childSpan.EXPECT().SetTag(string(ext.HTTPMethod), http.MethodGet)
	childSpan.EXPECT().SetTag(string(ext.HTTPUrl), "/")
	childSpan.EXPECT().SetTag(string(ext.PeerService), "TESTPATH")
	childSpan.EXPECT().Tracer().Return(tracer)
	childSpan.EXPECT().Context().Return(childSpanContext)
	tracer.EXPECT().Inject(childSpanContext, opentracing.TextMap, gomock.Any())
	childSpan.EXPECT().SetTag(string(ext.Error), true)
	childSpan.EXPECT().Finish()
	wrapped.RoundTrip(req.WithContext(ctx))
}

func TestHeaderCarrierInject(t *testing.T) {
	var headers = http.Header{}
	var carrier = httpHeaderTextMapCarrier(headers)
	carrier.Set("TEST", "VALUE")

	if headers.Get("TEST") != "VALUE" {
		t.Fatal("carrier did not inject header values")
	}
}

func TestHeaderCarrierForEachSuccess(t *testing.T) {
	var headers = http.Header{}
	headers.Add("TEST", "VALUE")
	var carrier = httpHeaderTextMapCarrier(headers)
	carrier.ForeachKey(func(k string, v string) error {
		if strings.ToUpper(k) != "TEST" || strings.ToUpper(v) != "VALUE" {
			t.Fatalf("invalid foreach values %s:%s", k, v)
		}
		return nil
	})
}

func TestHeaderCarrierInvalidEncoding(t *testing.T) {
	var headers = http.Header{}
	headers.Add("TEST", "%ZZ")
	var carrier = httpHeaderTextMapCarrier(headers)
	carrier.ForeachKey(func(k string, v string) error {
		t.Fatalf("invalid foreach values %s:%s", k, v)
		return nil
	})
}

func TestHeaderCarrierError(t *testing.T) {
	var headers = http.Header{}
	headers.Add("TEST", "VALUE")
	var carrier = httpHeaderTextMapCarrier(headers)
	var e = carrier.ForeachKey(func(k string, v string) error {
		return errors.New("")
	})
	if e == nil {
		t.Fatal("did not propagate error values")
	}
}
