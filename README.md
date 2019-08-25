# grpc-gateway-boilerplate

[![Run on Google Cloud](https://storage.googleapis.com/cloudrun/button.svg)](https://console.cloud.google.com/cloudshell/editor?shellonly=true&cloudshell_image=gcr.io/cloudrun/button&cloudshell_git_repo=https://github.com/johanbrandhorst/grpc-gateway-boilerplate.git)

All the boilerplate you need to get started with writing grpc-gateway powered
REST services in Go.

## Running

Running `main.go` starts a web server on https://0.0.0.0:11000/. You can configure
the port used with the `$PORT` environment variable, and to serve on HTTP set
`$SERVE_HTTP=true`.

An OpenAPI UI is served on https://0.0.0.0:11000/.

## Requirements

Generating the files requires the `protoc` protobuf compiler.
Please install it according to the
[installation instructions](https://github.com/google/protobuf#protocol-compiler-installation)
for your specific platform.

## Getting started

After cloning the repo, there are a couple of initial steps;

1. Install the generate dependencies with `make install`.
   This will install `protoc-gen-go`, `protoc-gen-grpc-gateway`, `protoc-gen-swagger` and `statik` which
   are necessary for us to generate the Go, swagger and static files.
1. If you forked this repo, or cloned it into a different directory from the github structure,
   you will need to correct the import paths. Here's a nice `find` one-liner for accomplishing this
   (replace `yourscmprovider.com/youruser/yourrepo` with your cloned repo path):
   ```bash
   $ find . -path ./vendor -prune -o -type f \( -name '*.go' -o -name '*.proto' \) -exec sed -i -e "s;github.com/johanbrandhorst/grpc-gateway-boilerplate;yourscmprovider.com/youruser/yourrepo;g" {} +
   ```
1. Finally, generate the files with `make generate`.
   If you encounter an error here, make sure you've installed
   `protoc` and it is accessible in your `$PATH`, and make sure
   you've performed step 1.

Now you can run the web server with `go run main.go`.

## Making it your own

The next step is to define the interface you want to expose in
`proto/example.proto`. See https://developers.google.com/protocol-buffers/
tutorials and guides on writing protofiles.

Once that is done, regenerate the files using
`make generate`. This will mean you'll need to implement any functions in
`server/server.go`, or else the build will fail since your struct won't
be implementing the interface defined by the generated file in `proto/example.pb.go`.

This should hopefully be all you need to get started playing around with the gRPC-Gateway!
