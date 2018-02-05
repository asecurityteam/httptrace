package stridetrace

import (
	"context"
	"testing"

	"bitbucket.org/atlassian/logevent"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
)

func TestCollectorNoParent(t *testing.T) {
	var emitted bool
	var logFunc = func(ctx context.Context, level logevent.LogLevel, message string, annotations map[string]interface{}) {
		emitted = true
		var traceData, ok = annotations["zipkin"].(jsonSpan)
		if !ok {
			t.Fatal("could not fetch the zipkin trace object", annotations)
		}
		if traceData.Name != "TEST" {
			t.Errorf("expected the name TEST but found %s", traceData.Name)
		}
		if traceData.ParentID != "" {
			t.Errorf("expected no ParentID but found %s", traceData.ParentID)
		}
		if traceData.TraceID != "0000000000000001" {
			t.Errorf("expected the TraceID 0000000000000001 but found %s", traceData.TraceID)
		}
		if traceData.SpanID != "0000000000000002" {
			t.Errorf("expected the SpanID 0000000000000002 but found %s", traceData.SpanID)
		}
		if len(traceData.BinaryAnnotations) != 1 {
			t.Errorf("expected 1 binary annotation, got %d", len(traceData.BinaryAnnotations))
		} else {
			if traceData.BinaryAnnotations[0].Key != "TESTTAG" || traceData.BinaryAnnotations[0].Value != "TESTVALUE" {
				t.Errorf("expected binary annotation of TESTTAG=TESTVALUE, got %s=%s", traceData.BinaryAnnotations[0].Key, traceData.BinaryAnnotations[0].Value)
			}
		}
		for _, a := range traceData.Annotations {
			switch a.Value {
			case "TESTANNOTATION":
			default:
				t.Errorf("expected annotation ss or sr, got %s", a.Value)
			}
		}
	}
	var logger = logevent.New(context.Background(), logFunc)
	var collector = collector{logger}
	var span = &zipkincore.Span{
		ParentID: nil,
		TraceID:  1,
		ID:       2,
		Name:     "TEST",
		Annotations: []*zipkincore.Annotation{
			&zipkincore.Annotation{
				Value: "TESTANNOTATION",
				Host: &zipkincore.Endpoint{
					ServiceName: "TESTSERVICE",
					Port:        80,
				},
			},
		},
		BinaryAnnotations: []*zipkincore.BinaryAnnotation{
			&zipkincore.BinaryAnnotation{
				Key:   "TESTTAG",
				Value: []byte("TESTVALUE"),
			},
		},
	}
	collector.Collect(span)
	if !emitted {
		t.Fatal("did not log")
	}
}

func TestCollectorWithParent(t *testing.T) {
	var emitted bool
	var logFunc = func(ctx context.Context, level logevent.LogLevel, message string, annotations map[string]interface{}) {
		emitted = true
		var traceData, ok = annotations["zipkin"].(jsonSpan)
		if !ok {
			t.Fatal("could not fetch the zipkin trace object")
		}
		if traceData.Name != "TEST" {
			t.Errorf("expected the name TEST but found %s", traceData.Name)
		}
		if traceData.ParentID != "0000000000000000" {
			t.Errorf("expected the ParentID 0000000000000001 but found %s", traceData.ParentID)
		}
		if traceData.TraceID != "0000000000000001" {
			t.Errorf("expected the TraceID 0000000000000001 but found %s", traceData.TraceID)
		}
		if traceData.SpanID != "0000000000000002" {
			t.Errorf("expected the SpanID 0000000000000002 but found %s", traceData.SpanID)
		}
		if len(traceData.BinaryAnnotations) != 1 {
			t.Errorf("expected 1 binary annotation, got %d", len(traceData.BinaryAnnotations))
		} else {
			if traceData.BinaryAnnotations[0].Key != "TESTTAG" || traceData.BinaryAnnotations[0].Value != "TESTVALUE" {
				t.Errorf("expected binary annotation of TESTTAG=TESTVALUE, got %s=%s", traceData.BinaryAnnotations[0].Key, traceData.BinaryAnnotations[0].Value)
			}
		}
		for _, a := range traceData.Annotations {
			switch a.Value {
			case "TESTANNOTATION":
			default:
				t.Errorf("expected annotation ss or sr, got %s", a.Value)
			}
		}
	}
	var logger = logevent.New(context.Background(), logFunc)
	var collector = collector{logger}
	var parentID = int64(0)
	var span = &zipkincore.Span{
		ParentID: &parentID,
		TraceID:  1,
		ID:       2,
		Name:     "TEST",
		Annotations: []*zipkincore.Annotation{
			&zipkincore.Annotation{
				Value: "TESTANNOTATION",
				Host: &zipkincore.Endpoint{
					ServiceName: "TESTSERVICE",
					Port:        80,
				},
			},
		},
		BinaryAnnotations: []*zipkincore.BinaryAnnotation{
			&zipkincore.BinaryAnnotation{
				Key:   "TESTTAG",
				Value: []byte("TESTVALUE"),
			},
		},
	}
	collector.Collect(span)
	if !emitted {
		t.Fatal("did not log")
	}
}
