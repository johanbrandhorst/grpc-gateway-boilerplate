# Build stage
FROM golang AS build-env
ADD . /src/grpc-gateway-boilerplate
ENV CGO_ENABLED=0
RUN cd /src/grpc-gateway-boilerplate && go build -o /app

# Production stage
FROM scratch
COPY --from=build-env /app /

ENTRYPOINT ["/app"]
