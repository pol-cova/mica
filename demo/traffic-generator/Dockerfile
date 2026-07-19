FROM golang:1.25-alpine AS build
WORKDIR /src
COPY main.go .
RUN CGO_ENABLED=0 go build -o /traffic-generator main.go

FROM alpine:3.21
COPY --from=build /traffic-generator /traffic-generator
ENTRYPOINT ["/traffic-generator"]
