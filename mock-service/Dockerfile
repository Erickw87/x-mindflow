FROM golang:1.24-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mock-server ./cmd/mock-server

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/mock-server .
COPY --from=builder /build/configs ./configs

EXPOSE 8080

ENTRYPOINT ["./mock-server"]
CMD ["--config", "/app/configs/user-service-v1.1.yaml"]
