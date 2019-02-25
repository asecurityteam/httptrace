package httptrace

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"

	"github.com/asecurityteam/logevent"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
)

// collector implements the openzipkin Collector interface.
type collector struct {
	logevent.Logger
}

func (c *collector) Collect(s *zipkincore.Span) error {
	var zipkinLog = structFromSpan(s)
	c.Info(zipkinLog)
	return nil
}

func (c *collector) Close() error {
	return nil
}

func structFromSpan(s *zipkincore.Span) frame {
	var result = frame{}

	result.Zipkin.TraceID = fmt.Sprintf("%016x", s.GetTraceID())
	result.Zipkin.SpanID = fmt.Sprintf("%016x", s.GetID())
	if s.IsSetParentID() {
		result.Zipkin.ParentID = fmt.Sprintf("%016x", s.GetParentID())
	}
	if s.IsSetDuration() {
		result.Zipkin.Duration = s.GetDuration()
	}
	if s.IsSetTimestamp() {
		result.Zipkin.Timestamp = s.GetTimestamp()
	}
	result.Zipkin.Name = s.GetName()
	var annotations = s.GetAnnotations()
	var binaryAnnotations = s.GetBinaryAnnotations()
	result.Zipkin.Annotations = make([]annotation, len(annotations))
	result.Zipkin.BinaryAnnotations = make([]binaryAnnotation, len(binaryAnnotations))

	for offset, an := range annotations {
		var host = an.GetHost()
		result.Zipkin.Annotations[offset] = annotation{
			Timestamp: an.GetTimestamp(),
			Value:     an.GetValue(),
		}
		if host != nil {
			var ip = make(net.IP, 4)
			binary.BigEndian.PutUint32(ip, uint32(host.GetIpv4()))
			result.Zipkin.Annotations[offset].Endpoint = endpoint{
				Ipv4:        netIP(ip),
				Port:        int(host.GetPort()),
				ServiceName: host.GetServiceName(),
			}
		}
	}
	for offset, binan := range binaryAnnotations {
		var host = binan.GetHost()
		result.Zipkin.BinaryAnnotations[offset] = binaryAnnotation{
			Key:   binan.GetKey(),
			Value: string(binan.GetValue()),
		}
		if host != nil {
			var ip = make(net.IP, 4)
			binary.BigEndian.PutUint32(ip, uint32(host.GetIpv4()))
			result.Zipkin.BinaryAnnotations[offset].Endpoint = endpoint{
				Ipv4:        netIP(ip),
				Port:        int(host.GetPort()),
				ServiceName: host.GetServiceName(),
			}
		}
	}
	return result
}

// This is a set of types we can do a Marshal on for JSON.
type netIP net.IP

// MarshalJSON implements the Marshal interface for the netIP type.
func (ip netIP) MarshalJSON() ([]byte, error) {
	return json.Marshal(net.IP(ip).String())
}

type endpoint struct {
	ServiceName string `logevent:"serviceName"`
	Port        int    `logevent:"port"`
	Ipv4        netIP  `logevent:"ipv4"`
}

type annotation struct {
	Timestamp int64    `logevent:"timestamp"`
	Value     string   `logevent:"value"`
	Endpoint  endpoint `logevent:"endpoint"`
}

type binaryAnnotation struct {
	Key      string   `logevent:"key"`
	Value    string   `logevent:"value"`
	Endpoint endpoint `logevent:"endpoint"`
}

type jsonSpan struct {
	TraceID           string             `logevent:"traceId"`
	SpanID            string             `logevent:"id"`
	ParentID          string             `logevent:"parentId"`
	Name              string             `logevent:"name"`
	Timestamp         int64              `logevent:"timestamp"`
	Duration          int64              `logevent:"duration"`
	Annotations       []annotation       `logevent:"annotations"`
	BinaryAnnotations []binaryAnnotation `logevent:"binaryAnnotations"`
}

type frame struct {
	Zipkin  jsonSpan `logevent:"zipkin"`
	Message string   `logevent:"message,default=span-complete"`
}
