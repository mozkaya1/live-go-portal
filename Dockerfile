# Dockerfile for main portal
FROM golang:alpine

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Expose port
EXPOSE 8080

# Run the application with full path
CMD ["/app/main"]

