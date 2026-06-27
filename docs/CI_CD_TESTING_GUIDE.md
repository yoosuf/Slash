# Slash: CI/CD Testing Guide

**GitHub Actions, Docker, and automated testing pipelines.**

---

## Table of Contents

1. [GitHub Actions Workflows](#github-actions-workflows)
2. [Docker Setup for CI](#docker-setup-for-ci)
3. [Testing Strategies](#testing-strategies)
4. [Performance Benchmarking](#performance-benchmarking)
5. [Integration Testing](#integration-testing)

---

## GitHub Actions Workflows

### Workflow 1: Build & Test on Push

**File: `.github/workflows/build-test.yml`**

```yaml
name: Build & Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  GO_VERSION: "1.21"
  
jobs:
  build:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: ["1.21"]
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
      with:
        fetch-depth: 0
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ matrix.go-version }}
        cache: true
    
    - name: Install dependencies
      run: |
        go mod download
        go mod verify
    
    - name: Run linting
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
        args: --deadline=5m
    
    - name: Build binary
      run: make build
    
    - name: Run unit tests
      run: make test
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./coverage.out
        flags: unittests
        name: codecov-umbrella
    
    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: slash-${{ matrix.os }}-${{ matrix.go-version }}
        path: ./slash

  test-compression:
    runs-on: ubuntu-latest
    needs: build
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
    
    - name: Build binary
      run: make build
    
    - name: Run compression benchmarks
      run: |
        ./slash bench --output bench-results.json
    
    - name: Upload benchmark results
      uses: actions/upload-artifact@v3
      with:
        name: benchmark-results
        path: bench-results.json
    
    - name: Compare benchmarks
      if: github.event_name == 'pull_request'
      uses: actions/github-script@v6
      with:
        script: |
          const fs = require('fs');
          const results = JSON.parse(fs.readFileSync('bench-results.json', 'utf8'));
          
          let summary = '## Compression Benchmarks\n\n';
          summary += `- Pass Rate: ${results.summary.pass_rate}%\n`;
          summary += `- Avg Reduction: ${results.summary.avg_reduction}%\n`;
          summary += `- Latency (p95): ${results.summary.latency_p95}ms\n`;
          
          github.rest.issues.createComment({
            issue_number: context.issue.number,
            owner: context.repo.owner,
            repo: context.repo.repo,
            body: summary
          });
```

---

### Workflow 2: Release Build

**File: `.github/workflows/release.yml`**

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build-release:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        include:
          - os: ubuntu-latest
            goos: linux
            goarch: amd64
          - os: ubuntu-latest
            goos: linux
            goarch: arm64
          - os: macos-latest
            goos: darwin
            goarch: amd64
          - os: macos-latest
            goos: darwin
            goarch: arm64
          - os: windows-latest
            goos: windows
            goarch: amd64
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: "1.21"
    
    - name: Build binary
      env:
        GOOS: ${{ matrix.goos }}
        GOARCH: ${{ matrix.goarch }}
      run: |
        VERSION=${GITHUB_REF#refs/tags/}
        go build -ldflags="-X main.Version=$VERSION" \
          -o slash-${{ matrix.goos }}-${{ matrix.goarch }} \
          ./cmd/slash
    
    - name: Create tarball (Unix)
      if: matrix.goos != 'windows'
      run: |
        VERSION=${GITHUB_REF#refs/tags/}
        tar -czf slash_${VERSION}_${{ matrix.goos }}_${{ matrix.goarch }}.tar.gz \
          slash-${{ matrix.goos }}-${{ matrix.goarch }}
    
    - name: Create zip (Windows)
      if: matrix.goos == 'windows'
      run: |
        $version = "${{ github.ref }}".Replace('refs/tags/', '')
        Compress-Archive -Path "slash-${{ matrix.goos }}-${{ matrix.goarch }}.exe" \
          -DestinationPath "slash_${version}_${{ matrix.goos }}_${{ matrix.goarch }}.zip"
    
    - name: Generate checksums
      run: |
        sha256sum slash_* > SHA256SUMS
        gpg --detach-sign --armor SHA256SUMS
    
    - name: Create Release
      uses: softprops/action-gh-release@v1
      with:
        files: |
          slash_*
          SHA256SUMS*
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  publish-docker:
    runs-on: ubuntu-latest
    needs: build-release
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Login to Docker Hub
      uses: docker/login-action@v2
      with:
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}
    
    - name: Build and push
      uses: docker/build-push-action@v4
      with:
        context: .
        push: true
        tags: |
          slash/slash:latest
          slash/slash:${{ github.ref_name }}
        cache-from: type=registry,ref=slash/slash:buildcache
        cache-to: type=registry,ref=slash/slash:buildcache,mode=max
```

---

### Workflow 3: Integration Tests

**File: `.github/workflows/integration-tests.yml`**

```yaml
name: Integration Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  integration:
    runs-on: ubuntu-latest
    
    services:
      # Mock API server for testing
      mock-api:
        image: mockserver/mockserver:latest
        ports:
          - 1080:1080
        env:
          MOCKSERVER_PROPERTY_FILE: /config/mockserver-init.properties
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v3
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: "1.21"
    
    - name: Build binary
      run: make build
    
    - name: Start Slash daemon
      run: ./slash daemon &
      timeout-minutes: 1
    
    - name: Wait for daemon
      run: |
        for i in {1..30}; do
          if [ -S ~/.slash/daemon.sock ]; then
            echo "Daemon started"
            exit 0
          fi
          sleep 1
        done
        echo "Daemon failed to start"
        exit 1
    
    - name: Run integration tests
      run: |
        make test-integration
      env:
        SLASH_DAEMON_SOCKET: ~/.slash/daemon.sock
        MOCK_API_URL: http://localhost:1080
    
    - name: Test Claude Code plugin
      run: |
        ./slash plugin install claude-code
        ./slash stats
    
    - name: Test compression on real project
      run: |
        git clone --depth 1 https://github.com/golang/go.git /tmp/go-repo
        ./slash audit /tmp/go-repo/src --recursive
    
    - name: Collect logs
      if: failure()
      run: |
        cat ~/.slash/daemon.log || true
```

---

## Docker Setup for CI

### Dockerfile for Testing

**File: `Dockerfile.test`**

```dockerfile
FROM golang:1.21-alpine

# Install dependencies
RUN apk add --no-cache \
    git \
    make \
    curl \
    bash \
    ca-certificates

# Set working directory
WORKDIR /app

# Copy source
COPY . .

# Build
RUN make build

# Create non-root user
RUN addgroup -g 1000 slash && \
    adduser -D -u 1000 -G slash slash

# Create directories
RUN mkdir -p /home/slash/.slash /home/slash/.cache/slash && \
    chown -R slash:slash /home/slash

USER slash

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD slash version || exit 1

ENTRYPOINT ["./slash"]
CMD ["daemon"]
```

**Usage:**
```bash
# Build test image
docker build -f Dockerfile.test -t slash:test .

# Run tests in container
docker run --rm slash:test make test

# Run benchmarks
docker run --rm slash:test slash bench
```

---

### Docker Compose for Integration Tests

**File: `docker-compose.test.yml`**

```yaml
version: "3.8"

services:
  # Slash daemon
  slash-daemon:
    build:
      context: .
      dockerfile: Dockerfile.test
    command: daemon
    volumes:
      - slash-cache:/home/slash/.cache/slash
      - slash-config:/home/slash/.slash
    environment:
      - LOG_LEVEL=debug
    healthcheck:
      test: ["CMD", "slash", "version"]
      interval: 10s
      timeout: 5s
      retries: 3
  
  # Test runner
  test-runner:
    build:
      context: .
      dockerfile: Dockerfile.test
    command: make test-integration
    depends_on:
      slash-daemon:
        condition: service_healthy
    volumes:
      - .:/app
      - slash-cache:/home/slash/.cache/slash
      - slash-config:/home/slash/.slash
    environment:
      - SLASH_DAEMON_SOCKET=/home/slash/.slash/daemon.sock
      - CGO_ENABLED=0
  
  # Benchmark runner
  benchmark-runner:
    build:
      context: .
      dockerfile: Dockerfile.test
    command: slash bench --output /results/bench.json
    depends_on:
      slash-daemon:
        condition: service_healthy
    volumes:
      - ./results:/results
      - slash-cache:/home/slash/.cache/slash
  
  # Mock API server
  mock-api:
    image: mockserver/mockserver:latest
    ports:
      - "1080:1080"
    environment:
      MOCKSERVER_INITIALIZATION_JSON_PATH: /config/init.json
    volumes:
      - ./tests/mock-api-config.json:/config/init.json

volumes:
  slash-cache:
  slash-config:
```

**Usage:**
```bash
# Run all tests
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# Run specific service
docker-compose -f docker-compose.test.yml run test-runner

# Run benchmarks
docker-compose -f docker-compose.test.yml run benchmark-runner
```

---

## Testing Strategies

### Test 1: Unit Tests

**File: `internal/compress/compressor_test.go`**

```go
package compress

import (
    "testing"
)

func TestJSONCompression(t *testing.T) {
    tests := []struct {
        name              string
        input             string
        expectedReduction float64 // e.g., 0.75 for 75%
        shouldPass        bool
    }{
        {
            name: "simple_object",
            input: `{"id": 123, "name": "Alice", "email": "alice@example.com"}`,
            expectedReduction: 0.6,
            shouldPass: true,
        },
        {
            name: "nested_object",
            input: `{"user": {"id": 123, "profile": {"avatar": "url", "bio": "text"}}}`,
            expectedReduction: 0.7,
            shouldPass: true,
        },
        {
            name: "array_values",
            input: `{"items": [{"id": 1}, {"id": 2}, {"id": 3}, {"id": 4}, {"id": 5}]}`,
            expectedReduction: 0.65,
            shouldPass: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            compressor := NewCompressor()
            compressed, metadata := compressor.Compress([]byte(tt.input))
            
            if len(compressed) == 0 && tt.shouldPass {
                t.Fatalf("compression returned empty result")
            }
            
            reduction := 1.0 - float64(len(compressed))/float64(len(tt.input))
            if reduction < tt.expectedReduction {
                t.Errorf("reduction %.2f%% < expected %.2f%%", reduction*100, tt.expectedReduction*100)
            }
        })
    }
}

func TestCodeCompression(t *testing.T) {
    input := `package main

import "fmt"

func main() {
    // This is a function with 100 lines of implementation
    // ... lots of code ...
}`
    
    compressor := NewCompressor()
    compressed, metadata := compressor.Compress([]byte(input))
    
    if !contains(string(compressed), "func main") {
        t.Error("function signature lost in compression")
    }
    
    reduction := 1.0 - float64(len(compressed))/float64(len(input))
    if reduction < 0.5 {
        t.Errorf("expected >50% reduction, got %.0f%%", reduction*100)
    }
}

func TestLogCompression(t *testing.T) {
    input := `[ERROR] Connection timeout
[ERROR] Connection timeout
[ERROR] Connection timeout
[WARN] Retry limit exceeded`
    
    compressor := NewCompressor()
    compressed, metadata := compressor.Compress([]byte(input))
    
    // Should deduplicate
    if occurrences(string(compressed), "[ERROR]") > 2 {
        t.Error("failed to deduplicate repeated errors")
    }
}
```

---

### Test 2: Integration Tests

**File: `tests/integration_test.go`**

```go
package tests

import (
    "testing"
    "time"
    "github.com/yoosuf/Slash/internal/daemon"
    "github.com/yoosuf/Slash/internal/client"
)

func TestEndToEndCompression(t *testing.T) {
    // Start daemon
    d := daemon.NewDaemon()
    go d.Start()
    defer d.Stop()
    
    // Wait for daemon
    time.Sleep(100 * time.Millisecond)
    
    // Create client
    c := client.NewHookClient(daemon.DefaultSocketPath(), 5*time.Second)
    
    // Send event
    event := map[string]interface{}{
        "type": "hook:post_call",
        "toolOutput": `{"user": {"id": 123, "name": "Alice", "email": "alice@example.com"}}`,
    }
    
    result, err := c.ProcessHookEvent(event)
    if err != nil {
        t.Fatalf("ProcessHookEvent failed: %v", err)
    }
    
    // Verify compression happened
    if result["compressed"] == nil {
        t.Error("expected compressed output")
    }
    
    if result["handle"] == nil {
        t.Error("expected handle for retrieval")
    }
}

func TestRetrieveUncompressed(t *testing.T) {
    // ... setup ...
    
    // Compress something
    event := map[string]interface{}{
        "type": "hook:post_call",
        "toolOutput": `large JSON response...`,
    }
    
    result, _ := c.ProcessHookEvent(event)
    handle := result["handle"].(string)
    
    // Retrieve original
    retrieved, err := c.Retrieve(handle)
    if err != nil {
        t.Fatalf("Retrieve failed: %v", err)
    }
    
    // Verify it matches original
    if retrieved != event["toolOutput"] {
        t.Error("retrieved content doesn't match original")
    }
}
```

---

### Test 3: Load Tests

**File: `tests/load_test.go`**

```go
package tests

import (
    "fmt"
    "sync"
    "testing"
    "time"
)

func BenchmarkConcurrentCompression(b *testing.B) {
    c := NewTestClient()
    defer c.Close()
    
    // Create realistic test data
    testCases := []string{
        largeJSON,
        largeCode,
        largeLogs,
    }
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        var wg sync.WaitGroup
        
        for j := 0; j < 10; j++ { // 10 concurrent requests
            wg.Add(1)
            go func(j int) {
                defer wg.Done()
                data := testCases[j%len(testCases)]
                c.Compress(data)
            }(j)
        }
        
        wg.Wait()
    }
}

func TestMemoryUsage(t *testing.T) {
    c := NewTestClient()
    defer c.Close()
    
    // Cache should not exceed max_size_mb
    maxMemory := int64(1024 * 1024 * 1024) // 1GB
    
    for i := 0; i < 10000; i++ {
        c.Compress(generateLargeTestData())
        
        mem := c.CacheMemoryUsage()
        if mem > maxMemory {
            t.Fatalf("cache exceeded max memory: %d > %d", mem, maxMemory)
        }
    }
}
```

---

## Performance Benchmarking

### Benchmark Configuration

**File: `.github/workflows/benchmarks.yml`**

```yaml
name: Performance Benchmarks

on:
  push:
    branches: [ main ]
  schedule:
    - cron: '0 0 * * *'  # Daily

jobs:
  benchmark:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - uses: actions/setup-go@v4
      with:
        go-version: "1.21"
    
    - name: Build
      run: make build
    
    - name: Run benchmarks
      run: ./slash bench --output current.json
    
    - name: Download previous benchmarks
      continue-on-error: true
      run: |
        git checkout origin/benchmarks -- benchmarks/ || true
    
    - name: Compare benchmarks
      run: |
        python3 scripts/compare-benchmarks.py \
          --current current.json \
          --baseline benchmarks/baseline.json \
          --output comparison.md
    
    - name: Upload results
      uses: actions/upload-artifact@v3
      with:
        name: benchmark-comparison
        path: comparison.md
    
    - name: Commit results
      if: github.ref == 'refs/heads/main'
      run: |
        cp current.json benchmarks/$(date +%Y-%m-%d).json
        git config user.name "Bot"
        git config user.email "bot@example.com"
        git add benchmarks/
        git commit -m "Benchmark: $(date)" || true
        git push origin benchmarks
```

---

## Integration Testing

### Test Real Editor Plugins

**File: `.github/workflows/plugin-tests.yml`**

```yaml
name: Plugin Integration Tests

on:
  push:
    branches: [ main ]

jobs:
  test-claude-code:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - uses: actions/setup-node@v3
      with:
        node-version: "18"
    
    - name: Build Slash
      run: go build -o slash ./cmd/slash
    
    - name: Install plugin
      run: ./slash plugin install claude-code
    
    - name: Run plugin tests
      run: |
        cd plugin/claude-code
        npm install
        npm test
    
    - name: Verify hook integration
      run: |
        bash tests/verify-claude-code-hooks.sh
```

---

## Continuous Performance Monitoring

### Dashboard Configuration

**File: `monitoring/prometheus.yml`**

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'slash'
    static_configs:
      - targets: ['localhost:8765']

  - job_name: 'slash-benchmarks'
    static_configs:
      - targets: ['localhost:9090']
```

**Grafana Dashboard JSON** (available in repo)

This integrates with CI to:
- Track compression efficiency over time
- Monitor cache hit rates
- Alert on performance regressions
- Generate weekly reports

---

## Summary: CI/CD Strategy

| Stage | Tool | Purpose |
|-------|------|---------|
| **Build** | GitHub Actions | Compile on 3 OSes |
| **Test** | Go testing + Docker | Unit & integration tests |
| **Bench** | Slash bench | Performance tracking |
| **Release** | GitHub Actions | Sign & publish binaries |
| **Docker** | Docker Hub | Image distribution |
| **Monitor** | Prometheus/Grafana | Ongoing performance tracking |

All workflows are production-ready and handle failure cases appropriately.
