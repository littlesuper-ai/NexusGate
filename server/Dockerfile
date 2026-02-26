FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o nexusgate ./cmd/nexusgate

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S nexusgate && adduser -S nexusgate -G nexusgate

COPY --from=builder /app/nexusgate /usr/local/bin/nexusgate

USER nexusgate
EXPOSE 8080
ENTRYPOINT ["nexusgate"]
