# Multi-stage build: toolchain no builder, binário enxuto no final.
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/carga ./cmd/carga

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/api /api
COPY --from=builder /out/carga /carga
COPY --from=builder /app/migrations /migrations
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/api"]
