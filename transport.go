package httptrace

import (
	"net/http"
	"net/url"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// httpHeaderMapCarrier satisfies both TextMapWriter and TextMapReader.
// This is a copy of the opentracing object with the same name. However, instead
// of using the 'Add' method for the header object it uses the 'Set' method
// to ensure multiple calls don't append onto each other.
type httpHeaderTextMapCarrier http.Header

// Set conforms to the TextMapWriter interface.
func (c httpHeaderTextMapCarrier) Set(key, val string) {
	var h = http.Header(c)
	h.Set(key, url.QueryEscape(val))
}

// ForeachKey conforms to the TextMapReader interface.
func (c httpHeaderTextMapCarrier) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range c {
		for _, v := range vals {
			var rawV, err = url.QueryUnescape(v)
			if err != nil {
				// We don't know if there was an error escaping an
				// OpenTracing-related header or something else; as such, we
				// continue rather than return the error.
				continue
			}
			if err = handler(k, rawV); err != nil {
				return err
			}
		}
	}
	return nil
}

// Transport adds zipkin style request tracing headers to outgoing requests.
type Transport struct {
	wrapped   http.RoundTripper
	spanName  string
	peerNamer func(*http.Request) string
}

// RoundTrip injects zipkin B3 headers into outgoing requests.
func (c *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	var parent = opentracing.SpanFromContext(r.Context())
	if parent == nil {
		return c.wrapped.RoundTrip(r)
	}
	var span = parent.Tracer().StartSpan(c.spanName, opentracing.ChildOf(parent.Context()))
	defer span.Finish()
	ext.SpanKindRPCClient.Set(span)
	ext.HTTPMethod.Set(span, r.Method)
	ext.HTTPUrl.Set(span, r.URL.Path)
	ext.PeerService.Set(span, c.peerNamer(r))
	_ = span.Tracer().Inject(span.Context(), opentracing.TextMap, httpHeaderTextMapCarrier(r.Header))
	var resp, er = c.wrapped.RoundTrip(r)
	if resp != nil {
		ext.HTTPStatusCode.Set(span, uint16(resp.StatusCode))
	}
	if er != nil {
		ext.Error.Set(span, true)
	}
	return resp, er
}

// TransportOption is a configuration setting for the Transport wrapper.
type TransportOption func(*Transport) *Transport

// TransportOptionSpanName sets the name of the zipkin span for the outgoing
// request. The default value for this is OutgoingHTTPRequest.
func TransportOptionSpanName(name string) TransportOption {
	return func(t *Transport) *Transport {
		t.spanName = name
		return t
	}
}

// TransportOptionPeerName sets the name of the remote peer being called. This
// setting establishes a single name for the peer being used in all outgoing
// calls. The default value for peers is dependency.
func TransportOptionPeerName(name string) TransportOption {
	return func(t *Transport) *Transport {
		t.peerNamer = func(*http.Request) string { return name }
		return t
	}
}

// TransportOptionPeerNamer is similar to TransportOptionPeerName but allows
// for mapping an outgoing request object to a particular peer name. This can
// be used for cases when an HTTP transport is used to call multiple
// dependencies instead of a single one.
func TransportOptionPeerNamer(namer func(*http.Request) string) TransportOption {
	return func(t *Transport) *Transport {
		t.peerNamer = namer
		return t
	}
}

// NewTransport creats an http.RoundTripper wrapper that injects zipkin
// headers into all outgoing requests.
func NewTransport(options ...TransportOption) func(c http.RoundTripper) http.RoundTripper {
	return func(c http.RoundTripper) http.RoundTripper {
		var wrapper = &Transport{
			spanName:  "OutgoingHTTPRequest",
			peerNamer: func(*http.Request) string { return "dependency" },
			wrapped:   c,
		}
		for _, option := range options {
			wrapper = option(wrapper)
		}
		return wrapper
	}
}
