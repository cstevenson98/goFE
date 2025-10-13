# GCP Automated Deployment via GitHub Actions

This guide sets up automated deployment of the Muse AI application to Google Kubernetes Engine (GKE) using GitHub Actions CI/CD pipeline.

## Overview

The automated deployment pipeline will:
1. **Build and test** the application on code changes
2. **Build and push** Docker images to Google Container Registry
3. **Deploy** to staging environment on feature branches
4. **Deploy** to production on main branch
5. **Run** automated tests and health checks
6. **Notify** on deployment success/failure

## Prerequisites

### Required Setup
- GKE cluster running (see [`gcp-manual-deployment.md`](./gcp-manual-deployment.md))
- GitHub repository with the Muse AI codebase
- GCP project with required APIs enabled
- Service account with deployment permissions

## Step 1: Set Up GCP Service Account for CI/CD

### Create Deployment Service Account

```bash
# Set project variables
export PROJECT_ID="your-project-id"
export SERVICE_ACCOUNT_NAME="muse-ai-deployer"

# Create service account
gcloud iam service-accounts create $SERVICE_ACCOUNT_NAME \
  --display-name="Muse AI GitHub Actions Deployer"

# Grant necessary permissions
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:$SERVICE_ACCOUNT_NAME@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/container.developer"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:$SERVICE_ACCOUNT_NAME@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/container.clusterAdmin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:$SERVICE_ACCOUNT_NAME@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/storage.admin"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:$SERVICE_ACCOUNT_NAME@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"

# Create and download service account key
gcloud iam service-accounts keys create ~/gcp-key.json \
  --iam-account=$SERVICE_ACCOUNT_NAME@$PROJECT_ID.iam.gserviceaccount.com
```

## Step 2: Configure GitHub Repository Secrets

### Add Repository Secrets

Go to your GitHub repository → Settings → Secrets and Variables → Actions, and add:

| Secret Name | Value | Description |
|-------------|-------|-------------|
| `GCP_PROJECT_ID` | `your-project-id` | Your GCP project ID |
| `GCP_SA_KEY` | Contents of `~/gcp-key.json` | Base64 encoded service account key |
| `GKE_CLUSTER_NAME` | `muse-ai-cluster` | Your GKE cluster name |
| `GKE_CLUSTER_REGION` | `us-central1` | Your GKE cluster region |
| `ANTHROPIC_API_KEY` | `your-anthropic-api-key` | Anthropic API key for AI features |
| `POSTGRES_PASSWORD` | `your-secure-password` | Database password |

### Base64 Encode Service Account Key

```bash
# Encode the service account key
cat ~/gcp-key.json | base64 -w 0

# Copy the output and paste it as GCP_SA_KEY secret
```

## Step 3: Create GitHub Actions Workflows

### Main CI/CD Pipeline

Create `.github/workflows/ci-cd.yml`:

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  PROJECT_ID: ${{ secrets.GCP_PROJECT_ID }}
  GKE_CLUSTER: ${{ secrets.GKE_CLUSTER_NAME }}
  GKE_REGION: ${{ secrets.GKE_CLUSTER_REGION }}
  REGISTRY: gcr.io
  NAMESPACE: muse-ai

