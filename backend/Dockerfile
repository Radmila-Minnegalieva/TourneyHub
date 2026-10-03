FROM golang:1.25.8-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S tourneyhub && adduser -S -G tourneyhub tourneyhub
WORKDIR /app
COPY --from=builder /out/api /out/migrate /app/
COPY --chown=tourneyhub:tourneyhub config /app/config
COPY --chown=tourneyhub:tourneyhub migrations /app/migrations
USER tourneyhub
EXPOSE 8080
CMD ["/app/api"]
