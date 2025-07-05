FROM golang:1.21-alpine

WORKDIR /app

# Copy the entire project to maintain the replace directive structure
COPY . .

# Change to the api-example directory
WORKDIR /app/backends/api-example

# Download dependencies and generate go.sum
RUN go mod download && go mod tidy

# Build the application
RUN go build -o main .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./main"] 