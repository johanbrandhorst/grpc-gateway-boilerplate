package main

import (
	"io/ioutil"
	"net"
	"os"

	xray "contrib.go.opencensus.io/exporter/aws"
	"github.com/johanbrandhorst/grpc-gateway-boilerplate/gateway"
	"github.com/johanbrandhorst/grpc-gateway-boilerplate/insecure"
	usersv1 "github.com/johanbrandhorst/grpc-gateway-boilerplate/proto/users/v1"
	"github.com/johanbrandhorst/grpc-gateway-boilerplate/server"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
)

func main() {
	// Adds gRPC internal logs. This is quite verbose, so adjust as desired!
	log := grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(log)
	if os.Getenv("AWS_REGION") == "" {
		log.Fatalln("AWS_REGION must be set")
	}

	addr := "0.0.0.0:10000"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	// Register the AWS X-Ray exporter to be able to retrieve
	// the collected spans.
	xrayExporter, err := xray.NewExporter(
		xray.WithVersion("latest"),
		xray.WithRegion(os.Getenv(os.Getenv("AWS_REGION"))),
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

	s := grpc.NewServer(
		// TODO: Replace with your own certificate!
		grpc.Creds(credentials.NewServerTLSFromCert(&insecure.Cert)),
		// Add ocgrpc.ServerHandler{} for tracing the grpc server
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	)
	usersv1.RegisterUserServiceServer(s, server.New())

	// Serve gRPC Server
	log.Info("Serving gRPC on https://", addr)
	go func() {
		log.Fatal(s.Serve(lis))
	}()

	err = gateway.Run("dns:///" + addr)
	log.Fatalln(err)
}
