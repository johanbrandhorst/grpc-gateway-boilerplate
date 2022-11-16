BUF_VERSION:=v1.9.0

generate:
	go run github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION) generate

lint:
	go run github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION) lint
	go run github.com/bufbuild/buf/cmd/buf@$(BUF_VERSION) breaking --against 'https://github.com/johanbrandhorst/grpc-gateway-boilerplate.git#branch=master'
