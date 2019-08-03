generate:
	# Generate go, gRPC-Gateway, swagger output.
	#
	# -I declares import folders, in order of importance
	# This is how proto resolves the protofile imports.
	# It will check for the protofile relative to each of these
	# folders and use the first one it finds.
	#
	# --go_out generates go Protobuf output with gRPC plugin enabled.
	# 		paths=source_relative means the file should be generated
	# 		relative to the input proto file.
	# --grpc-gateway_out generates gRPC-Gateway output.
	# --swagger_out generates an OpenAPI 2.0 specification for our gRPC-Gateway endpoints.
	#
	# proto/example.proto is the location of the protofile we use.
	protoc \
		-I proto \
		-I third_party/grpc-gateway/ \
		-I third_party/googleapis \
		--go_out=plugins=grpc,paths=source_relative:./proto \
		--grpc-gateway_out=./proto \
		--swagger_out=third_party/OpenAPI/ \
		proto/example.proto

	mv ./proto/github.com/johanbrandhorst/grpc-gateway-boilerplate/proto/* ./proto/
	rm -r ./proto/github.com

	# Generate static assets for OpenAPI UI
	statik -m -f -src third_party/OpenAPI/

install:
	GO111MODULE=on go get \
		github.com/golang/protobuf/protoc-gen-go \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
		github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger \
		github.com/rakyll/statik
