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

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/rakyll/statik/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	"github.com/johanbrandhorst/grpc-gateway-boilerplate/insecure"
	pbExample "github.com/johanbrandhorst/grpc-gateway-boilerplate/proto"
	"github.com/johanbrandhorst/grpc-gateway-boilerplate/server"

	// Static files
	_ "github.com/johanbrandhorst/grpc-gateway-boilerplate/statik"
)

// serveOpenAPI serves an OpenAPI UI on /openapi-ui/
// Adapted from https://github.com/philips/grpc-gateway-example/blob/a269bcb5931ca92be0ceae6130ac27ae89582ecc/cmd/serve.go#L63
func serveOpenAPI(mux *http.ServeMux) error {
	mime.AddExtensionType(".svg", "image/svg+xml")

	statikFS, err := fs.New()
	if err != nil {
		return err
	}

	// Expose files in static on <host>/openapi-ui
	fileServer := http.FileServer(statikFS)
	prefix := "/openapi-ui/"
	mux.Handle(prefix, http.StripPrefix(prefix, fileServer))
	return nil
}

func main() {
	// Adds gRPC internal logs. This is quite verbose, so adjust as desired!
	log := grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(log)

	addr := "0.0.0.0:10000"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	s := grpc.NewServer(
		// TODO: Replace with your own certificate!
		grpc.Creds(credentials.NewServerTLSFromCert(&insecure.Cert)),
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
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	jsonpb := &runtime.JSONPb{
		EmitDefaults: true,
		Indent:       "  ",
		OrigName:     true,
	}
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, jsonpb),
	)
	err = pbExample.RegisterUserServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", gwmux)
	err = serveOpenAPI(mux)
	if err != nil {
		log.Fatalln("Failed to serve OpenAPI UI")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "11000"
	}
	gatewayAddr := "0.0.0.0:" + port
	gwServer := &http.Server{
		Addr:    gatewayAddr,
		Handler: mux,
	}
	// Empty parameters mean use the TLS Config specified with the server.
	if strings.ToLower(os.Getenv("SERVE_HTTP")) == "true" {
		log.Info("Serving gRPC-Gateway on http://", gatewayAddr)
		log.Info("Serving OpenAPI Documentation on http://", gatewayAddr, "/openapi-ui/")
		log.Fatalln(gwServer.ListenAndServe())
	}

	gwServer.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{insecure.Cert},
	}
	log.Info("Serving gRPC-Gateway on https://", gatewayAddr)
	log.Info("Serving OpenAPI Documentation on https://", gatewayAddr, "/openapi-ui/")
	log.Fatalln(gwServer.ListenAndServeTLS("", ""))
}
