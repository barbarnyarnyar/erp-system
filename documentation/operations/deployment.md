# Production Deployment

Deploy the ERP system to production environments using Kubernetes, Docker Swarm, or cloud platforms.

## Deployment Options

### Kubernetes (Recommended)
Best for enterprise deployments requiring high availability, auto-scaling, and advanced orchestration.

### Docker Swarm  
Good for smaller deployments that need container orchestration without Kubernetes complexity.

### Cloud Platforms
Native cloud services like AWS ECS, Google Cloud Run, or Azure Container Instances.

## Kubernetes Deployment

### Prerequisites
- Kubernetes cluster 1.25+
- kubectl configured for cluster access
- Helm 3.0+ (recommended)
- Ingress controller (NGINX, Traefik, or cloud provider)
- Storage class for persistent volumes

### Step 1: Create Namespace and Secrets
```bash
# Create namespace
kubectl create namespace erp-system

# Create database credentials
kubectl create secret generic db-credentials \
  --from-literal=username=postgres \
  --from-literal=password=production-db-password \
  --namespace=erp-system

# Create Redis authentication
kubectl create secret generic redis-auth \
  --from-literal=password=redis-auth-token \
  --namespace=erp-system

# Create JWT secret
kubectl create secret generic jwt-secret \
  --from-literal=secret=jwt-signing-secret-256-bit \
  --namespace=erp-system
```

### Step 2: Deploy Infrastructure Services
```yaml
# infrastructure.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: erp-system
spec:
  serviceName: postgres
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
        image: postgres:15-alpine
        ports:
        - containerPort: 5432
        env:
        - name: POSTGRES_DB
          value: "erp_system"
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: username
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: password
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1"
  volumeClaimTemplates:
  - metadata:
      name: postgres-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi

---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: erp-system
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
```

### Step 3: Deploy Application Services
```yaml
# services.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: fm-service
  namespace: erp-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: fm-service
  template:
    metadata:
      labels:
        app: fm-service
    spec:
      containers:
      - name: fm-service
        image: erp-system/fm-service:v1.0.0
        ports:
        - containerPort: 8001
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: DB_HOST
          value: "postgres"
        - name: DB_USERNAME
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: username
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: db-credentials
              key: password
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: jwt-secret
              key: secret
        livenessProbe:
          httpGet:
            path: /health
            port: 8001
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8001
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
```

### Step 4: Configure Ingress
```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: erp-ingress
  namespace: erp-system
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/rate-limit: "100"
spec:
  tls:
  - hosts:
    - erp.company.com
    secretName: erp-tls
  rules:
  - host: erp.company.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: api-gateway
            port:
              number: 80
```

### Step 5: Deploy with Kubectl
```bash
# Deploy infrastructure
kubectl apply -f infrastructure.yaml

# Wait for infrastructure to be ready
kubectl wait --for=condition=ready pod -l app=postgres --timeout=300s -n erp-system

# Deploy services
kubectl apply -f services.yaml

# Deploy ingress
kubectl apply -f ingress.yaml

# Verify deployment
kubectl get pods -n erp-system
kubectl get services -n erp-system
kubectl get ingress -n erp-system
```

## Docker Swarm Deployment

### Step 1: Initialize Swarm Cluster
```bash
# On first manager node
docker swarm init --advertise-addr <MANAGER-IP>

# Join additional managers
docker swarm join-token manager
# Run output command on other manager nodes

# Join worker nodes
docker swarm join-token worker
# Run output command on worker nodes

# Verify cluster
docker node ls
```

### Step 2: Create Networks and Secrets
```bash
# Create overlay networks
docker network create --driver overlay --attachable erp-backend
docker network create --driver overlay --attachable erp-database

# Create secrets
echo "production-db-password" | docker secret create db_password -
echo "jwt-signing-secret-256-bit" | docker secret create jwt_secret -
```

### Step 3: Deploy Stack
```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/db_password
      POSTGRES_DB: erp_system
    secrets:
      - db_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - erp-database
    deploy:
      replicas: 1
      placement:
        constraints: [node.role == manager]
      resources:
        limits:
          memory: 2G
          cpus: '1.0'

  fm-service:
    image: erp-system/fm-service:${VERSION:-latest}
    environment:
      - ENVIRONMENT=production
      - DB_HOST=postgres
      - DB_PASSWORD_FILE=/run/secrets/db_password
      - JWT_SECRET_FILE=/run/secrets/jwt_secret
    secrets:
      - db_password
      - jwt_secret
    networks:
      - erp-backend
      - erp-database
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 30s
        failure_action: rollback
      resources:
        limits:
          memory: 512M
          cpus: '0.5'

  api-gateway:
    image: erp-system/api-gateway:${VERSION:-latest}
    environment:
      - ENVIRONMENT=production
      - FM_SERVICE_URL=http://fm-service:8001
      - JWT_SECRET_FILE=/run/secrets/jwt_secret
    secrets:
      - jwt_secret
    ports:
      - "80:8080"
      - "443:8443"
    networks:
      - erp-backend
    deploy:
      replicas: 2

volumes:
  postgres_data:

networks:
  erp-backend:
    external: true
  erp-database:
    external: true

secrets:
  db_password:
    external: true
  jwt_secret:
    external: true
```

