package trace

import (
	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	httpreporter "github.com/openzipkin/zipkin-go/reporter/http"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	logreporter "github.com/openzipkin/zipkin-go/reporter/log"
	"log"
	"os"
	"time"
)

func GetZipkinTracer(serviceName string, hostPort string) opentracing.Tracer {
	// set up a span reporter
	reporter := zipkinhttp.NewReporter("http://127.0.0.1:9411/api/v2/spans")
	//defer reporter.Close()

	// create our local service endpoint
	endpoint, err := zipkin.NewEndpoint(serviceName, hostPort)
	if err != nil {
		log.Fatalf("unable to create local endpoint: %+v\n", err)
	}

	// initialize our tracer
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		log.Fatalf("unable to create tracer: %+v\n", err)
	}

	// use zipkin-go-opentracing to wrap our tracer
	tracer := zipkinot.Wrap(nativeTracer)

	// optionally set as Global OpenTracing tracer instance
	//opentracing.SetGlobalTracer(tracer)
	return tracer
}

func GetZipkinLogTracer(serviceName string, hostPort string) *zipkin.Tracer {
	// set up a span reporter
	reporter := logreporter.NewReporter(log.New(os.Stderr, "", log.LstdFlags))
	defer reporter.Close()

	// create our local service endpoint
	endpoint, err := zipkin.NewEndpoint(serviceName, hostPort)
	if err != nil {
		log.Fatalf("unable to create local endpoint: %+v\n", err)
	}

	// initialize our tracer
	tracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(endpoint))
	if err != nil {
		log.Fatalf("unable to create tracer: %+v\n", err)
	}
	return tracer
}

func GetZipkinBasicTracer(serviceName, hostPort string) *zipkin.Tracer {
	// create a reporter to be used by the tracer
	reporter := httpreporter.NewReporter("http://localhost:9411/api/v2/spans")
	//defer reporter.Close()

	// set-up the local endpoint for our service
	endpoint, err := zipkin.NewEndpoint(serviceName, hostPort)
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
	//span := tracer.StartSpan("user.UserService.GetUserInfo")
	//span.Finish()
	return tracer
}
