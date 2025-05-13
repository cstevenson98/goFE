FROM golang:1.21-alpine

WORKDIR /app

# Copy go mod file
COPY go.mod ./

# Download dependencies and generate go.sum
RUN go mod download && go mod tidy

# Copy the source code
COPY . .

# Build the application
RUN go build -o main .

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["./main"] 