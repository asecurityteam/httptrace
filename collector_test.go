package stridetrace

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
)

func TestCollectorNoParent(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var logger = NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any()).Do(func(event interface{}) {
		var evt frame
		var ok bool
		if evt, ok = event.(frame); !ok {
			t.Error("did not log a zipkin frame")
		}
		if evt.Zipkin.ParentID != "" {
			t.Errorf("expected no parent but found %s", evt.Zipkin.ParentID)
		}
		if evt.Zipkin.TraceID != "0000000000000001" {
			t.Errorf("expected trace 0000000000000001 but found %s", evt.Zipkin.TraceID)
		}
		if evt.Zipkin.SpanID != "0000000000000002" {
			t.Errorf("expected span 0000000000000002 but found %s", evt.Zipkin.SpanID)
		}
		if evt.Zipkin.Name != "TEST" {
			t.Errorf("expected name TEST but found %s", evt.Zipkin.Name)
		}
		if len(evt.Zipkin.BinaryAnnotations) != 1 {
			t.Errorf("expected 1 binary annotation but got %d", len(evt.Zipkin.BinaryAnnotations))
		} else {
			if evt.Zipkin.BinaryAnnotations[0].Key != "TESTTAG" || evt.Zipkin.BinaryAnnotations[0].Value != "TESTVALUE" {
				t.Errorf("expected binary annotation of TESTTAG=TESTVALUE but got %s=%s", evt.Zipkin.BinaryAnnotations[0].Key, evt.Zipkin.BinaryAnnotations[0].Value)
			}
		}
		if len(evt.Zipkin.Annotations) != 1 {
			t.Errorf("expected 1 annotation but got %d", len(evt.Zipkin.Annotations))
		} else {
			if evt.Zipkin.Annotations[0].Value != "TESTANNOTATION" || evt.Zipkin.Annotations[0].Endpoint.ServiceName != "TESTSERVICE" {
				t.Errorf("expected annotation TESTANNOTATION with endpoint TESTSERVICE but got %s:%s", evt.Zipkin.Annotations[0].Value, evt.Zipkin.Annotations[0].Endpoint.ServiceName)
			}
		}
	})
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
}

func TestCollectorWithParent(t *testing.T) {
	var ctrl = gomock.NewController(t)
	defer ctrl.Finish()

	var logger = NewMockLogger(ctrl)
	logger.EXPECT().Info(gomock.Any()).Do(func(event interface{}) {
		var evt frame
		var ok bool
		if evt, ok = event.(frame); !ok {
			t.Error("did not log a zipkin frame")
		}
		if evt.Zipkin.ParentID != "0000000000000003" {
			t.Errorf("expected parent 0000000000000003 but found %s", evt.Zipkin.ParentID)
		}
		if evt.Zipkin.TraceID != "0000000000000001" {
			t.Errorf("expected trace 0000000000000001 but found %s", evt.Zipkin.TraceID)
		}
		if evt.Zipkin.SpanID != "0000000000000002" {
			t.Errorf("expected span 0000000000000002 but found %s", evt.Zipkin.SpanID)
		}
		if evt.Zipkin.Name != "TEST" {
			t.Errorf("expected name TEST but found %s", evt.Zipkin.Name)
		}
		if len(evt.Zipkin.BinaryAnnotations) != 1 {
			t.Errorf("expected 1 binary annotation but got %d", len(evt.Zipkin.BinaryAnnotations))
		} else {
			if evt.Zipkin.BinaryAnnotations[0].Key != "TESTTAG" || evt.Zipkin.BinaryAnnotations[0].Value != "TESTVALUE" {
				t.Errorf("expected binary annotation of TESTTAG=TESTVALUE but got %s=%s", evt.Zipkin.BinaryAnnotations[0].Key, evt.Zipkin.BinaryAnnotations[0].Value)
			}
		}
		if len(evt.Zipkin.Annotations) != 1 {
			t.Errorf("expected 1 annotation but got %d", len(evt.Zipkin.Annotations))
		} else {
			if evt.Zipkin.Annotations[0].Value != "TESTANNOTATION" || evt.Zipkin.Annotations[0].Endpoint.ServiceName != "TESTSERVICE" {
				t.Errorf("expected annotation TESTANNOTATION with endpoint TESTSERVICE but got %s:%s", evt.Zipkin.Annotations[0].Value, evt.Zipkin.Annotations[0].Endpoint.ServiceName)
			}
		}
	})
	var collector = collector{logger}
	var parentID = int64(3)
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
}
