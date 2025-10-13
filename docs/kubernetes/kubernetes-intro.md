# Kubernetes Introduction: Core Concepts and Commands

This document provides a practical introduction to Kubernetes using the Muse AI application deployment as real-world examples. You'll learn the core concepts, terminology, and essential kubectl commands with actual output from our running cluster.

## Table of Contents
- [What is Kubernetes?](#what-is-kubernetes)
- [Core Kubernetes Entities](#core-kubernetes-entities)
- [Essential kubectl Commands](#essential-kubectl-commands)
- [Practical Examples from Muse AI](#practical-examples-from-muse-ai)
- [Common Operations](#common-operations)
- [Troubleshooting](#troubleshooting)

## What is Kubernetes?

Kubernetes (often abbreviated as K8s) is an open-source container orchestration platform that automates the deployment, scaling, and management of containerized applications. Think of it as a smart system that:

- **Manages containers** across multiple machines
- **Ensures applications stay running** by restarting failed containers
- **Handles networking** between different parts of your application
- **Manages storage** and configuration
- **Scales applications** up or down based on demand

## Core Kubernetes Entities

### 1. **Cluster**
The foundation - a set of machines (nodes) that run your containerized applications.

### 2. **Namespace**
A way to organize resources within a cluster, like folders on a computer.

### 3. **Pod**
The smallest deployable unit - contains one or more containers that share storage and network.

### 4. **Deployment**
Manages a set of identical Pods, ensuring the desired number are always running.

### 5. **Service**
Provides stable networking to access Pods (which can come and go).

### 6. **ConfigMap & Secret**
Store configuration data and sensitive information respectively.

### 7. **PersistentVolume (PV) & PersistentVolumeClaim (PVC)**
Handle persistent storage that survives Pod restarts.

### 8. **StatefulSet**
Like Deployment but for stateful applications that need persistent storage and stable network identities.

---

## Essential kubectl Commands

`kubectl` is the command-line tool for interacting with Kubernetes clusters. Let's explore the essential commands using our actual Muse AI deployment.

### Cluster Information

**Check cluster status:**

```bash
kubectl cluster-info
```
```
Kubernetes control plane is running at https://127.0.0.1:32771
CoreDNS is running at https://127.0.0.1:32771/api/v1/namespaces/kube-system/services/kube-dns:dns/proxy

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
```

**View cluster nodes:**

```bash
kubectl get nodes
```
```
NAME       STATUS   ROLES           AGE   VERSION
minikube   Ready    control-plane   68m   v1.33.1
```

---

### Working with Namespaces

**List all namespaces:**

```bash
kubectl get namespaces
```
```
NAME              STATUS   AGE
default           Active   68m
ingress-nginx     Active   59m
kube-node-lease   Active   68m
kube-public       Active   68m
kube-system       Active   68m
muse-ai           Active   43m
```

**Detailed explanation of all namespaces:**

### `default`
- **Purpose**: Where resources go if no namespace is specified
- **Current contents**: Empty (no pods running)
- **Usage**: Default namespace for user workloads, but best practice is to create dedicated namespaces

### `kube-system` 
- **Purpose**: Kubernetes system components and infrastructure pods
- **Current contents**: Core Kubernetes services
  ```bash
  kubectl get pods -n kube-system
  ```
  ```
  NAME                               READY   STATUS    RESTARTS      AGE
  coredns-674b8bbfcf-74kgd           1/1     Running   0             75m
  etcd-minikube                      1/1     Running   0             75m
  kube-apiserver-minikube            1/1     Running   0             75m
  kube-controller-manager-minikube   1/1     Running   0             75m
  kube-proxy-t4ntw                   1/1     Running   0             75m
  kube-scheduler-minikube            1/1     Running   0             75m
  metrics-server-7fbb699795-4bms2    1/1     Running   0             54m
  storage-provisioner                1/1     Running   1 (74m ago)   75m
  ```

- **Key components explained**:
  - `coredns` - DNS service for the cluster (enables service discovery)
  - `etcd` - Key-value store that holds all cluster data
  - `kube-apiserver` - API server that handles all cluster operations
  - `kube-controller-manager` - Manages controllers (deployments, services, etc.)
  - `kube-proxy` - Network proxy for service load balancing
  - `kube-scheduler` - Decides which nodes pods should run on
  - `metrics-server` - Collects resource usage metrics (we enabled this)
  - `storage-provisioner` - Creates persistent volumes (minikube-specific)

### `ingress-nginx`
- **Purpose**: NGINX Ingress Controller for external traffic routing
- **How it got here**: Created when we ran `minikube addons enable ingress`
- **Current contents**: NGINX controller and setup jobs
  ```bash
  kubectl get pods -n ingress-nginx
  ```
  ```
  NAME                                       READY   STATUS      RESTARTS   AGE
  ingress-nginx-admission-create-c5lpx       0/1     Completed   0          65m
  ingress-nginx-admission-patch-qx2xm        0/1     Completed   0          65m
  ingress-nginx-controller-67c5cb88f-lwpb5   1/1     Running     0          65m
  ```

- **Components explained**:
  - `ingress-nginx-controller` - Main NGINX reverse proxy for routing external traffic
  - `admission-create/patch` - One-time setup jobs (Completed status)
- **Purpose in our setup**: Provides ingress capabilities (though we're using NodePort for our nginx service)

### `kube-public`
- **Purpose**: Publicly accessible namespace (readable by all users, including unauthenticated)
- **Current contents**: Empty (no pods running)
- **Typical usage**: Cluster information that should be visible to all users

### `kube-node-lease`
- **Purpose**: Node heartbeat system for cluster health monitoring
- **Current contents**: Lease objects (not pods) for node health tracking
- **Usage**: Internal Kubernetes mechanism to detect when nodes become unhealthy

### `muse-ai`
- **Purpose**: Our application namespace (we created this)
- **Current contents**: All our Muse AI application components
- **Why we created it**: 
  - **Isolation**: Separates our app from system components
  - **Organization**: Groups related resources together
  - **Security**: Enables namespace-specific permissions and network policies
  - **Resource management**: Allows namespace-level quotas and limits

---

### Viewing Pods

**List pods in a specific namespace:**

```bash
kubectl get pods -n muse-ai
```
```
NAME                        READY   STATUS    RESTARTS      AGE
backend-c77cc85dc-9vc25     1/1     Running   2 (43m ago)   44m
frontend-56979f98c7-wqpn4   1/1     Running   0             44m
lilypond-69765dc988-wrcq8   1/1     Running   0             44m
minio-0                     1/1     Running   0             44m
nginx-8fb669c57-qlzcp       1/1     Running   0             41m
postgres-0                  1/1     Running   0             44m
```

**Understanding the output:**
- `NAME` - Unique pod identifier
- `READY` - Containers ready vs total containers in pod (1/1 means 1 out of 1 is ready)
- `STATUS` - Current state (Running, Pending, Error, etc.)
- `RESTARTS` - How many times the pod has been restarted
- `AGE` - How long the pod has been running

**Get detailed information about a specific pod:**

```bash
kubectl describe pod backend-c77cc85dc-9vc25 -n muse-ai
```
```
Name:             backend-c77cc85dc-9vc25
Namespace:        muse-ai
Priority:         0
Service Account:  default
Node:             minikube/192.168.49.2
Start Time:       Sun, 31 Aug 2025 13:54:10 +0100
Labels:           app=backend
                  pod-template-hash=c77cc85dc
Annotations:      <none>
Status:           Running
IP:               10.244.0.10
IPs:
  IP:           10.244.0.10
Controlled By:  ReplicaSet/backend-c77cc85dc
Containers:
  backend:
    Container ID:   docker://faac96e71d3697b3c287b3119f0e1e3d52fb6ecec47f023c06d076686e8542ad
    Image:          muse-ai-backend:latest
    Image ID:       docker://sha256:f1501b44656712dcbdca4cef57865fbf09d49a8ae7e6cd0d59512b6ab912c954
    Port:           8081/TCP
    ... (output truncated)
```

**View pod logs:**

```bash
kubectl logs backend-c77cc85dc-9vc25 -n muse-ai --tail=5
```
```
2025/08/31 12:54:42 Default user ID set to: 894cc426-10d2-4737-aa75-00b0372f9c89
Muse AI Backend API server starting on port 8081
Database: Connected to PostgreSQL
Storage: Connected to MinIO
LilyPond Service: http://muse-ai-lilypond:8082
```

---

### Working with Services

**List services in a namespace:**

```bash
kubectl get services -n muse-ai
```
```
NAME               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
backend-service    ClusterIP   10.110.3.51     <none>        8081/TCP            44m
frontend-service   ClusterIP   10.107.76.214   <none>        3000/TCP            44m
lilypond-service   ClusterIP   10.103.26.137   <none>        8082/TCP            44m
minio-service      ClusterIP   10.104.25.67    <none>        9000/TCP,9001/TCP   44m
nginx-service      NodePort    10.111.188.26   <none>        80:30080/TCP        44m
postgres-service   ClusterIP   10.100.39.39    <none>        5432/TCP            44m
```

**Understanding service types:**
- `ClusterIP` - Only accessible within the cluster (internal services)
- `NodePort` - Accessible from outside the cluster via node IP:port (like nginx-service:30080)
- `LoadBalancer` - Creates an external load balancer (cloud environments)

---

### Working with Deployments

**List deployments:**

```bash
kubectl get deployments -n muse-ai
```
```
NAME       READY   UP-TO-DATE   AVAILABLE   AGE
backend    1/1     1            1           44m
frontend   1/1     1            1           44m
lilypond   1/1     1            1           44m
nginx      1/1     1            1           44m
```

**Understanding deployment status:**
- `READY` - Ready replicas vs desired replicas
- `UP-TO-DATE` - Replicas at the latest version
- `AVAILABLE` - Replicas available to serve traffic

---

### Working with StatefulSets

**List StatefulSets (for stateful applications):**

```bash
kubectl get statefulsets -n muse-ai
```
```
NAME       READY   AGE
minio      1/1     44m
postgres   1/1     44m
```

StatefulSets are used for:
- **PostgreSQL** - Database needs persistent storage and stable network identity
- **MinIO** - Object storage needs persistent volumes

---

### Storage: Persistent Volume Claims

**List persistent volume claims:**

```bash
kubectl get pvc -n muse-ai
```
```
NAME           STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
minio-pvc      Bound    pvc-f7e838ab-39ee-402b-bf7f-ea5527b510b1   10Gi       RWO            standard       <unset>                 44m
postgres-pvc   Bound    pvc-60718ce6-b59d-4554-854e-407fa4da11ea   5Gi        RWO            standard       <unset>                 44m
```

**Understanding PVC status:**
- `STATUS: Bound` - Storage is allocated and attached
- `CAPACITY` - Amount of storage allocated
- `ACCESS MODES: RWO` - ReadWriteOnce (can be mounted by one node at a time)

---

## Practical Examples from Muse AI

### Application Architecture in Kubernetes

Our Muse AI application demonstrates a typical microservices architecture in Kubernetes:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   User Browser  │────│  nginx-service   │────│  nginx Pod      │
└─────────────────┘    │  (NodePort)      │    │  (Reverse Proxy)│
                       └──────────────────┘    └─────────────────┘
                                │                        │
                                └────────────────────────┼────────────────┐
                                                         │                │
                                                         ▼                ▼
                                              ┌─────────────────┐ ┌─────────────────┐
                                              │ frontend-service│ │ backend-service │
                                              │   (ClusterIP)   │ │   (ClusterIP)   │
                                              └─────────────────┘ └─────────────────┘
                                                         │                │
                                                         ▼                ▼
                                              ┌─────────────────┐ ┌─────────────────┐
                                              │  frontend Pod   │ │   backend Pod   │
                                              │   (SvelteKit)   │ │   (Go Server)   │
                                              └─────────────────┘ └─────────────────┘
                                                                          │
                                                    ┌─────────────────────┼─────────────────────┐
                                                    │                     │                     │
                                                    ▼                     ▼                     ▼
                                         ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
                                         │postgres-service │ │  minio-service  │ │lilypond-service │
                                         │   (ClusterIP)   │ │   (ClusterIP)   │ │   (ClusterIP)   │
                                         └─────────────────┘ └─────────────────┘ └─────────────────┘
                                                    │                     │                     │
                                                    ▼                     ▼                     ▼
                                         ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
                                         │   postgres-0    │ │     minio-0     │ │  lilypond Pod   │
                                         │ (StatefulSet)   │ │  (StatefulSet)  │ │ (Deployment)    │
                                         │     + PVC       │ │      + PVC      │ └─────────────────┘
                                         └─────────────────┘ └─────────────────┘
```

### Configuration Management

**View ConfigMaps:**

```bash
kubectl get configmaps -n muse-ai
```
```
NAME               DATA   AGE
kube-root-ca.crt   1      46m
muse-ai-config     7      46m
```

**View ConfigMap details:**

```bash
kubectl describe configmap muse-ai-config -n muse-ai
```
```
Name:         muse-ai-config
Namespace:    muse-ai
Labels:       <none>
Annotations:  <none>

Data
====
PORT:
----
8082

POSTGRES_DB:
----
muse_ai

POSTGRES_USER:
----
muse_ai

DATABASE_URL:
----
postgres://muse_ai:muse_ai_pass@postgres-service:5432/muse_ai?sslmode=disable

GO_ENV:
----
development

MINIO_BUCKET:
----
muse-ai-files

MINIO_ENDPOINT:
----
minio-service:9000
```

**View Secrets:**

```bash
kubectl get secrets -n muse-ai
```
```
NAME              TYPE     DATA   AGE
muse-ai-secrets   Opaque   4      46m
```

*Note: Secret values are base64 encoded and not shown for security reasons.*

---

## Common Operations

### Scaling Applications

**Scale a deployment:**

```bash
# Scale backend to 3 replicas
kubectl scale deployment backend --replicas=3 -n muse-ai

# Check the scaling result
kubectl get pods -n muse-ai | grep backend
```

### Port Forwarding for Local Access

**Forward a pod port to your local machine:**

```bash
# Forward nginx service port 80 to local port 8080
kubectl port-forward -n muse-ai service/nginx-service 8080:80

# Access at http://localhost:8080
```

### Executing Commands in Pods

**Connect to a running pod:**

```bash
# Open a shell in the postgres pod
kubectl exec -it postgres-0 -n muse-ai -- /bin/sh

# Run a single command
kubectl exec -it postgres-0 -n muse-ai -- psql -U muse_ai -d muse_ai -c "SELECT version();"
```

### Viewing Resource Usage

**Get resource usage:**

```bash
# View pod resource usage (requires metrics-server)
kubectl top pods -n muse-ai

# View node resource usage
kubectl top nodes
```

### Rolling Updates

**Update a deployment:**

```bash
# Update the backend image
kubectl set image deployment/backend backend=muse-ai-backend:v2.0 -n muse-ai

# Check rollout status
kubectl rollout status deployment/backend -n muse-ai

# View rollout history
kubectl rollout history deployment/backend -n muse-ai

# Rollback if needed
kubectl rollout undo deployment/backend -n muse-ai
```

---

## Troubleshooting

### Common kubectl Commands for Debugging

**Check pod status and events:**

```bash
# Get detailed pod information
kubectl describe pod <pod-name> -n muse-ai

# Check events in the namespace
kubectl get events -n muse-ai --sort-by='.lastTimestamp'

# Follow logs in real-time
kubectl logs -f deployment/backend -n muse-ai

# Get logs from previous pod instance (if crashed)
kubectl logs deployment/backend -n muse-ai --previous
```

**Check service connectivity:**

```bash
# Test service resolution from within a pod
kubectl exec -it backend-c77cc85dc-9vc25 -n muse-ai -- nslookup postgres-service

# Check service endpoints
kubectl get endpoints -n muse-ai

# Test port connectivity
kubectl exec -it backend-c77cc85dc-9vc25 -n muse-ai -- nc -zv postgres-service 5432
```

**Resource issues:**

```bash
# Check resource quotas
kubectl describe namespace muse-ai

# Check node resources
kubectl describe nodes

# Check if PVCs are bound
kubectl get pv,pvc -n muse-ai
```

### Common Pod States and Their Meanings

- **Running** - Pod is executing normally
- **Pending** - Pod is waiting to be scheduled (usually resource constraints)
- **CrashLoopBackOff** - Pod keeps crashing and restarting
- **ImagePullBackOff** - Cannot pull the container image
- **Error** - Pod has terminated with an error
- **Completed** - Pod has successfully completed (for jobs)

---

## Key Takeaways

### Why Kubernetes for Muse AI?

1. **Microservices Architecture** - Each component (frontend, backend, database, etc.) runs independently
2. **Scalability** - Can easily scale individual components based on demand
3. **High Availability** - Automatic restart of failed components
4. **Service Discovery** - Components find each other using DNS names (like `postgres-service`)
5. **Configuration Management** - Environment-specific settings managed centrally
6. **Storage Management** - Persistent data survives pod restarts

### Development Workflow

1. **Build** Docker images locally
2. **Deploy** using kubectl apply
3. **Debug** using logs and exec commands
4. **Update** by pushing new images and updating deployments
5. **Scale** by adjusting replica counts

### Best Practices Demonstrated

- **Namespacing** - Isolate application resources
- **Labels** - Organize and select resources
- **Health Checks** - Readiness and liveness probes
- **Resource Limits** - Prevent resource starvation
- **Secrets Management** - Separate sensitive data
- **Persistent Storage** - Data survives container restarts

This Kubernetes setup provides a robust, scalable foundation for the Muse AI application that can grow from development to production seamlessly.
