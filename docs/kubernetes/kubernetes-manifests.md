# Kubernetes Manifests

This document contains all Kubernetes manifests required to deploy the Muse AI application.

## Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: muse-ai
  labels:
    name: muse-ai
```

## ConfigMaps

### Application Configuration
```yaml
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
  GO_ENV: "production"
  PORT: "8082"
```

## Secrets

### Database and Storage Secrets
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: muse-ai-secrets
  namespace: muse-ai
type: Opaque
data:
  # Base64 encoded values
  POSTGRES_PASSWORD: bXVzZV9haV9wYXNz  # muse_ai_pass
  MINIO_ACCESS_KEY: bXVzZV9haQ==        # muse_ai
  MINIO_SECRET_KEY: bXVzZV9haV9wYXNz    # muse_ai_pass
  ANTHROPIC_API_KEY: ""                 # Set your actual API key
```

## Storage

### PostgreSQL PersistentVolume
```yaml
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
      storage: 10Gi
  storageClassName: standard
```

### MinIO PersistentVolume
```yaml
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
      storage: 50Gi
  storageClassName: standard
```

## StatefulSets

### PostgreSQL Database
```yaml
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
        livenessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - muse_ai
            - -d
            - muse_ai
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
        readinessProbe:
          exec:
            command:
            - pg_isready
            - -U
            - muse_ai
            - -d
            - muse_ai
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        securityContext:
          runAsNonRoot: true
          runAsUser: 999
          runAsGroup: 999
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
```

### MinIO Object Storage
```yaml
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
        livenessProbe:
          httpGet:
            path: /minio/health/live
            port: 9000
          initialDelaySeconds: 30
          periodSeconds: 20
          timeoutSeconds: 10
        readinessProbe:
          httpGet:
            path: /minio/health/ready
            port: 9000
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        securityContext:
          runAsNonRoot: true
          runAsUser: 1000
          runAsGroup: 1000
      volumes:
      - name: minio-storage
        persistentVolumeClaim:
          claimName: minio-pvc
```

## Deployments

### Backend Service
```yaml
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
      containers:
      - name: backend
        image: muse-ai-backend:latest
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
        livenessProbe:
          httpGet:
            path: /api/health
            port: 8081
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
        readinessProbe:
          httpGet:
            path: /api/health
            port: 8081
          initialDelaySeconds: 5
          periodSeconds: 5
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
          readOnlyRootFilesystem: false
          allowPrivilegeEscalation: false
```

### LilyPond Service
```yaml
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
        ports:
        - containerPort: 8082
        env:
        - name: PORT
          valueFrom:
            configMapKeyRef:
              name: muse-ai-config
              key: PORT
        livenessProbe:
          httpGet:
            path: /health
            port: 8082
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8082
          initialDelaySeconds: 5
          periodSeconds: 5
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
          readOnlyRootFilesystem: false
          allowPrivilegeEscalation: false
        volumeMounts:
        - name: temp-storage
          mountPath: /tmp/lilypond
      volumes:
      - name: temp-storage
        emptyDir: {}
```

### Frontend Service
```yaml
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
        image: muse-ai-frontend:latest
        ports:
        - containerPort: 3000
        livenessProbe:
          httpGet:
            path: /
            port: 3000
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
        readinessProbe:
          httpGet:
            path: /
            port: 3000
          initialDelaySeconds: 5
          periodSeconds: 5
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
          readOnlyRootFilesystem: false
          allowPrivilegeEscalation: false
```

### Nginx Reverse Proxy
```yaml
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
        image: muse-ai-nginx:latest
        ports:
        - containerPort: 80
        livenessProbe:
          httpGet:
            path: /health
            port: 80
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
        readinessProbe:
          httpGet:
            path: /health
            port: 80
          initialDelaySeconds: 5
          periodSeconds: 5
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
          readOnlyRootFilesystem: false
          allowPrivilegeEscalation: false
```

## Services

### PostgreSQL Service
```yaml
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
```

### MinIO Service
```yaml
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
```

### Backend Service
```yaml
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
```

### LilyPond Service
```yaml
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
```

### Frontend Service
```yaml
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
```

### Nginx LoadBalancer Service
```yaml
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

## Horizontal Pod Autoscaler

### Backend HPA
```yaml
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
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Frontend HPA
```yaml
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

## Network Policies

### Database Network Policy
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: postgres-netpol
  namespace: muse-ai
spec:
  podSelector:
    matchLabels:
      app: postgres
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: backend
    ports:
    - protocol: TCP
      port: 5432
```

### MinIO Network Policy
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: minio-netpol
  namespace: muse-ai
spec:
  podSelector:
    matchLabels:
      app: minio
  policyTypes:
  - Ingress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: backend
    ports:
    - protocol: TCP
      port: 9000
```

## Pod Disruption Budgets

### Backend PDB
```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: backend-pdb
  namespace: muse-ai
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: backend
```

### Frontend PDB
```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: frontend-pdb
  namespace: muse-ai
spec:
  minAvailable: 1
  selector:
    matchLabels:
      app: frontend
```
