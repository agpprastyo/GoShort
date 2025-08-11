FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o goshort ./cmd/app

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/goshort .
COPY .env .
EXPOSE 8080
CMD ["./goshort"]