package main

import (
	"log"
	"time"

	"github.com/openzipkin/zipkin-go"
	httpreporter "github.com/openzipkin/zipkin-go/reporter/http"
)

func main() {
	// create a reporter to be used by the tracer
	reporter := httpreporter.NewReporter("http://localhost:9411/api/v2/spans")
	defer reporter.Close()

	// set-up the local endpoint for our service
	endpoint, err := zipkin.NewEndpoint("demoService", "172.20.23.100:80")
	if err != nil {
		log.Fatalf("unable to create local endpoint: %+v\n", err)
	}

	// set-up our sampling strategy
	sampler, err := zipkin.NewBoundarySampler(1, time.Now().UnixNano())
	if err != nil {
		log.Fatalf("unable to create sampler: %+v\n", err)
	}

	// initialize the tracer
	tracer, err := zipkin.NewTracer(
		reporter,
		zipkin.WithLocalEndpoint(endpoint),
		zipkin.WithSampler(sampler),
	)
	if err != nil {
		log.Fatalf("unable to create tracer: %+v\n", err)
	}

	// tracer can now be used to create spans.
	span := tracer.StartSpan("some_operation")
	// ... do some work ...
	span.Finish()

}