jobs:
  test:
    name: Test Application
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_PASSWORD: test_password
          POSTGRES_DB: muse_ai_test
          POSTGRES_USER: muse_ai
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24'

    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
        restore-keys: |
          ${{ runner.os }}-go-

    - name: Run Go tests
      env:
        DATABASE_URL: postgres://muse_ai:test_password@localhost:5432/muse_ai_test?sslmode=disable
      run: |
        go mod download
        go test -v ./...

    - name: Set up Node.js
      uses: actions/setup-node@v4
      with:
        node-version: '20'
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json

    - name: Install frontend dependencies
      working-directory: frontend
      run: npm ci

    - name: Run frontend tests
      working-directory: frontend
      run: npm run test

    - name: Build frontend
      working-directory: frontend
      run: npm run build

  build-and-deploy:
    name: Build and Deploy
    runs-on: ubuntu-latest
    needs: test
    if: github.ref == 'refs/heads/main' || github.ref == 'refs/heads/develop'

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Cloud SDK
      uses: google-github-actions/setup-gcloud@v1
      with:
        service_account_key: ${{ secrets.GCP_SA_KEY }}
        project_id: ${{ secrets.GCP_PROJECT_ID }}

    - name: Configure Docker for GCR
      run: gcloud auth configure-docker

    - name: Set environment variables
      run: |
        echo "IMAGE_TAG=$(echo $GITHUB_SHA | cut -c1-7)" >> $GITHUB_ENV
        echo "BRANCH_NAME=$(echo ${GITHUB_REF#refs/heads/})" >> $GITHUB_ENV

    - name: Build Docker images
      run: |
        # Build backend
        docker build -t $REGISTRY/$PROJECT_ID/muse-ai-backend:$IMAGE_TAG \
                     -t $REGISTRY/$PROJECT_ID/muse-ai-backend:latest \
                     -f Dockerfile .
        
        # Build frontend
        docker build -t $REGISTRY/$PROJECT_ID/muse-ai-frontend:$IMAGE_TAG \
                     -t $REGISTRY/$PROJECT_ID/muse-ai-frontend:latest \
                     -f frontend/Dockerfile ./frontend
        
        # Build LilyPond service
        docker build -t $REGISTRY/$PROJECT_ID/muse-ai-lilypond:$IMAGE_TAG \
                     -t $REGISTRY/$PROJECT_ID/muse-ai-lilypond:latest \
                     -f lilypond-service/Dockerfile ./lilypond-service
        
        # Build nginx
        docker build -t $REGISTRY/$PROJECT_ID/muse-ai-nginx:$IMAGE_TAG \
                     -t $REGISTRY/$PROJECT_ID/muse-ai-nginx:latest \
                     -f nginx/Dockerfile ./nginx

    - name: Push Docker images
      run: |
        docker push $REGISTRY/$PROJECT_ID/muse-ai-backend:$IMAGE_TAG
        docker push $REGISTRY/$PROJECT_ID/muse-ai-backend:latest
        docker push $REGISTRY/$PROJECT_ID/muse-ai-frontend:$IMAGE_TAG
        docker push $REGISTRY/$PROJECT_ID/muse-ai-frontend:latest
        docker push $REGISTRY/$PROJECT_ID/muse-ai-lilypond:$IMAGE_TAG
        docker push $REGISTRY/$PROJECT_ID/muse-ai-lilypond:latest
        docker push $REGISTRY/$PROJECT_ID/muse-ai-nginx:$IMAGE_TAG
        docker push $REGISTRY/$PROJECT_ID/muse-ai-nginx:latest

    - name: Get GKE credentials
      run: |
        gcloud container clusters get-credentials $GKE_CLUSTER --region $GKE_REGION

    - name: Set target namespace
      run: |
        if [ "$BRANCH_NAME" = "main" ]; then
          echo "TARGET_NAMESPACE=muse-ai" >> $GITHUB_ENV
          echo "ENVIRONMENT=production" >> $GITHUB_ENV
        else
          echo "TARGET_NAMESPACE=muse-ai-staging" >> $GITHUB_ENV
          echo "ENVIRONMENT=staging" >> $GITHUB_ENV
        fi

    - name: Create namespace if not exists
      run: |
        kubectl create namespace $TARGET_NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

    - name: Create/update secrets
      run: |
        kubectl create secret generic muse-ai-secrets \
          --namespace=$TARGET_NAMESPACE \
          --from-literal=POSTGRES_PASSWORD="${{ secrets.POSTGRES_PASSWORD }}" \
          --from-literal=ANTHROPIC_API_KEY="${{ secrets.ANTHROPIC_API_KEY }}" \
          --dry-run=client -o yaml | kubectl apply -f -

    - name: Deploy to Kubernetes
      run: |
        # Generate manifests with current image tags
        envsubst < k8s/templates/deployment.yaml | kubectl apply -f -
      env:
        IMAGE_TAG: ${{ env.IMAGE_TAG }}
        TARGET_NAMESPACE: ${{ env.TARGET_NAMESPACE }}
        PROJECT_ID: ${{ env.PROJECT_ID }}

    - name: Verify deployment
      run: |
        kubectl rollout status deployment/backend -n $TARGET_NAMESPACE --timeout=300s
        kubectl rollout status deployment/frontend -n $TARGET_NAMESPACE --timeout=300s
        kubectl rollout status deployment/lilypond -n $TARGET_NAMESPACE --timeout=300s
        kubectl rollout status deployment/nginx -n $TARGET_NAMESPACE --timeout=300s

    - name: Run health checks
      run: |
        # Wait for services to be ready
        sleep 30
        
        # Get service IP
        SERVICE_IP=$(kubectl get service nginx-service -n $TARGET_NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        
        # Test health endpoints
        curl -f http://$SERVICE_IP/api/health || exit 1
        curl -f http://$SERVICE_IP/ || exit 1

    - name: Post deployment notification
      if: always()
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        channel: '#deployments'
        text: |
          Deployment to ${{ env.ENVIRONMENT }} ${{ job.status }}!
          Branch: ${{ github.ref }}
          Commit: ${{ github.sha }}
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

### Create Deployment Templates

Create `k8s/templates/deployment.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: muse-ai-config
  namespace: ${TARGET_NAMESPACE}
data:
  DATABASE_URL: "postgres://muse_ai:${POSTGRES_PASSWORD}@127.0.0.1:5432/muse_ai?sslmode=require"
  GCS_BUCKET: "${PROJECT_ID}-muse-ai-files"
  POSTGRES_DB: "muse_ai"
  POSTGRES_USER: "muse_ai"
  GO_ENV: "production"
  PORT: "8082"
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: muse-ai-backend-sa
  namespace: ${TARGET_NAMESPACE}
  annotations:
    iam.gke.io/gcp-service-account: muse-ai-backend@${PROJECT_ID}.iam.gserviceaccount.com
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: ${TARGET_NAMESPACE}
  labels:
    app: backend
    version: ${IMAGE_TAG}
spec:
  replicas: 2
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
        version: ${IMAGE_TAG}
    spec:
      serviceAccountName: muse-ai-backend-sa
      containers:
      - name: backend
        image: gcr.io/${PROJECT_ID}/muse-ai-backend:${IMAGE_TAG}
        ports:
        - containerPort: 8081
        env:
        - name: DATABASE_URL
          valueFrom:
            configMapKeyRef:
              name: muse-ai-config
              key: DATABASE_URL
        - name: GCS_BUCKET
          valueFrom:
            configMapKeyRef:
              name: muse-ai-config
              key: GCS_BUCKET
        - name: ANTHROPIC_API_KEY
          valueFrom:
            secretKeyRef:
              name: muse-ai-secrets
              key: ANTHROPIC_API_KEY
        - name: GO_ENV
          valueFrom:
            configMapKeyRef:
              name: muse-ai-config
              key: GO_ENV
        readinessProbe:
          httpGet:
            path: /api/health
            port: 8081
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        securityContext:
          runAsNonRoot: true
          runAsUser: 1000
          runAsGroup: 1000
          allowPrivilegeEscalation: false
      - name: cloud-sql-proxy
        image: gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.8.0
        args:
        - "--structured-logs"
        - "--port=5432"
        - "${PROJECT_ID}:${GKE_REGION}:muse-ai-postgres"
        securityContext:
          runAsNonRoot: true
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: ${TARGET_NAMESPACE}
  labels:
    app: frontend
    version: ${IMAGE_TAG}
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
        version: ${IMAGE_TAG}
    spec:
      containers:
      - name: frontend
        image: gcr.io/${PROJECT_ID}/muse-ai-frontend:${IMAGE_TAG}
        ports:
        - containerPort: 3000
        readinessProbe:
          httpGet:
            path: /
            port: 3000
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /
            port: 3000
          initialDelaySeconds: 30
          periodSeconds: 10
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        securityContext:
          runAsNonRoot: true
          runAsUser: 1000
          runAsGroup: 1000
          allowPrivilegeEscalation: false
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lilypond
  namespace: ${TARGET_NAMESPACE}
  labels:
    app: lilypond
    version: ${IMAGE_TAG}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: lilypond
  template:
    metadata:
      labels:
        app: lilypond
        version: ${IMAGE_TAG}
    spec:
      containers:
      - name: lilypond
        image: gcr.io/${PROJECT_ID}/muse-ai-lilypond:${IMAGE_TAG}
        ports:
        - containerPort: 8082
        env:
        - name: PORT
          valueFrom:
            configMapKeyRef:
              name: muse-ai-config
              key: PORT
        readinessProbe:
          httpGet:
            path: /health
            port: 8082
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /health
            port: 8082
          initialDelaySeconds: 30
          periodSeconds: 10
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        securityContext:
          runAsNonRoot: true
          runAsUser: 1001
          runAsGroup: 1001
          allowPrivilegeEscalation: false
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: ${TARGET_NAMESPACE}
  labels:
    app: nginx
    version: ${IMAGE_TAG}
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
        version: ${IMAGE_TAG}
    spec:
      containers:
      - name: nginx
        image: gcr.io/${PROJECT_ID}/muse-ai-nginx:${IMAGE_TAG}
        ports:
        - containerPort: 80
        readinessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /
            port: 80
          initialDelaySeconds: 10
          periodSeconds: 10
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "128Mi"
            cpu: "200m"
        securityContext:
          runAsNonRoot: true
          runAsUser: 101
          runAsGroup: 101
          allowPrivilegeEscalation: false
---
apiVersion: v1
kind: Service
metadata:
  name: backend-service
  namespace: ${TARGET_NAMESPACE}
spec:
  selector:
    app: backend
  ports:
  - port: 8081
    targetPort: 8081
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: frontend-service
  namespace: ${TARGET_NAMESPACE}
spec:
  selector:
    app: frontend
  ports:
  - port: 3000
    targetPort: 3000
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: lilypond-service
  namespace: ${TARGET_NAMESPACE}
spec:
  selector:
    app: lilypond
  ports:
  - port: 8082
    targetPort: 8082
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
  namespace: ${TARGET_NAMESPACE}
spec:
  selector:
    app: nginx
  ports:
  - port: 80
    targetPort: 80
  type: LoadBalancer
```

## Step 4: Create Additional Workflows

### Staging Deployment Workflow

Create `.github/workflows/staging-deploy.yml`:

```yaml
name: Deploy to Staging

on:
  pull_request:
    branches: [ main ]
    types: [ opened, synchronize, reopened ]

env:
  PROJECT_ID: ${{ secrets.GCP_PROJECT_ID }}
  GKE_CLUSTER: ${{ secrets.GKE_CLUSTER_NAME }}
  GKE_REGION: ${{ secrets.GKE_CLUSTER_REGION }}
  REGISTRY: gcr.io
  NAMESPACE: muse-ai-staging

jobs:
  deploy-staging:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    environment: staging

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Cloud SDK
      uses: google-github-actions/setup-gcloud@v1
      with:
        service_account_key: ${{ secrets.GCP_SA_KEY }}
        project_id: ${{ secrets.GCP_PROJECT_ID }}

    - name: Configure Docker for GCR
      run: gcloud auth configure-docker

    - name: Set environment variables
      run: |
        echo "IMAGE_TAG=pr-${{ github.event.number }}-$(echo $GITHUB_SHA | cut -c1-7)" >> $GITHUB_ENV
        echo "PR_NUMBER=${{ github.event.number }}" >> $GITHUB_ENV

    - name: Build and push Docker images
      run: |
        docker build -t $REGISTRY/$PROJECT_ID/muse-ai-backend:$IMAGE_TAG -f Dockerfile .
        docker build -t $REGISTRY/$PROJECT_ID/muse-ai-frontend:$IMAGE_TAG -f frontend/Dockerfile ./frontend
        docker build -t $REGISTRY/$PROJECT_ID/muse-ai-lilypond:$IMAGE_TAG -f lilypond-service/Dockerfile ./lilypond-service
        docker build -t $REGISTRY/$PROJECT_ID/muse-ai-nginx:$IMAGE_TAG -f nginx/Dockerfile ./nginx
        
        docker push $REGISTRY/$PROJECT_ID/muse-ai-backend:$IMAGE_TAG
        docker push $REGISTRY/$PROJECT_ID/muse-ai-frontend:$IMAGE_TAG
        docker push $REGISTRY/$PROJECT_ID/muse-ai-lilypond:$IMAGE_TAG
        docker push $REGISTRY/$PROJECT_ID/muse-ai-nginx:$IMAGE_TAG

    - name: Get GKE credentials
      run: |
        gcloud container clusters get-credentials $GKE_CLUSTER --region $GKE_REGION

    - name: Create staging namespace
      run: |
        kubectl create namespace muse-ai-staging --dry-run=client -o yaml | kubectl apply -f -

    - name: Deploy to staging
      run: |
        envsubst < k8s/templates/deployment.yaml | kubectl apply -f -
      env:
        IMAGE_TAG: ${{ env.IMAGE_TAG }}
        TARGET_NAMESPACE: muse-ai-staging
        PROJECT_ID: ${{ env.PROJECT_ID }}

    - name: Wait for deployment
      run: |
        kubectl rollout status deployment/backend -n muse-ai-staging --timeout=300s
        kubectl rollout status deployment/frontend -n muse-ai-staging --timeout=300s

    - name: Get staging URL
      id: get-url
      run: |
        EXTERNAL_IP=$(kubectl get service nginx-service -n muse-ai-staging -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        echo "staging_url=http://$EXTERNAL_IP" >> $GITHUB_OUTPUT

    - name: Comment PR with staging URL
      uses: actions/github-script@v6
      with:
        script: |
          github.rest.issues.createComment({
            issue_number: context.issue.number,
            owner: context.repo.owner,
            repo: context.repo.repo,
            body: '🚀 Staging deployment ready!\n\n**URL:** ${{ steps.get-url.outputs.staging_url }}\n**Image Tag:** `${{ env.IMAGE_TAG }}`'
          })
```

### Production Rollback Workflow

Create `.github/workflows/rollback.yml`:

```yaml
name: Production Rollback

on:
  workflow_dispatch:
    inputs:
      image_tag:
        description: 'Image tag to rollback to'
        required: true
        type: string

env:
  PROJECT_ID: ${{ secrets.GCP_PROJECT_ID }}
  GKE_CLUSTER: ${{ secrets.GKE_CLUSTER_NAME }}
  GKE_REGION: ${{ secrets.GKE_CLUSTER_REGION }}

jobs:
  rollback:
    name: Rollback Production
    runs-on: ubuntu-latest
    environment: production

    steps:
    - name: Checkout code
      uses: actions/checkout@v4

    - name: Set up Cloud SDK
      uses: google-github-actions/setup-gcloud@v1
      with:
        service_account_key: ${{ secrets.GCP_SA_KEY }}
        project_id: ${{ secrets.GCP_PROJECT_ID }}

    - name: Get GKE credentials
      run: |
        gcloud container clusters get-credentials $GKE_CLUSTER --region $GKE_REGION

    - name: Rollback deployments
      run: |
        kubectl set image deployment/backend backend=gcr.io/$PROJECT_ID/muse-ai-backend:${{ github.event.inputs.image_tag }} -n muse-ai
        kubectl set image deployment/frontend frontend=gcr.io/$PROJECT_ID/muse-ai-frontend:${{ github.event.inputs.image_tag }} -n muse-ai
        kubectl set image deployment/lilypond lilypond=gcr.io/$PROJECT_ID/muse-ai-lilypond:${{ github.event.inputs.image_tag }} -n muse-ai
        kubectl set image deployment/nginx nginx=gcr.io/$PROJECT_ID/muse-ai-nginx:${{ github.event.inputs.image_tag }} -n muse-ai

    - name: Wait for rollback
      run: |
        kubectl rollout status deployment/backend -n muse-ai --timeout=300s
        kubectl rollout status deployment/frontend -n muse-ai --timeout=300s
        kubectl rollout status deployment/lilypond -n muse-ai --timeout=300s
        kubectl rollout status deployment/nginx -n muse-ai --timeout=300s

    - name: Verify rollback
      run: |
        kubectl get deployments -n muse-ai -o wide
```

## Step 5: Set Up Monitoring and Alerts

### Create Monitoring Workflow

Create `.github/workflows/monitoring.yml`:

```yaml
name: Post-Deployment Monitoring

on:
  workflow_run:
    workflows: ["CI/CD Pipeline"]
    types:
      - completed

jobs:
  monitor:
    name: Monitor Deployment
    runs-on: ubuntu-latest
    if: ${{ github.event.workflow_run.conclusion == 'success' }}

    steps:
    - name: Set up Cloud SDK
      uses: google-github-actions/setup-gcloud@v1
      with:
        service_account_key: ${{ secrets.GCP_SA_KEY }}
        project_id: ${{ secrets.GCP_PROJECT_ID }}

    - name: Get GKE credentials
      run: |
        gcloud container clusters get-credentials ${{ secrets.GKE_CLUSTER_NAME }} --region ${{ secrets.GKE_CLUSTER_REGION }}

    - name: Check deployment health
      run: |
        # Wait for pods to stabilize
        sleep 60
        
        # Check pod status
        kubectl get pods -n muse-ai
        
        # Check if all pods are ready
        kubectl wait --for=condition=ready pod -l app=backend -n muse-ai --timeout=300s
        kubectl wait --for=condition=ready pod -l app=frontend -n muse-ai --timeout=300s
        kubectl wait --for=condition=ready pod -l app=lilypond -n muse-ai --timeout=300s
        kubectl wait --for=condition=ready pod -l app=nginx -n muse-ai --timeout=300s

    - name: Run health checks
      run: |
        EXTERNAL_IP=$(kubectl get service nginx-service -n muse-ai -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
        
        # Test endpoints
        curl -f http://$EXTERNAL_IP/api/health
        curl -f http://$EXTERNAL_IP/
        
        # Check response times
        curl -w "Response time: %{time_total}s\n" -o /dev/null -s http://$EXTERNAL_IP/api/health

    - name: Alert on failure
      if: failure()
      uses: 8398a7/action-slack@v3
      with:
        status: failure
        channel: '#alerts'
        text: |
          🚨 Post-deployment health check failed!
          Please investigate the production environment immediately.
      env:
        SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

## Step 6: Create Deployment Scripts

### Create Deployment Script

Create `scripts/deploy.sh`:

```bash
#!/bin/bash

set -e

# Configuration
PROJECT_ID=${PROJECT_ID:-"your-project-id"}
CLUSTER_NAME=${CLUSTER_NAME:-"muse-ai-cluster"}
CLUSTER_REGION=${CLUSTER_REGION:-"us-central1"}
NAMESPACE=${NAMESPACE:-"muse-ai"}
IMAGE_TAG=${IMAGE_TAG:-"latest"}

echo "🚀 Starting deployment..."
echo "Project: $PROJECT_ID"
echo "Cluster: $CLUSTER_NAME"
echo "Namespace: $NAMESPACE"
echo "Image Tag: $IMAGE_TAG"

# Get cluster credentials
echo "📡 Getting cluster credentials..."
gcloud container clusters get-credentials $CLUSTER_NAME --region $CLUSTER_REGION

# Create namespace if it doesn't exist
echo "🏗️  Creating namespace..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Apply configurations
echo "⚙️  Applying configurations..."
envsubst < k8s/templates/deployment.yaml | kubectl apply -f -

# Wait for deployments to be ready
echo "⏳ Waiting for deployments..."
kubectl rollout status deployment/backend -n $NAMESPACE --timeout=300s
kubectl rollout status deployment/frontend -n $NAMESPACE --timeout=300s
kubectl rollout status deployment/lilypond -n $NAMESPACE --timeout=300s
kubectl rollout status deployment/nginx -n $NAMESPACE --timeout=300s

# Get service URL
EXTERNAL_IP=$(kubectl get service nginx-service -n $NAMESPACE -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

echo "✅ Deployment completed successfully!"
echo "🌐 Application URL: http://$EXTERNAL_IP"

# Run health check
echo "🔍 Running health check..."
curl -f http://$EXTERNAL_IP/api/health || echo "❌ Health check failed"

echo "🎉 Deployment finished!"
```

### Make Script Executable

```bash
chmod +x scripts/deploy.sh
```

## Step 7: Environment-Specific Configurations

### Production Environment Variables

Create `.github/environments/production.yml`:

```yaml
name: production
url: https://muse-ai.yourdomain.com
protection_rules:
  - type: required_reviewers
    required_reviewers:
      - team: devops-team
  - type: wait_timer
    wait_timer: 5
variables:
  DEPLOYMENT_TIMEOUT: "600"
  HEALTH_CHECK_RETRIES: "5"
secrets:
  DATABASE_PASSWORD: ${{ secrets.PROD_DATABASE_PASSWORD }}
  ANTHROPIC_API_KEY: ${{ secrets.PROD_ANTHROPIC_API_KEY }}
```

### Staging Environment Variables

Create `.github/environments/staging.yml`:

```yaml
name: staging
url: https://staging.muse-ai.yourdomain.com
variables:
  DEPLOYMENT_TIMEOUT: "300"
  HEALTH_CHECK_RETRIES: "3"
secrets:
  DATABASE_PASSWORD: ${{ secrets.STAGING_DATABASE_PASSWORD }}
  ANTHROPIC_API_KEY: ${{ secrets.STAGING_ANTHROPIC_API_KEY }}
```

## Step 8: Testing the Pipeline

### Trigger Deployment

1. **Push to main branch** to trigger production deployment
2. **Create pull request** to trigger staging deployment
3. **Use manual rollback** if needed

### Monitor Deployment

```bash
# Watch GitHub Actions
# Go to: https://github.com/your-username/muse-ai/actions

# Monitor Kubernetes deployment
kubectl get pods -n muse-ai -w

# Check application logs
kubectl logs -f deployment/backend -n muse-ai
```

## Step 9: Best Practices and Security

### Security Considerations

1. **Least Privilege**: Service accounts have minimal required permissions
2. **Secret Management**: Sensitive data stored in GitHub Secrets
3. **Image Scanning**: Add vulnerability scanning to pipeline
4. **Network Policies**: Restrict pod-to-pod communication
5. **RBAC**: Implement role-based access control

### Performance Optimization

1. **Image Caching**: Use Docker layer caching in CI/CD
2. **Parallel Builds**: Build images in parallel
3. **Registry Cleanup**: Clean up old images regularly
4. **Resource Limits**: Set appropriate CPU/memory limits

### Monitoring and Alerting

1. **Health Checks**: Comprehensive endpoint monitoring
2. **Slack Notifications**: Real-time deployment status
3. **Log Aggregation**: Centralized logging with GCP Logging
4. **Metrics Collection**: Use Prometheus and Grafana

## Next Steps

1. **Set up SSL certificates** for HTTPS
2. **Configure custom domains** with Cloud DNS
3. **Implement blue-green deployments** for zero-downtime
4. **Add integration tests** to the pipeline
5. **Set up proper monitoring** and alerting
6. **Implement backup strategies** for data

## Troubleshooting

### Common Issues

1. **Authentication Errors**: Verify service account permissions
2. **Image Pull Errors**: Check registry access and image tags
3. **Deployment Timeouts**: Increase timeout values and check resource limits
4. **Health Check Failures**: Verify endpoint availability and network policies

### Debug Commands

```bash
# Check pipeline logs
gh run list
gh run view [run-id]

# Check Kubernetes resources
kubectl describe pod [pod-name] -n muse-ai
kubectl logs [pod-name] -n muse-ai
kubectl get events -n muse-ai --sort-by='.lastTimestamp'

# Test service connectivity
kubectl exec -it deployment/backend -n muse-ai -- curl http://frontend-service:3000
```

This automated deployment setup provides a robust CI/CD pipeline for the Muse AI application with proper testing, security, and monitoring capabilities.