### Step 4: Deploy Stack
```bash
# Deploy the stack
export VERSION=v1.0.0
docker stack deploy -c docker-compose.prod.yml erp

# Verify deployment
docker stack ps erp
docker service ls
```

## AWS ECS Deployment

### Prerequisites
- AWS CLI configured
- ECS Cluster created
- Application Load Balancer
- RDS PostgreSQL instance
- ElastiCache Redis cluster

### Step 1: Create Task Definition
```json
{
  "family": "fm-service",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "fm-service",
      "image": "ACCOUNT.dkr.ecr.REGION.amazonaws.com/erp-system/fm-service:latest",
      "portMappings": [
        {
          "containerPort": 8001,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "ENVIRONMENT",
          "value": "production"
        },
        {
          "name": "DB_HOST", 
          "value": "erp-postgres.cluster-xyz.us-west-2.rds.amazonaws.com"
        }
      ],
      "secrets": [
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:REGION:ACCOUNT:secret:erp/db/password"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/fm-service",
          "awslogs-region": "us-west-2",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

### Step 2: Create ECS Service
```bash
# Create service for Financial Management
aws ecs create-service \
  --cluster erp-production \
  --service-name fm-service \
  --task-definition fm-service:1 \
  --desired-count 2 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-12345,subnet-67890],securityGroups=[sg-abcdef],assignPublicIp=DISABLED}" \
  --load-balancers "targetGroupArn=arn:aws:elasticloadbalancing:us-west-2:ACCOUNT:targetgroup/fm-service/123456,containerName=fm-service,containerPort=8001"
```

## Environment Configuration

### Production Environment Variables
```bash
# Application Configuration
ENVIRONMENT=production
LOG_LEVEL=warn

# Database Configuration  
DB_HOST=prod-postgresql.company.com
DB_PORT=5432
DB_USERNAME=erp_user
DB_PASSWORD_FILE=/run/secrets/db_password

# Cache Configuration
REDIS_HOST=prod-redis.company.com
REDIS_AUTH_FILE=/run/secrets/redis_password

# Message Queue Configuration
KAFKA_BROKERS=prod-kafka1.company.com:9092,prod-kafka2.company.com:9092

# Security Configuration
JWT_SECRET_FILE=/run/secrets/jwt_secret
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=300s

# TLS Configuration
TLS_ENABLED=true
TLS_CERT_FILE=/certs/tls.crt
TLS_KEY_FILE=/certs/tls.key
```

## Health Checks and Readiness

### Service Health Check
All services expose health endpoints:
```bash
# Check service health
curl https://erp.company.com/api/v1/finance/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2024-03-15T10:30:00Z",
  "services": {
    "database": {
      "status": "up",
      "response_time_ms": 2
    },
    "redis": {
      "status": "up", 
      "response_time_ms": 1
    }
  },
  "version": "1.0.0"
}
```

### Kubernetes Probes
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8001
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /health
    port: 8001
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 2
```

## Deployment Verification

### Post-Deployment Checklist
- [ ] All services are running and healthy
- [ ] Database connections are working
- [ ] API endpoints are responding
- [ ] Authentication is working
- [ ] SSL/TLS certificates are valid
- [ ] Monitoring and logging are operational
- [ ] Backup procedures are in place

### Verification Commands
```bash
# Check all services
kubectl get pods -n erp-system

# Verify ingress
kubectl get ingress -n erp-system

# Test API endpoints
curl -k https://erp.company.com/api/v1/finance/accounts

# Check logs
kubectl logs -f deployment/fm-service -n erp-system
```

## Rollback Procedures

### Kubernetes Rollback
```bash
# Check rollout status
kubectl rollout status deployment/fm-service -n erp-system

# View rollout history
kubectl rollout history deployment/fm-service -n erp-system

# Rollback to previous version
kubectl rollout undo deployment/fm-service -n erp-system

# Rollback to specific revision
kubectl rollout undo deployment/fm-service --to-revision=2 -n erp-system
```

### Docker Swarm Rollback
```bash
# Rollback service to previous image
docker service rollback erp_fm-service

# Update to specific version
docker service update --image erp-system/fm-service:v1.0.1 erp_fm-service
```

## Next Steps

After successful deployment:
- [Configure Monitoring](monitoring.md) - Set up metrics and alerting
- [Security Configuration](security.md) - Harden your deployment
- [Backup and Recovery](backup-recovery.md) - Protect your data
- [Performance Optimization](performance.md) - Tune for optimal performance