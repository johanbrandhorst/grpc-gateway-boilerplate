# Build stage
FROM golang AS build-env
ADD . /go/src/github.com/johanbrandhorst/grpc-gateway-boilerplate
ENV CGO_ENABLED=0
RUN cd /go/src/github.com/johanbrandhorst/grpc-gateway-boilerplate && go build -o /app

# Production stage
FROM scratch
COPY --from=build-env /app /

ENTRYPOINT ["/app"]
