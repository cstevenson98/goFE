# Google Cloud Platform (GCP) Manual Deployment Guide

This guide covers manual deployment of the Muse AI application to Google Kubernetes Engine (GKE) on Google Cloud Platform.

## Prerequisites

### Required Tools
- **Google Cloud SDK (gcloud)**: [Installation guide](https://cloud.google.com/sdk/docs/install)
- **kubectl**: Kubernetes command-line tool
- **Docker**: For building and pushing images
- **GCP Account**: With billing enabled

### Required GCP APIs
Enable the following APIs in your GCP project:
```bash
gcloud services enable \
  container.googleapis.com \
  containerregistry.googleapis.com \
  cloudsql.googleapis.com \
  compute.googleapis.com \
  storage-component.googleapis.com
```

## Step 1: Set Up GCP Project and Authentication

### Configure gcloud

```bash
# Login to Google Cloud
gcloud auth login

# List available projects
gcloud projects list

# Set your project ID
export PROJECT_ID="your-project-id"
gcloud config set project $PROJECT_ID

# Enable required APIs
gcloud services enable container.googleapis.com
gcloud services enable containerregistry.googleapis.com
gcloud services enable artifactregistry.googleapis.com

# Configure Docker to use gcloud as credential helper
gcloud auth configure-docker
```

### Set Environment Variables

```bash
# Project configuration
export PROJECT_ID="your-project-id"
export REGION="us-central1"
export ZONE="us-central1-a"
export CLUSTER_NAME="muse-ai-cluster"

# Application configuration
export NAMESPACE="muse-ai"
export ANTHROPIC_API_KEY="your-anthropic-api-key"
```

## Step 2: Create Google Kubernetes Engine (GKE) Cluster

### Create GKE Cluster

```bash
# Create a regional GKE cluster with auto-scaling
gcloud container clusters create $CLUSTER_NAME \
  --region=$REGION \
  --num-nodes=1 \
  --min-nodes=1 \
  --max-nodes=5 \
  --enable-autoscaling \
  --machine-type=e2-standard-2 \
  --disk-size=50GB \
  --disk-type=pd-ssd \
  --enable-autorepair \
  --enable-autoupgrade \
  --enable-ip-alias \
  --enable-network-policy \
  --addons=HorizontalPodAutoscaling,HttpLoadBalancing,NetworkPolicy

# Get cluster credentials
gcloud container clusters get-credentials $CLUSTER_NAME --region=$REGION

# Verify cluster connection
kubectl cluster-info
kubectl get nodes
```

### Configure kubectl Context

```bash
# Set current context
kubectl config current-context

# Create namespace
kubectl create namespace $NAMESPACE
```

## Step 3: Set Up Container Registry

### Option A: Google Container Registry (GCR)

```bash
# Configure Docker for GCR
gcloud auth configure-docker

# Set registry variables
export REGISTRY="gcr.io/$PROJECT_ID"
```

### Option B: Google Artifact Registry (Recommended)

```bash
# Create Artifact Registry repository
gcloud artifacts repositories create muse-ai-images \
  --repository-format=docker \
  --location=$REGION \
  --description="Muse AI Docker images"

# Configure Docker for Artifact Registry
gcloud auth configure-docker $REGION-docker.pkg.dev

# Set registry variables
export REGISTRY="$REGION-docker.pkg.dev/$PROJECT_ID/muse-ai-images"
```

## Step 4: Build and Push Docker Images

### Build and Tag Images

```bash
# Navigate to project root
cd /path/to/muse-ai

# Build and tag images with registry prefix
docker build -t $REGISTRY/muse-ai-backend:latest -f Dockerfile .
docker build -t $REGISTRY/muse-ai-frontend:latest -f frontend/Dockerfile ./frontend
docker build -t $REGISTRY/muse-ai-lilypond:latest -f lilypond-service/Dockerfile ./lilypond-service
docker build -t $REGISTRY/muse-ai-nginx:latest -f nginx/Dockerfile ./nginx

# Verify images
docker images | grep $REGISTRY
```

### Push Images to Registry

```bash
# Push all images
docker push $REGISTRY/muse-ai-backend:latest
docker push $REGISTRY/muse-ai-frontend:latest
docker push $REGISTRY/muse-ai-lilypond:latest
docker push $REGISTRY/muse-ai-nginx:latest

# Verify images in registry
gcloud container images list --repository=$REGISTRY
```

## Step 5: Set Up Cloud Storage and Database

### Option A: Use Managed Services (Recommended)

#### Create Cloud SQL PostgreSQL Instance

```bash
# Create Cloud SQL instance
gcloud sql instances create muse-ai-postgres \
  --database-version=POSTGRES_15 \
  --tier=db-f1-micro \
  --region=$REGION \
  --storage-size=20GB \
  --storage-type=SSD \
  --backup-start-time=02:00 \
  --enable-bin-log

# Create database
gcloud sql databases create muse_ai --instance=muse-ai-postgres

# Create user
gcloud sql users create muse_ai \
  --instance=muse-ai-postgres \
  --password=your-secure-password

# Get connection name
gcloud sql instances describe muse-ai-postgres --format="value(connectionName)"
```

#### Create Cloud Storage Bucket

```bash
# Create bucket for file storage
gsutil mb gs://$PROJECT_ID-muse-ai-files

# Set bucket permissions (adjust as needed)
gsutil iam ch allUsers:objectViewer gs://$PROJECT_ID-muse-ai-files
```

### Option B: Use In-Cluster Services

If you prefer to run PostgreSQL and MinIO in the cluster, continue with the Kubernetes manifests approach.

## Step 6: Create Kubernetes Manifests for GCP

### Create Production Manifests

```bash
mkdir -p k8s/gcp
cd k8s/gcp
```

### ConfigMap for GCP Environment

```yaml
# config-gcp.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: muse-ai-config
  namespace: muse-ai
data:
  # For Cloud SQL
  DATABASE_URL: "postgres://muse_ai:your-secure-password@127.0.0.1:5432/muse_ai?sslmode=require"
  # For in-cluster PostgreSQL
  # DATABASE_URL: "postgres://muse_ai:muse_ai_pass@postgres-service:5432/muse_ai?sslmode=disable"
  
  # For Cloud Storage
  GCS_BUCKET: "$PROJECT_ID-muse-ai-files"
  # For in-cluster MinIO
  # MINIO_ENDPOINT: "minio-service:9000"
  # MINIO_BUCKET: "muse-ai-files"
  
  POSTGRES_DB: "muse_ai"
  POSTGRES_USER: "muse_ai"
  GO_ENV: "production"
  PORT: "8082"
```

### Secrets for GCP

```yaml
# secrets-gcp.yaml
apiVersion: v1
kind: Secret
metadata:
  name: muse-ai-secrets
  namespace: muse-ai
type: Opaque
data:
  POSTGRES_PASSWORD: eW91ci1zZWN1cmUtcGFzc3dvcmQ=  # your-secure-password (base64)
  ANTHROPIC_API_KEY: ""  # your-anthropic-api-key (base64 encoded)
  # For MinIO if using in-cluster
  # MINIO_ACCESS_KEY: bXVzZV9haQ==
  # MINIO_SECRET_KEY: bXVzZV9haV9wYXNz
```

### Backend Deployment with Cloud SQL Proxy

```yaml
# backend-deployment-gcp.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: muse-ai
spec:
  replicas: 2
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      serviceAccountName: muse-ai-backend-sa
      containers:
      - name: backend
        image: gcr.io/PROJECT_ID/muse-ai-backend:latest  # Replace PROJECT_ID
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
      # Cloud SQL Proxy sidecar
      - name: cloud-sql-proxy
        image: gcr.io/cloud-sql-connectors/cloud-sql-proxy:2.8.0
        args:
        - "--structured-logs"
        - "--port=5432"
        - "PROJECT_ID:REGION:muse-ai-postgres"  # Replace with your values
        securityContext:
          runAsNonRoot: true
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"
```

### Service Account for GCP Access

```yaml
# service-account.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: muse-ai-backend-sa
  namespace: muse-ai
  annotations:
    iam.gke.io/gcp-service-account: muse-ai-backend@PROJECT_ID.iam.gserviceaccount.com
```

### Complete Application Deployment

```yaml
# app-deployment-gcp.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: muse-ai
---
# Include ConfigMap, Secrets, and Service Account from above
---
# LilyPond Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lilypond
  namespace: muse-ai
spec:
  replicas: 1
  selector:
    matchLabels:
      app: lilypond
  template:
    metadata:
      labels:
        app: lilypond
    spec:
      containers:
      - name: lilypond
        image: gcr.io/PROJECT_ID/muse-ai-lilypond:latest  # Replace PROJECT_ID
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
# Frontend Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: muse-ai
spec:
  replicas: 2
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: frontend
        image: gcr.io/PROJECT_ID/muse-ai-frontend:latest  # Replace PROJECT_ID
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
# Nginx Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: muse-ai
spec:
  replicas: 2
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: gcr.io/PROJECT_ID/muse-ai-nginx:latest  # Replace PROJECT_ID
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
# Services
apiVersion: v1
kind: Service
metadata:
  name: backend-service
  namespace: muse-ai
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
  name: lilypond-service
  namespace: muse-ai
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
  name: frontend-service
  namespace: muse-ai
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
  name: nginx-service
  namespace: muse-ai
spec:
  selector:
    app: nginx
  ports:
  - port: 80
    targetPort: 80
  type: LoadBalancer
```

## Step 7: Configure GCP IAM and Service Accounts

### Create GCP Service Account

```bash
# Create service account for the application
gcloud iam service-accounts create muse-ai-backend \
  --display-name="Muse AI Backend Service Account"

# Grant necessary permissions
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:muse-ai-backend@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:muse-ai-backend@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/storage.admin"

# Enable Workload Identity
gcloud iam service-accounts add-iam-policy-binding muse-ai-backend@$PROJECT_ID.iam.gserviceaccount.com \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:$PROJECT_ID.svc.id.goog[muse-ai/muse-ai-backend-sa]"
```

### Enable Workload Identity on Cluster

```bash
# Update cluster to enable Workload Identity
gcloud container clusters update $CLUSTER_NAME \
  --region=$REGION \
  --workload-pool=$PROJECT_ID.svc.id.goog
```

## Step 8: Deploy to GCP

### Update Manifests with Your Project ID

```bash
# Replace PROJECT_ID placeholder in manifests
sed -i "s/PROJECT_ID/$PROJECT_ID/g" *.yaml
```

### Apply Secrets (Base64 Encode Your Values)

```bash
# Create secrets with your actual values
kubectl create secret generic muse-ai-secrets \
  --namespace=muse-ai \
  --from-literal=POSTGRES_PASSWORD="your-secure-password" \
  --from-literal=ANTHROPIC_API_KEY="your-anthropic-api-key"
```

### Deploy Application

```bash
# Apply all manifests
kubectl apply -f config-gcp.yaml
kubectl apply -f service-account.yaml
kubectl apply -f app-deployment-gcp.yaml

# Watch deployment
kubectl get pods -n muse-ai -w
```

### Configure Ingress (Optional)

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: muse-ai-ingress
  namespace: muse-ai
  annotations:
    kubernetes.io/ingress.class: "gce"
    kubernetes.io/ingress.global-static-ip-name: "muse-ai-ip"
    networking.gke.io/managed-certificates: "muse-ai-ssl-cert"
spec:
  rules:
  - host: yourdomain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: nginx-service
            port:
              number: 80
---
apiVersion: networking.gke.io/v1
kind: ManagedCertificate
metadata:
  name: muse-ai-ssl-cert
  namespace: muse-ai
spec:
  domains:
  - yourdomain.com
```

### Reserve Static IP

```bash
# Reserve global static IP
gcloud compute addresses create muse-ai-ip --global

# Get the IP address
gcloud compute addresses describe muse-ai-ip --global --format="value(address)"
```

## Step 9: Verify Deployment

### Check Application Status

```bash
# Check all pods are running
kubectl get pods -n muse-ai

# Check services
kubectl get services -n muse-ai

# Get external IP
kubectl get service nginx-service -n muse-ai

# Check logs
kubectl logs -n muse-ai deployment/backend
kubectl logs -n muse-ai deployment/frontend
kubectl logs -n muse-ai deployment/lilypond
```

### Test Application

```bash
# Get external IP
EXTERNAL_IP=$(kubectl get service nginx-service -n muse-ai -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

# Test API endpoint
curl http://$EXTERNAL_IP/api/health

# Access application
echo "Application available at: http://$EXTERNAL_IP"
```

## Step 10: Set Up Monitoring and Logging

### Enable GKE Monitoring

```bash
# Update cluster to enable monitoring
gcloud container clusters update $CLUSTER_NAME \
  --region=$REGION \
  --enable-cloud-logging \
  --enable-cloud-monitoring
```

### Configure Horizontal Pod Autoscaling

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: backend-hpa
  namespace: muse-ai
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: backend
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: frontend-hpa
  namespace: muse-ai
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: frontend
  minReplicas: 2
  maxReplicas: 8
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

```bash
# Apply HPA configuration
kubectl apply -f hpa.yaml
```

## Step 11: Backup and Maintenance

### Database Backups

```bash
# Create automatic backup policy for Cloud SQL
gcloud sql instances patch muse-ai-postgres \
  --backup-start-time=02:00 \
  --backup-location=$REGION
```

### Update Application

```bash
# Build new image version
docker build -t $REGISTRY/muse-ai-backend:v1.1.0 -f Dockerfile .
docker push $REGISTRY/muse-ai-backend:v1.1.0

# Update deployment
kubectl set image deployment/backend backend=$REGISTRY/muse-ai-backend:v1.1.0 -n muse-ai

# Watch rollout
kubectl rollout status deployment/backend -n muse-ai
```

## Cleanup

### Delete Application

```bash
# Delete application resources
kubectl delete namespace muse-ai

# Delete GKE cluster
gcloud container clusters delete $CLUSTER_NAME --region=$REGION

# Delete Cloud SQL instance
gcloud sql instances delete muse-ai-postgres

# Delete Cloud Storage bucket
gsutil rm -r gs://$PROJECT_ID-muse-ai-files

# Delete static IP
gcloud compute addresses delete muse-ai-ip --global

# Delete service account
gcloud iam service-accounts delete muse-ai-backend@$PROJECT_ID.iam.gserviceaccount.com
```

## Next Steps

- Set up automated deployment with [`gcp-automated-deployment.md`](./gcp-automated-deployment.md)
- Implement proper monitoring and alerting
- Set up SSL certificates and custom domains
- Configure backup and disaster recovery procedures
- Implement blue-green or canary deployments
