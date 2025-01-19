# Use official golang image as the base image
FROM golang:1.23.5-alpine3.21 AS builder

# Install git
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Clone your repository (replace with your actual repo URL)
RUN git clone https://github.com/tryy3/weight-api-golang.git .

# Download dependencies
RUN go mod download

# Create the tmp directory
RUN mkdir -p tmp

# Build the application
RUN CGO_ENABLED=0 go build \
    -ldflags="-w -s" \
    -trimpath \
    -o tmp/serial-api bin/main.go

FROM alpine:3.19

# Create non-root user for security
RUN adduser -D app
# Add user to dialout group for serial port access
RUN addgroup -S dialout && adduser app dialout

# Document which ports are intended to be published
EXPOSE 8080

# Copy the binary from the builder stage
COPY --from=builder /app/tmp/serial-api /app/serial-api

# Set ownership of the application
RUN chown app:app /app/serial-api

# Switch to non-root user
USER app

# Run the application
CMD ["/app/serial-api"]