FROM golang:1.25-alpine AS build
WORKDIR /src
COPY main.go .
RUN CGO_ENABLED=0 go build -o /checkout-service main.go

FROM alpine:3.21
COPY --from=build /checkout-service /checkout-service
EXPOSE 8081
ENTRYPOINT ["/checkout-service"]
