# Muse AI Kubernetes Deployment Guide

This directory contains comprehensive documentation for deploying the Muse AI application to Kubernetes, from local development to production on Google Cloud Platform.

## 📁 Documentation Structure

### Core Documentation
- [`kubernetes-intro.md`](./kubernetes-intro.md) - **START HERE**: Complete introduction to Kubernetes with practical examples
- [`infrastructure-analysis.md`](./infrastructure-analysis.md) - Analysis of current Docker Compose setup and Kubernetes design decisions
- [`kubernetes-manifests.md`](./kubernetes-manifests.md) - Complete Kubernetes manifests for all services
- [`dockerfile-improvements.md`](./dockerfile-improvements.md) - Recommended improvements to existing Dockerfiles

### Deployment Guides
- [`local-deployment.md`](./local-deployment.md) - Step-by-step guide for local Kubernetes deployment with minikube
- [`gcp-manual-deployment.md`](./gcp-manual-deployment.md) - Manual deployment to Google Kubernetes Engine (GKE)
- [`gcp-automated-deployment.md`](./gcp-automated-deployment.md) - Automated deployment via GitHub Actions

### Scripts and Automation
- [`scripts/`](./scripts/) - Deployment scripts and automation tools
- [`github-actions/`](./github-actions/) - CI/CD pipeline configurations

## 🚀 Quick Start

1. **New to Kubernetes?**: Start with [`kubernetes-intro.md`](./kubernetes-intro.md) to learn the fundamentals
2. **Local Development**: Follow [`local-deployment.md`](./local-deployment.md) to deploy with minikube
3. **Production Deployment**: Use [`gcp-manual-deployment.md`](./gcp-manual-deployment.md) for GKE
4. **CI/CD Setup**: Implement [`gcp-automated-deployment.md`](./gcp-automated-deployment.md) for automation

## 🏗️ Architecture Overview

The Muse AI application consists of:
- **Backend**: Go 1.24+ API server with AI agent capabilities
- **Frontend**: React 18+ with TypeScript (SvelteKit)
- **LilyPond Service**: Music notation compilation service
- **PostgreSQL**: Database for metadata and document management
- **MinIO**: Object storage for files (.ly, .pdf, .svg, .midi)
- **Nginx**: Reverse proxy and load balancer

## 📋 Prerequisites

- **Local Development**: minikube, kubectl, Docker
- **GCP Deployment**: Google Cloud SDK, GKE cluster
- **CI/CD**: GitHub repository with Actions enabled

## 🔧 Best Practices Implemented

- **Security**: Non-root containers, resource limits, network policies
- **Observability**: Health checks, readiness probes, structured logging
- **Scalability**: Horizontal Pod Autoscaling, resource requests/limits
- **Reliability**: Multi-replica deployments, rolling updates, graceful shutdowns
- **Configuration**: ConfigMaps and Secrets for environment-specific settings

## 🎯 Next Steps

Choose your deployment path:
- For local development and testing → [`local-deployment.md`](./local-deployment.md)
- For production deployment → [`gcp-manual-deployment.md`](./gcp-manual-deployment.md)
- For automated CI/CD → [`gcp-automated-deployment.md`](./gcp-automated-deployment.md)
