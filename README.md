<a id="markdown-httptrace--zipkin-tracing-integration-for-http-services" name="httptrace--zipkin-tracing-integration-for-http-services"></a>
# httptrace -Zipkin tracing integration for HTTP services. #
[![GoDoc](https://godoc.org/github.com/asecurityteam/httptrace?status.svg)](https://godoc.org/github.com/asecurityteam/httptrace)
[![Build Status](https://travis-ci.org/asecurityteam/httptrace.png?branch=master)](https://travis-ci.org/asecurityteam/httptrace)
[![codecov.io](https://codecov.io/github/asecurityteam/httptrace/coverage.svg?branch=master)](https://codecov.io/github/asecurityteam/httptrace?branch=master)

*Status: Production*

<!-- TOC -->

- [httptrace -Zipkin tracing integration for HTTP services.](#httptrace--zipkin-tracing-integration-for-http-services)
    - [Usage](#usage)
        - [HTTP Service](#http-service)
        - [HTTP Client](#http-client)
    - [Span Logs](#span-logs)
    - [Contributing](#contributing)
        - [License](#license)
        - [Contributing Agreement](#contributing-agreement)

<!-- /TOC -->

This project contains middleware for HTTP services and clients that uses
[openzipkin-go](https://github.com/openzipkin/zipkin-go-opentracing) and
[logevent](https://github.com/asecurityteam/logevent) to both propagate traces
between HTTP services and emit traces to the service logs.

<a id="markdown-usage" name="usage"></a>
## Usage ##

<a id="markdown-http-service" name="http-service"></a>
### HTTP Service ###

The middleware exported is a `func(http.Handler) http.Handler` and should
work with virtually any router/mux implementation that supports middleware.

```go
var middleware = httptrace.NewMiddleware(
  httptrace.MiddlewareOptionServiceName("my-service"),
)
```

The middleware uses
[opentracing-go](https://github.com/opentracing/opentracing-go) to manage
spans and should interoperate with other uses of opentracing that leverage the
`context` as a source of the tracer. If no trace is found in the incoming
request then the middleware will generate a new trace and root span using the
given service name. If the incoming request does contain a trace, via zipkin
headers, then the middleware generate a span within that trace that is a child
of the incoming span.

If you need the identifier of the active trace at any point within a request,
you can use the `TraceIDFromContext` or `SpanIDFromContext` helpers which will
return the ID in a hex encoded string which is what typically ships over via
headers to other services.

<a id="markdown-http-client" name="http-client"></a>
### HTTP Client ###

In addition to an HTTP middleware, there is also an `http.RoundTripper` wrapper
included that will properly manage spans for outgoing HTTP requests. To apply:

```golang
var client = &http.Client{
  Transport: httptrace.NewTransport(
    httptrace.TransportOptionSpanName("outgoing_http_request"),
    httptrace.TransportOptionPeerName("remote-service-name"),
  )(http.DefaultTransport),
}
```

<a id="markdown-span-logs" name="span-logs"></a>
## Span Logs ##

As each span is marked as complete the middlewares will use the
`logevent.Logger` contained within the request context to emit a line like:

```json
{"message": "span-complete", "zipkin": {"traceId": "", "id": "", "parentId": "", "name": "", "timestamp": "", "duration": "", "annotations": [{"timestamp": "", "value": ""}], "binaryAnnotations": [{"key": "", "value": ""}]}}
```

<a id="markdown-contributing" name="contributing"></a>
## Contributing ##

<a id="markdown-license" name="license"></a>
### License ###

This project is licensed under Apache 2.0. See LICENSE.txt for details.

<a id="markdown-contributing-agreement" name="contributing-agreement"></a>
### Contributing Agreement ###

Atlassian requires signing a contributor's agreement before we can accept a
patch. If you are an individual you can fill out the
[individual CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the
[corporate CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
