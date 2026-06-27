# Slash: Docker Setup & Containerization

**Docker images, compose files, and container orchestration.**

---

## Table of Contents

1. [Dockerfile](#dockerfile)
2. [Docker Compose](#docker-compose)
3. [Local Development](#local-development)
4. [Production Deployment](#production-deployment)
5. [Image Optimization](#image-optimization)

---

## Dockerfile

### Production Dockerfile (Slim)

**File: `Dockerfile`**

```dockerfile
# Multi-stage build for minimal production image

# Stage 1: Builder
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make ca-certificates

WORKDIR /app

# Copy source
COPY . .

# Build
RUN make build-static

# Verify binary
RUN ./slash version

---

# Stage 2: Runtime (minimal)
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    sqlite \
    tini

# Create non-root user
RUN addgroup -g 1000 slash && \
    adduser -D -u 1000 -G slash slash

# Create directories
RUN mkdir -p /app /etc/slash /var/cache/slash && \
    chown -R slash:slash /app /etc/slash /var/cache/slash

# Copy binary from builder
COPY --from=builder /app/slash /usr/local/bin/slash

# Set permissions
RUN chmod +x /usr/local/bin/slash

# Switch to non-root user
USER slash

WORKDIR /app

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD slash version || exit 1

# Use tini as init process (handle signals properly)
ENTRYPOINT ["/sbin/tini", "--"]

# Default command
CMD ["slash", "daemon"]
```

**Build:**
```bash
docker build -t slash:latest .
docker build -t slash:1.0.0 -t slash:latest .
```

**Test:**
```bash
docker run --rm slash:latest slash version
→ Slash v1.0.0
```

---

### Development Dockerfile

**File: `Dockerfile.dev`**

```dockerfile
# Development image with hot reload and debugging

FROM golang:1.21-alpine

# Install dev dependencies
RUN apk add --no-cache \
    git \
    make \
    bash \
    ca-certificates \
    sqlite \
    delve \
    curl

# Install Air for hot reload
RUN go install github.com/cosmtrek/air@latest

WORKDIR /app

# Copy go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Create volumes
RUN mkdir -p /app /root/.slash /root/.cache/slash

# Expose debug port
EXPOSE 2345

# Expose MCP port
EXPOSE 8765

# Install on start
CMD ["air"]
```

**Usage:**
```bash
docker build -f Dockerfile.dev -t slash:dev .
docker run -it -p 2345:2345 -p 8765:8765 \
  -v $(pwd):/app \
  -v slash-cache:/root/.cache/slash \
  slash:dev
```

---

### Testing Dockerfile

**File: `Dockerfile.test`**

```dockerfile
# Test runner image

FROM golang:1.21-alpine

RUN apk add --no-cache git make bash ca-certificates

WORKDIR /app

COPY . .

RUN go mod download

# Run tests
CMD ["make", "test"]
```

---

## Docker Compose

### Local Development Setup

**File: `docker-compose.yml`**

```yaml
version: "3.8"

services:
  # Slash daemon
  slash:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: slash-daemon
    command: daemon
    
    ports:
      - "8765:8765"  # MCP server
    
    volumes:
      - slash-cache:/var/cache/slash
      - slash-config:/etc/slash
      - ./data:/app/data
    
    environment:
      LOG_LEVEL: "debug"
      SLASH_CONFIG: /etc/slash/config.json
    
    healthcheck:
      test: ["CMD", "slash", "stats"]
      interval: 10s
      timeout: 5s
      retries: 3
    
    restart: unless-stopped

  # Optional: Prometheus for monitoring
  prometheus:
    image: prom/prometheus:latest
    container_name: slash-prometheus
    
    ports:
      - "9090:9090"
    
    volumes:
      - ./monitoring/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus-data:/prometheus
    
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
    
    depends_on:
      - slash

  # Optional: Grafana for dashboards
  grafana:
    image: grafana/grafana:latest
    container_name: slash-grafana
    
    ports:
      - "3000:3000"
    
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
      GF_INSTALL_PLUGINS: grafana-piechart-panel
    
    volumes:
      - grafana-data:/var/lib/grafana
      - ./monitoring/grafana-dashboards:/etc/grafana/provisioning/dashboards
    
    depends_on:
      - prometheus

volumes:
  slash-cache:
  slash-config:
  prometheus-data:
  grafana-data:
```

**Usage:**
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f slash

# Stop all services
docker-compose down

# View container status
docker-compose ps
```

---

### Multi-Service Setup (with mocks)

**File: `docker-compose.services.yml`**

```yaml
version: "3.8"

services:
  # Slash daemon
  slash:
    build: .
    container_name: slash-daemon
    command: daemon
    ports:
      - "8765:8765"
    volumes:
      - slash-cache:/var/cache/slash
    depends_on:
      - postgres
    healthcheck:
      test: ["CMD", "slash", "stats"]
      interval: 10s
      timeout: 5s
      retries: 3

  # PostgreSQL for testing
  postgres:
    image: postgres:15-alpine
    container_name: slash-postgres
    
    environment:
      POSTGRES_DB: slash_test
      POSTGRES_USER: slash
      POSTGRES_PASSWORD: dev
    
    ports:
      - "5432:5432"
    
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql
    
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U slash"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis for caching
  redis:
    image: redis:7-alpine
    container_name: slash-redis
    
    ports:
      - "6379:6379"
    
    volumes:
      - redis-data:/data
    
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 3

  # Mock API server
  mockapi:
    image: mockserver/mockserver:latest
    container_name: slash-mockapi
    
    ports:
      - "1080:1080"
    
    environment:
      MOCKSERVER_INITIALIZATION_JSON_PATH: /config/init.json
    
    volumes:
      - ./tests/mock-api-config.json:/config/init.json

volumes:
  slash-cache:
  postgres-data:
  redis-data:
```

---

### Testing Setup

**File: `docker-compose.test.yml`**

```yaml
version: "3.8"

services:
  # Test runner
  test:
    build:
      context: .
      dockerfile: Dockerfile.test
    container_name: slash-test
    
    volumes:
      - ./coverage:/app/coverage
    
    depends_on:
      slash-daemon:
        condition: service_healthy

  # Slash daemon for tests
  slash-daemon:
    build: .
    container_name: slash-daemon-test
    command: daemon
    
    volumes:
      - slash-cache-test:/var/cache/slash
    
    healthcheck:
      test: ["CMD", "slash", "version"]
      interval: 5s
      timeout: 3s
      retries: 5

  # PostgreSQL for tests
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: test
      POSTGRES_USER: test
      POSTGRES_PASSWORD: test
    
    volumes:
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql

volumes:
  slash-cache-test:
```

---

## Local Development

### Quick Start with Docker

```bash
# Clone repo
git clone https://github.com/yoosuf/Slash.git
cd slash

# Start Slash
docker-compose up -d

# Check it's running
docker-compose ps

# View logs
docker-compose logs -f slash

# Test it works
docker exec slash-daemon slash version

# Install plugin
docker exec slash-daemon slash plugin install claude-code

# View stats
docker exec slash-daemon slash stats
```

---

### Development with Hot Reload

```bash
# Build dev image
docker build -f Dockerfile.dev -t slash:dev .

# Start with volume mount for code changes
docker run -it \
  -p 2345:2345 \
  -p 8765:8765 \
  -v $(pwd):/app \
  slash:dev

# Changes to source code auto-reload
# Modify cmd/slash/main.go → Auto recompiles

# Debug with Delve
# In VS Code: attach to localhost:2345
```

---

### Database Development

```bash
# Start with all services
docker-compose -f docker-compose.services.yml up -d

# Connect to PostgreSQL
docker exec -it slash-postgres psql -U slash -d slash_test

# Run migrations
docker exec slash-daemon make migrate

# View database
docker exec -it slash-postgres psql -U slash -d slash_test -c "\dt"
```

---

## Production Deployment

### Kubernetes Deployment

**File: `k8s/slash-deployment.yaml`**

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: slash

---
apiVersion: v1
kind: ConfigMap
metadata:
  name: slash-config
  namespace: slash
data:
  config.json: |
    {
      "daemon": {
        "port": 8765,
        "log_level": "warn",
        "auto_start": true
      },
      "compression": {
        "enabled": true,
        "diff_only_reads": true,
        "output_compress": true
      },
      "cache": {
        "ttl_hours": 24,
        "max_size_mb": 2048
      }
    }

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: slash
  namespace: slash
spec:
  replicas: 3
  selector:
    matchLabels:
      app: slash
  template:
    metadata:
      labels:
        app: slash
    spec:
      containers:
      - name: slash
        image: slash:1.0.0
        imagePullPolicy: IfNotPresent
        
        command: ["slash", "daemon"]
        
        ports:
        - name: mcp
          containerPort: 8765
        
        resources:
          requests:
            cpu: 250m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi
        
        volumeMounts:
        - name: cache
          mountPath: /var/cache/slash
        - name: config
          mountPath: /etc/slash
        
        livenessProbe:
          exec:
            command: ["slash", "stats"]
          initialDelaySeconds: 5
          periodSeconds: 10
        
        readinessProbe:
          exec:
            command: ["slash", "version"]
          initialDelaySeconds: 3
          periodSeconds: 5
      
      volumes:
      - name: cache
        emptyDir:
          sizeLimit: 2Gi
      - name: config
        configMap:
          name: slash-config

---
apiVersion: v1
kind: Service
metadata:
  name: slash
  namespace: slash
spec:
  type: ClusterIP
  selector:
    app: slash
  ports:
  - name: mcp
    port: 8765
    targetPort: 8765
```

**Deploy:**
```bash
kubectl apply -f k8s/slash-deployment.yaml

# Verify
kubectl get pods -n slash
kubectl logs -n slash -l app=slash

# Scale
kubectl scale deployment slash -n slash --replicas=5
```

---

## Image Optimization

### Size Comparison

```
Dockerfile        Size    Notes
─────────────────────────────────────
Dockerfile        15MB    Multi-stage, optimized
Dockerfile.dev    500MB   Full Go toolchain
Dockerfile.test   400MB   Full Go toolchain

Production best practice: Use multi-stage build
Result: 15MB for production vs 500MB+ for dev
```

---

### Push to Registry

```bash
# Tag for registry
docker tag slash:latest myregistry.azurecr.io/slash:latest
docker tag slash:latest myregistry.azurecr.io/slash:1.0.0

# Login to registry
az acr login --name myregistry

# Push
docker push myregistry.azurecr.io/slash:latest
docker push myregistry.azurecr.io/slash:1.0.0

# Verify
az acr repository list --name myregistry

# Pull from K8s
kubectl set image deployment/slash \
  slash=myregistry.azurecr.io/slash:1.0.0 \
  -n slash
```

---

### Docker Security Best Practices

```dockerfile
# ✓ Use specific base image version (not latest)
FROM alpine:3.18

# ✓ Use non-root user
USER slash

# ✓ Set read-only filesystem
security:
  readOnlyRootFilesystem: true

# ✓ Don't run privileged
securityContext:
  privileged: false

# ✓ Scan for vulnerabilities
docker scan slash:latest

# ✓ Sign images
docker trust sign slash:latest
```

---

## Monitoring & Logs

### View Logs

```bash
# All services
docker-compose logs

# Specific service
docker-compose logs slash

# Follow logs
docker-compose logs -f slash

# Last 100 lines
docker-compose logs --tail=100 slash
```

---

### Health Checks

```bash
# Check container health
docker inspect slash-daemon | grep -A 5 Health

# Output:
# "Status": "healthy"
# "FailingStreak": 0
# "LastCheck": "2024-06-27T14:22:15.123456789Z"

# Manual health check
curl http://localhost:8765/health
→ {"status": "healthy", "service": "slash-mcp"}
```

---

### Performance Monitoring

```bash
# Container stats
docker stats slash-daemon

# Memory usage
docker exec slash-daemon slash stats | grep cache_size

# CPU/Memory limits
docker stats --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}"
```

---

## Common Tasks

### Rebuild Image

```bash
# Remove old image
docker rmi slash:latest

# Build new image
docker build -t slash:latest .

# Restart containers
docker-compose down
docker-compose up -d
```

---

### Clean Up

```bash
# Remove unused images
docker image prune

# Remove unused volumes
docker volume prune

# Full cleanup (warning: removes containers & networks!)
docker system prune -a
```

---

### Backup & Restore

```bash
# Backup cache volume
docker run --rm \
  -v slash-cache:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/slash-cache.tar.gz -C /data .

# Restore cache volume
docker run --rm \
  -v slash-cache:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/slash-cache.tar.gz -C /data
```

---

## Summary: Docker Quick Reference

| Task | Command |
|------|---------|
| **Build** | `docker build -t slash:latest .` |
| **Run** | `docker run slash:latest slash daemon` |
| **Compose up** | `docker-compose up -d` |
| **View logs** | `docker-compose logs -f slash` |
| **Test** | `docker-compose -f docker-compose.test.yml up` |
| **Deploy to K8s** | `kubectl apply -f k8s/slash-deployment.yaml` |
| **Push to registry** | `docker push myregistry/slash:latest` |
| **Health check** | `docker-compose ps` |
| **Cleanup** | `docker-compose down && docker volume prune` |

---

**Docker setup is production-ready. Choose based on your needs:**
- **Development:** Use `docker-compose.yml` (15 seconds to start)
- **Testing:** Use `docker-compose.test.yml` (isolated, clean)
- **Production:** Use Kubernetes (scalable, resilient)
