package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"mime"
	"net"
	"net/http"
	"os"
	"strings"

	xray "contrib.go.opencensus.io/exporter/aws"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rakyll/statik/fs"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	"github.com/johanbrandhorst/grpc-gateway-boilerplate/insecure"
	pbExample "github.com/johanbrandhorst/grpc-gateway-boilerplate/proto"
	"github.com/johanbrandhorst/grpc-gateway-boilerplate/server"

	// Static files
	_ "github.com/johanbrandhorst/grpc-gateway-boilerplate/statik"
)

// getOpenAPIHandler serves an OpenAPI UI.
// Adapted from https://github.com/philips/grpc-gateway-example/blob/a269bcb5931ca92be0ceae6130ac27ae89582ecc/cmd/serve.go#L63
func getOpenAPIHandler() http.Handler {
	mime.AddExtensionType(".svg", "image/svg+xml")

	statikFS, err := fs.New()
	if err != nil {
		panic("creating OpenAPI filesystem: " + err.Error())
	}

	return http.FileServer(statikFS)
}

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
	pbExample.RegisterUserServiceServer(s, server.New())

	// Serve gRPC Server
	log.Info("Serving gRPC on https://", addr)
	go func() {
		log.Fatal(s.Serve(lis))
	}()

	// See https://github.com/grpc/grpc/blob/master/doc/naming.md
	// for gRPC naming standard information.
	dialAddr := fmt.Sprintf("dns:///%s", addr)
	// Create a client connection to the gRPC Server we just started.
	// This is where the gRPC-Gateway proxies the requests.
	conn, err := grpc.DialContext(
		context.Background(),
		dialAddr,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(insecure.CertPool, "")),
		// Dial blocks until the underlying connection is established
		grpc.WithBlock(),
		// Add ocgrpc.ClientHandler for tracing the grpc client calls
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux(
		runtime.WithErrorHandler(server.CustomErrorHandler),
	)
	err = pbExample.RegisterUserServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	oa := getOpenAPIHandler()

	port := os.Getenv("PORT")
	if port == "" {
		port = "11000"
	}

	// Wrap the gateway mux with the OpenCensus HTTP handler
	openCensusHandler := &ochttp.Handler{
		Handler: gwmux,
	}

	gatewayAddr := "0.0.0.0:" + port
	gwServer := &http.Server{
		Addr: gatewayAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/api") {
				openCensusHandler.ServeHTTP(w, r)
				return
			}
			oa.ServeHTTP(w, r)
		}),
	}
	// Empty parameters mean use the TLS Config specified with the server.
	if strings.ToLower(os.Getenv("SERVE_HTTP")) == "true" {
		log.Info("Serving gRPC-Gateway and OpenAPI Documentation on http://", gatewayAddr)
		log.Fatalln(gwServer.ListenAndServe())
	}

	gwServer.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{insecure.Cert},
	}
	log.Info("Serving gRPC-Gateway and OpenAPI Documentation on https://", gatewayAddr)
	log.Fatalln(gwServer.ListenAndServeTLS("", ""))
}
