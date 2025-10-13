# Dockerfile Improvements for Kubernetes

This document outlines recommended improvements to the existing Dockerfiles to optimize them for Kubernetes deployment.

## Current Dockerfile Analysis

### 1. Backend Dockerfile (`/Dockerfile`)
**Current Status**: ✅ Good - Already follows many best practices

**Strengths**:
- Multi-stage build for smaller final image
- Non-root user execution
- Health check included
- Minimal Alpine base image

**Recommended Improvements**:

```dockerfile
# Multi-stage build for Go backend
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o main .

# Production stage with minimal base image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -s /bin/sh -u 1000 -G appuser appuser

# Create app directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy migration files and scripts
COPY --from=builder /app/scripts ./scripts

# Change ownership and set execute permissions
RUN chown -R appuser:appuser . && \
    chmod +x ./main

# Switch to non-root user
USER appuser

# Expose port 8081
EXPOSE 8081

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8081/api/health || exit 1

# Run the application
CMD ["./main"]
```

**Key Improvements**:
- Added build optimizations (`-ldflags="-w -s"`)
- Added timezone data for proper time handling
- Improved RUN command grouping
- More specific health check endpoint

### 2. Frontend Dockerfile (`/frontend/Dockerfile`)
**Current Status**: ⚠️ Needs Improvement

**Current Issues**:
- No health check
- No non-root user
- Missing production optimizations
- No multi-stage build for size optimization

**Improved Version**:

```dockerfile
# Multi-stage build for SvelteKit frontend
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies with npm ci for production builds
RUN npm ci --only=production && npm cache clean --force

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM node:20-alpine

# Install dumb-init for proper signal handling
RUN apk add --no-cache dumb-init

# Create non-root user
RUN addgroup -g 1001 -S nodeuser && \
    adduser -u 1001 -S nodeuser -G nodeuser

# Set working directory
WORKDIR /app

# Copy built application from builder stage
COPY --from=builder --chown=nodeuser:nodeuser /app/build ./build
COPY --from=builder --chown=nodeuser:nodeuser /app/package*.json ./
COPY --from=builder --chown=nodeuser:nodeuser /app/node_modules ./node_modules

# Switch to non-root user
USER nodeuser

# Expose port 3000
EXPOSE 3000

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/ || exit 1

# Use dumb-init for proper signal handling
ENTRYPOINT ["dumb-init", "--"]

# Start SvelteKit server
CMD ["node", "build"]
```

**Key Improvements**:
- Multi-stage build for smaller final image
- Non-root user execution
- Added dumb-init for proper signal handling
- Health check endpoint
- Production-only dependencies
- Proper file ownership

### 3. LilyPond Service Dockerfile (`/lilypond-service/Dockerfile`)
**Current Status**: ✅ Good - Well structured

**Recommended Improvements**:

```dockerfile
# Multi-stage build for LilyPond service
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o lilypond-service .

# Production stage
FROM alpine:latest

# Install LilyPond and dependencies in a single layer
RUN apk add --no-cache \
    lilypond \
    ghostscript \
    font-dejavu \
    fontconfig \
    ttf-dejavu \
    wget \
    && rm -rf /var/cache/apk/*

# Create non-root user with specific UID/GID for consistency
RUN addgroup -g 1001 -S lilypond && \
    adduser -u 1001 -S lilypond -G lilypond

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder --chown=lilypond:lilypond /app/lilypond-service .

# Create temp directory for LilyPond compilation with proper permissions
RUN mkdir -p /tmp/lilypond && \
    chown -R lilypond:lilypond /tmp/lilypond && \
    chmod 755 /tmp/lilypond

# Change to non-root user
USER lilypond

# Expose port
EXPOSE 8082

# Health check with proper timeout
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8082/health || exit 1

# Command to run the service
CMD ["./lilypond-service"]
```

**Key Improvements**:
- Added build optimizations
- Added wget for health checks
- Improved permission handling
- Consistent UID/GID

### 4. Nginx Dockerfile (`/nginx/Dockerfile`)
**Current Status**: ⚠️ Basic - Needs security improvements

**Improved Version**:

```dockerfile
FROM nginx:1.25-alpine

# Install wget for health checks
RUN apk add --no-cache wget

# Create nginx user with specific UID
RUN addgroup -g 101 -S nginx && \
    adduser -u 101 -S nginx -G nginx

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Create necessary directories with proper permissions
RUN mkdir -p /var/cache/nginx /var/run /var/log/nginx && \
    chown -R nginx:nginx /var/cache/nginx /var/run /var/log/nginx && \
    chmod -R 755 /var/cache/nginx /var/run /var/log/nginx

# Create health check endpoint configuration
RUN echo 'location /health { return 200 "healthy\\n"; add_header Content-Type text/plain; }' > /etc/nginx/conf.d/health.conf

# Switch to non-root user
USER nginx

# Expose port 80
EXPOSE 80

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
```

**Key Improvements**:
- Non-root user execution
- Health check endpoint
- Proper directory permissions
- Added wget for health checks

## Security Enhancements

### 1. Multi-Stage Build Best Practices
- Use specific image tags instead of `latest`
- Minimize layers in final image
- Copy only necessary files to production stage
- Use `.dockerignore` files to exclude unnecessary files

### 2. Security Context Requirements
All containers should run with:
```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000  # or appropriate UID
  runAsGroup: 1000 # or appropriate GID
  readOnlyRootFilesystem: true  # when possible
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
```

### 3. Resource Optimization
- Use Alpine Linux base images
- Remove package managers after installation
- Use multi-stage builds to exclude build dependencies
- Implement proper health checks

## Recommended .dockerignore Files

### Root .dockerignore
```
.git
.gitignore
.dockerignore
Dockerfile*
docker-compose*.yml
README*.md
docs/
terraform/
.cursor/
node_modules
*.log
*.tmp
.env*
coverage/
```

### Frontend .dockerignore
```
node_modules
npm-debug.log
.git
.gitignore
README.md
.env*
coverage/
.nyc_output
.DS_Store
```

## Build Optimization Script

Create a build script for optimized Docker builds:

```bash
#!/bin/bash
# scripts/build-optimized.sh

set -e

echo "Building optimized Docker images..."

# Backend
docker build \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  --cache-from muse-ai-backend:latest \
  -t muse-ai-backend:latest \
  -f Dockerfile .

# Frontend
docker build \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  --cache-from muse-ai-frontend:latest \
  -t muse-ai-frontend:latest \
  -f frontend/Dockerfile ./frontend

# LilyPond Service
docker build \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  --cache-from muse-ai-lilypond:latest \
  -t muse-ai-lilypond:latest \
  -f lilypond-service/Dockerfile ./lilypond-service

# Nginx
docker build \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  --cache-from muse-ai-nginx:latest \
  -t muse-ai-nginx:latest \
  -f nginx/Dockerfile ./nginx

echo "All images built successfully!"
```

## Image Scanning Integration

Add vulnerability scanning to your build process:

```yaml
# .github/workflows/security-scan.yml
name: Security Scan
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  scan:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Build image
      run: docker build -t test-image .
    
    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        image-ref: 'test-image'
        format: 'sarif'
        output: 'trivy-results.sarif'
    
    - name: Upload Trivy scan results to GitHub Security tab
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: 'trivy-results.sarif'
```

These improvements will enhance security, performance, and maintainability of your Docker images in a Kubernetes environment.
