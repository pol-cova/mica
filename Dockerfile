FROM node:22-alpine AS web-build
WORKDIR /web
COPY web/package*.json ./
RUN npm ci
COPY web ./
RUN npm run build

FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
COPY --from=web-build /web/dist ./cmd/mica/web/dist
RUN CGO_ENABLED=0 go build -o /mica ./cmd/mica

FROM alpine:3.21
COPY --from=build /mica /mica
EXPOSE 8787
ENTRYPOINT ["/mica"]
CMD ["serve", "--addr", "0.0.0.0:8787", "--prometheus-url", "http://prometheus:9090"]
