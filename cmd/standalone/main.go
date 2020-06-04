package main

import (
	"flag"
	"io/ioutil"
	"os"

	xray "contrib.go.opencensus.io/exporter/aws"
	"github.com/johanbrandhorst/grpc-gateway-boilerplate/gateway"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/grpclog"
)

var (
	serverAddress = flag.String(
		"server-address",
		"dns:///0.0.0.0:10000",
		"The address to the gRPC server, in the gRPC standard naming format. "+
			"See https://github.com/grpc/grpc/blob/master/doc/naming.md for more information.",
	)
	awsRegion = flag.String(
		"aws-region",
		"us-east-2",
		"The AWS region to use with the X-Ray tracing exporter",
	)
)

func main() {
	flag.Parse()

	// Adds gRPC internal logs. This is quite verbose, so adjust as desired!
	log := grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(log)

	// Register the AWS X-Ray exporter to be able to retrieve
	// the collected spans.
	xrayExporter, err := xray.NewExporter(
		xray.WithVersion("latest"),
		xray.WithRegion(*awsRegion),
		xray.WithOnExport(func(in xray.OnExport) {
			log.Info("Publishing trace with ID: ", in.TraceID)
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create the AWS X-Ray exporter: %v", err)
	}
	// Do not forget to call Flush() before the application terminates.
	defer xrayExporter.Flush()

	// Register the trace exporter.
	trace.RegisterExporter(xrayExporter)

	// Always trace for this demonstration.
	// In production this can be set to a trace.ProbabilitySampler.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	err = gateway.Run(*serverAddress)
	log.Fatalln(err)
}
