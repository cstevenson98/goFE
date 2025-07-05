FROM golang:1.21-alpine

# Install git for go mod download
RUN apk add --no-cache git

WORKDIR /app

# Copy the entire project to maintain the replace directive structure
COPY . .

# Change to the anthropic-agent directory
WORKDIR /app/backends/anthropic-agent

# Download dependencies and generate go.sum
RUN go mod download && go mod tidy

# Build the application
RUN go build -o main .

# Expose port 8081
EXPOSE 8081

# Run the application
CMD ["./main"] 