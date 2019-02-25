package httptrace

import (
	"github.com/asecurityteam/logevent"
	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

// NewTracer generates an opentracing.Tracer implementation that uses the given
// Logger and metadata when generating and emitting spans.
func NewTracer(logger logevent.Logger, serviceName string, hostPort string) (opentracing.Tracer, error) {
	var collector = &collector{logger}
	var recorder = zipkin.NewRecorder(collector, false, hostPort, serviceName)
	return zipkin.NewTracer(recorder)
}
