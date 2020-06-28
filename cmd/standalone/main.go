package main

import (
	"flag"
	"io/ioutil"
	"os"

	"github.com/johanbrandhorst/grpc-gateway-boilerplate/gateway"
	"google.golang.org/grpc/grpclog"
)

var serverAddress = flag.String(
	"server-address",
	"dns:///0.0.0.0:10000",
	"The address to the gRPC server, in the gRPC standard naming format. "+
		"See https://github.com/grpc/grpc/blob/master/doc/naming.md for more information.",
)

func main() {
	flag.Parse()

	// Adds gRPC internal logs. This is quite verbose, so adjust as desired!
	log := grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(log)

	err := gateway.Run(*serverAddress)
	log.Fatalln(err)
}
