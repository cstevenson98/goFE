# Local Kubernetes Deployment Guide with minikube

This guide walks you through deploying the Muse AI application on a local Kubernetes cluster using minikube.

## Prerequisites

### Required Tools
- **Docker**: For building container images
- **kubectl**: Kubernetes command-line tool
- **minikube**: Local Kubernetes cluster - [Installation guide](https://minikube.sigs.k8s.io/docs/start/)

### System Requirements
- **CPU**: 4+ cores recommended
- **Memory**: 8GB+ RAM
- **Storage**: 20GB+ free space
- **OS**: Windows, macOS, or Linux

## Step 1: Set Up minikube Cluster

### Install minikube

```bash
# macOS (using Homebrew)
brew install minikube

# Linux
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube /usr/local/bin/

# Windows (using Chocolatey)
choco install minikube

# Verify installation
minikube version
```

### Start minikube Cluster

```bash
# Start minikube with sufficient resources
minikube start --cpus=4 --memory=8192 --disk-size=20g

# Enable useful addons
minikube addons enable ingress
minikube addons enable metrics-server

# Verify cluster is running
kubectl cluster-info
kubectl get nodes
```

## Step 2: Configure kubectl Context

```bash
# minikube automatically configures kubectl context
kubectl config current-context
# Should show: minikube

# Verify cluster connection
kubectl get nodes
```

## Step 3: Build Docker Images

### Configure Docker Environment

```bash
# Configure Docker to use minikube's Docker daemon
# This allows minikube to use locally built images
eval $(minikube docker-env)

# Verify Docker is pointing to minikube
docker ps
```

### Build All Images

```bash
# Navigate to project root
cd /path/to/muse-ai

# Build backend image
docker build -t muse-ai-backend:latest -f Dockerfile .

# Build frontend image
docker build -t muse-ai-frontend:latest -f frontend/Dockerfile ./frontend

# Build LilyPond service image
docker build -t muse-ai-lilypond:latest -f lilypond-service/Dockerfile ./lilypond-service

# Build nginx image
docker build -t muse-ai-nginx:latest -f nginx/Dockerfile ./nginx

# Verify images are available in minikube
docker images | grep muse-ai
```

## Step 4: Create Kubernetes Manifests

Create the manifest files directory:

```bash
mkdir -p k8s/local
cd k8s/local
```

### Create All-in-One Deployment File

```bash
cat <<EOF > muse-ai-local.yaml
# Namespace
apiVersion: v1
kind: Namespace
metadata:
  name: muse-ai
---
# ConfigMap
apiVersion: v1
kind: ConfigMap
metadata:
  name: muse-ai-config
  namespace: muse-ai
data:
  DATABASE_URL: "postgres://muse_ai:muse_ai_pass@postgres-service:5432/muse_ai?sslmode=disable"
  MINIO_ENDPOINT: "minio-service:9000"
  MINIO_BUCKET: "muse-ai-files"
  POSTGRES_DB: "muse_ai"
  POSTGRES_USER: "muse_ai"
  GO_ENV: "development"
  PORT: "8082"
---
# Secrets
apiVersion: v1
kind: Secret
metadata:
  name: muse-ai-secrets
  namespace: muse-ai
type: Opaque
data:
  POSTGRES_PASSWORD: bXVzZV9haV9wYXNz  # muse_ai_pass
  MINIO_ACCESS_KEY: bXVzZV9haQ==        # muse_ai
  MINIO_SECRET_KEY: bXVzZV9haV9wYXNz    # muse_ai_pass
  ANTHROPIC_API_KEY: ""                 # Add your API key here (base64 encoded)
---
# PostgreSQL PVC
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  namespace: muse-ai
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi
---
# MinIO PVC
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: minio-pvc
  namespace: muse-ai
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
---
# PostgreSQL StatefulSet
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: muse-ai
spec:
  serviceName: postgres-service
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:16-alpine
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_DB
          valueFrom:
            configMapKeyRef:
              name: muse-ai-config
              key: POSTGRES_DB
        - name: POSTGRES_USER
          valueFrom:
            configMapKeyRef:
              name: muse-ai-config
              key: POSTGRES_USER
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: muse-ai-secrets
              key: POSTGRES_PASSWORD
        - name: PGDATA
          value: /var/lib/postgresql/data/pgdata
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
---
# PostgreSQL Service
apiVersion: v1
kind: Service
metadata:
  name: postgres-service
  namespace: muse-ai
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
  type: ClusterIP
---
# MinIO StatefulSet
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: minio
  namespace: muse-ai
spec:
  serviceName: minio-service
  replicas: 1
  selector:
    matchLabels:
      app: minio
  template:
    metadata:
      labels:
        app: minio
    spec:
      containers:
      - name: minio
        image: minio/minio:latest
        command: ["minio", "server", "/data", "--console-address", ":9001"]
        ports:
        - containerPort: 9000
        - containerPort: 9001
        env:
        - name: MINIO_ROOT_USER
          valueFrom:
            secretKeyRef:
              name: muse-ai-secrets
              key: MINIO_ACCESS_KEY
        - name: MINIO_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: muse-ai-secrets
              key: MINIO_SECRET_KEY
        volumeMounts:
        - name: minio-storage
          mountPath: /data
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "200m"
      volumes:
      - name: minio-storage
        persistentVolumeClaim:
          claimName: minio-pvc
---
# MinIO Service
apiVersion: v1
kind: Service
metadata:
  name: minio-service
  namespace: muse-ai
spec:
  selector:
    app: minio
  ports:
  - name: api
    port: 9000
    targetPort: 9000
  - name: console
    port: 9001
    targetPort: 9001
  type: ClusterIP
---
# Backend Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: muse-ai
spec:
  replicas: 1
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: backend
        image: muse-ai-backend:latest
        imagePullPolicy: Never  # Use local images
        ports:
        - containerPort: 8081
        env:
        - name: DATABASE_URL
          valueFrom:
            configMapKeyRef:
              name: muse-ai-config
              key: DATABASE_URL
        - name: MINIO_ENDPOINT
          valueFrom:
            configMapKeyRef:
              name: muse-ai-config
              key: MINIO_ENDPOINT
        - name: MINIO_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: muse-ai-secrets
              key: MINIO_ACCESS_KEY
        - name: MINIO_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: muse-ai-secrets
              key: MINIO_SECRET_KEY
        - name: MINIO_BUCKET
          valueFrom:
            configMapKeyRef:
              name: muse-ai-config
              key: MINIO_BUCKET
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
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
---
# Backend Service
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
        image: muse-ai-lilypond:latest
        imagePullPolicy: Never  # Use local images
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
            memory: "256Mi"
            cpu: "200m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
# LilyPond Service
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
# Frontend Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  namespace: muse-ai
spec:
  replicas: 1
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
        image: muse-ai-frontend:latest
        imagePullPolicy: Never  # Use local images
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
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
---
# Frontend Service
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
# Nginx Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  namespace: muse-ai
spec:
  replicas: 1
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
        image: muse-ai-nginx:latest
        imagePullPolicy: Never  # Use local images
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
            memory: "32Mi"
            cpu: "50m"
          limits:
            memory: "64Mi"
            cpu: "100m"
---
# Nginx Service (NodePort for local access)
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
    nodePort: 30080
  type: NodePort
EOF
```

## Step 5: Deploy the Application

### Set up Anthropic API Key (if you have one)

```bash
# Encode your API key in base64
echo -n "your-anthropic-api-key-here" | base64

# Update the secret in the YAML file with the encoded value
# Edit muse-ai-local.yaml and replace the empty ANTHROPIC_API_KEY value
```

### Deploy to Kubernetes

```bash
# Apply the manifests
kubectl apply -f muse-ai-local.yaml

# Check deployment status
kubectl get pods -n muse-ai

# Watch pods start up
kubectl get pods -n muse-ai -w
```

### Wait for All Pods to be Ready

```bash
# Check all pods are running (this may take 2-5 minutes)
kubectl get pods -n muse-ai

# Check services
kubectl get services -n muse-ai

# Check persistent volumes
kubectl get pvc -n muse-ai
```

## Step 6: Access the Application

### Get Access URL with minikube

```bash
# Get the service URL directly
minikube service nginx-service -n muse-ai --url

# This will output something like: http://192.168.49.2:30080
# Open this URL in your browser

# Alternative: Use minikube tunnel for LoadBalancer services
# In a separate terminal, run:
minikube tunnel

# Then get the external IP:
kubectl get service nginx-service -n muse-ai

# Or use port forwarding
kubectl port-forward -n muse-ai service/nginx-service 8080:80
# Then access http://localhost:8080
```

### Access Individual Services (for debugging)

```bash
# Backend API
kubectl port-forward -n muse-ai service/backend-service 8081:8081
# Access http://localhost:8081/api/health

# LilyPond Service
kubectl port-forward -n muse-ai service/lilypond-service 8082:8082
# Access http://localhost:8082/health

# MinIO Console
kubectl port-forward -n muse-ai service/minio-service 9001:9001
# Access http://localhost:9001 (login: muse_ai / muse_ai_pass)

# Frontend directly
kubectl port-forward -n muse-ai service/frontend-service 3000:3000
# Access http://localhost:3000
```

## Step 7: Verify Deployment

### Check Application Health

```bash
# Check all pods are running
kubectl get pods -n muse-ai

# Check logs for any issues
kubectl logs -n muse-ai deployment/backend
kubectl logs -n muse-ai deployment/frontend
kubectl logs -n muse-ai deployment/lilypond
kubectl logs -n muse-ai statefulset/postgres
kubectl logs -n muse-ai statefulset/minio

# Test API endpoint
curl http://localhost:8081/api/health
```

### Database Initialization

```bash
# If you need to run database migrations
kubectl exec -it -n muse-ai deployment/backend -- ls -la scripts/
```

## Step 8: Development Workflow

### Rebuilding and Updating Images

```bash
# Make sure docker-env is set to use minikube's Docker daemon
eval $(minikube docker-env)

# Rebuild an image (example: backend)
docker build -t muse-ai-backend:latest -f Dockerfile .

# Restart deployment to use new image
kubectl rollout restart deployment/backend -n muse-ai

# Watch rollout status
kubectl rollout status deployment/backend -n muse-ai
```

### Viewing Logs

```bash
# Tail logs from a specific pod
kubectl logs -f -n muse-ai deployment/backend

# Get logs from all pods with a label
kubectl logs -f -n muse-ai -l app=backend

# Get logs from previous pod instance (if pod crashed)
kubectl logs -n muse-ai deployment/backend --previous
```

## Troubleshooting

### Common Issues

1. **Pods stuck in Pending state**
   ```bash
   # Check node resources
   kubectl describe nodes
   
   # Check PVC status
   kubectl get pvc -n muse-ai
   
   # Check pod events
   kubectl describe pod -n muse-ai <pod-name>
   ```

2. **ImagePullBackOff errors**
   ```bash
   # Ensure imagePullPolicy is set to Never for local images
   # Make sure docker-env is set and rebuild images
   eval $(minikube docker-env)
   docker build -t muse-ai-backend:latest -f Dockerfile .
   ```

3. **Service connectivity issues**
   ```bash
   # Test service resolution
   kubectl exec -it -n muse-ai deployment/backend -- nslookup postgres-service
   
   # Check service endpoints
   kubectl get endpoints -n muse-ai
   ```

4. **Database connection issues**
   ```bash
   # Check if postgres is ready
   kubectl exec -it -n muse-ai statefulset/postgres -- pg_isready -U muse_ai
   
   # Connect to database
   kubectl exec -it -n muse-ai statefulset/postgres -- psql -U muse_ai -d muse_ai
   ```

### Cleanup

```bash
# Delete the entire application
kubectl delete -f muse-ai-local.yaml

# Delete namespace (removes everything)
kubectl delete namespace muse-ai

# Stop minikube cluster
minikube stop

# Delete minikube cluster (if you want to remove it completely)
minikube delete
```

## Next Steps

- For production deployment, see [`gcp-manual-deployment.md`](./gcp-manual-deployment.md)
- For automated CI/CD, see [`gcp-automated-deployment.md`](./gcp-automated-deployment.md)
- For Dockerfile improvements, see [`dockerfile-improvements.md`](./dockerfile-improvements.md)
