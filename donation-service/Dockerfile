FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o donation-service -ldflags "-s -w" .

FROM alpine:3.21

RUN apk upgrade --no-cache && \
    addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/donation-service .

USER appuser

EXPOSE 8082

CMD ["./donation-service"]
