# builder stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
# Copy only necessary files for dependency resolution
COPY go.mod go.sum ./
RUN go mod download
# Copy the rest of the source code
COPY . .
# Build the application
RUN go build -o /goshort ./cmd/app/main.go

# final stage
FROM alpine:latest
WORKDIR /app
# Copy only the built binary from the builder stage
COPY --from=builder /goshort .
EXPOSE 8080
CMD ["./goshort"]