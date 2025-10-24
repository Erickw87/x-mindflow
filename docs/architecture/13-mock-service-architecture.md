# Mock Service 架构设计文档

**Issue**: #13  
**版本**: v1.0  
**日期**: 2025-10-24  
**状态**: 设计阶段

---

## 📌 概述

Mock Service 是一个轻量级、可配置的 HTTP 服务模拟器，用于验证 x-mindflow 智能发布系统的核心功能。本文档描述其整体架构、核心模块设计和技术选型。

---

## 🏗️ 整体架构

### 系统架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    Mock Service Instance                     │
│                                                               │
│  ┌────────────────────────────────────────────────────┐     │
│  │              HTTP Server (Gin)                      │     │
│  │  - 动态路由注册                                     │     │
│  │  - 中间件链 (日志、延迟、错误注入)                  │     │
│  │  - 健康检查 /health                                 │     │
│  └────────────────────────────────────────────────────┘     │
│                           ↓                                   │
│  ┌────────────────────────────────────────────────────┐     │
│  │          Configuration Loader                       │     │
│  │  - YAML 解析 (gopkg.in/yaml.v3)                    │     │
│  │  - 配置验证                                         │     │
│  │  - 热加载监听 (可选)                                │     │
│  └────────────────────────────────────────────────────┘     │
│                           ↓                                   │
│  ┌────────────────────────────────────────────────────┐     │
│  │          Request Handler                            │     │
│  │  ┌──────────────┐  ┌──────────────┐               │     │
│  │  │ Static       │  │ Proxy        │               │     │
│  │  │ Response     │  │ Handler      │               │     │
│  │  │ Handler      │  │              │               │     │
│  │  └──────────────┘  └──────────────┘               │     │
│  │                                                     │     │
│  │  - 模板渲染 (text/template)                        │     │
│  │  - 响应延迟                                        │     │
│  │  - 错误注入                                        │     │
│  └────────────────────────────────────────────────────┘     │
│                                                               │
│  ┌────────────────────────────────────────────────────┐     │
│  │          Behavior Simulator                         │     │
│  │  ┌──────────────┐  ┌──────────────┐               │     │
│  │  │ CPU Worker   │  │ Memory Worker│               │     │
│  │  │              │  │              │               │     │
│  │  │ - Busy Loop  │  │ - Alloc Byte │               │     │
│  │  │ - Multi-Core │  │ - Hold Memory│               │     │
│  │  └──────────────┘  └──────────────┘               │     │
│  │                                                     │     │
│  │  ┌──────────────┐  ┌──────────────┐               │     │
│  │  │ Crash Worker │  │ Startup Check│               │     │
│  │  │              │  │              │               │     │
│  │  │ - Delayed    │  │ - Fail Fast  │               │     │
│  │  │   Exit       │  │ - Delay Start│               │     │
│  │  └──────────────┘  └──────────────┘               │     │
│  └────────────────────────────────────────────────────┘     │
│                                                               │
│  ┌────────────────────────────────────────────────────┐     │
│  │          Proxy Handler                              │     │
│  │  - HTTP Client (net/http)                          │     │
│  │  - Header 透传 (X-* 前缀)                          │     │
│  │  - 超时控制                                        │     │
│  │  - 错误传播                                        │     │
│  └────────────────────────────────────────────────────┘     │
│                                                               │
│  ┌────────────────────────────────────────────────────┐     │
│  │          Structured Logging (zap)                   │     │
│  │  - JSON 格式                                        │     │
│  │  - Request ID 追踪                                  │     │
│  │  - 链路日志                                        │     │
│  └────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────┘
                           ↓
         基础设施指标采集 (由外部 node_exporter 完成)
```

---

## 🔧 核心模块设计

### 1. Configuration Loader 模块

**职责**: 加载、解析和验证配置文件

**技术选型**:
- `gopkg.in/yaml.v3`: YAML 解析
- `github.com/go-playground/validator/v10`: 配置验证

**核心接口**:
```go
type ConfigLoader interface {
    Load(path string) (*ServiceConfig, error)
    Validate(config *ServiceConfig) error
}
```

**数据结构**:
```go
type ServiceConfig struct {
    Service      ServiceInfo         `yaml:"service" validate:"required"`
    Behaviors    BehaviorConfig      `yaml:"behaviors"`
    Endpoints    []EndpointConfig    `yaml:"endpoints" validate:"required,dive"`
    Observability ObservabilityConfig `yaml:"observability"`
}

type ServiceInfo struct {
    Name    string `yaml:"name" validate:"required"`
    Version string `yaml:"version" validate:"required,semver"`
    Port    int    `yaml:"port" validate:"required,min=1,max=65535"`
}
```

**错误处理**:
- 文件不存在: 返回 `ErrConfigNotFound`
- YAML 语法错误: 返回详细的解析错误信息
- 验证失败: 返回字段级别的验证错误

---

### 2. HTTP Server 模块

**职责**: 提供 HTTP 服务，动态注册路由

**技术选型**:
- `github.com/gin-gonic/gin`: Web 框架
- Gin 的中间件机制

**核心接口**:
```go
type MockServer interface {
    Start() error
    Shutdown(ctx context.Context) error
    RegisterEndpoints(endpoints []EndpointConfig) error
}
```

**中间件链**:
```go
r := gin.New()
r.Use(
    middleware.RequestID(),      // 生成 Request ID
    middleware.Logger(logger),   // 结构化日志
    middleware.Recovery(),       // Panic 恢复
    middleware.Latency(),        // 延迟注入
    middleware.ErrorInjection(), // 错误注入
)
```

**动态路由注册**:
```go
func (s *MockServer) RegisterEndpoints(endpoints []EndpointConfig) error {
    for _, ep := range endpoints {
        handler := s.createHandler(ep)
        s.engine.Handle(ep.Method, ep.Path, handler)
    }
    return nil
}
```

---

### 3. Request Handler 模块

**职责**: 处理 HTTP 请求，返回响应或代理转发

#### 3.1 Static Response Handler

**功能**: 根据配置返回静态或模板化的响应

**模板引擎**: Go `text/template`

**模板变量**:
```go
type TemplateData struct {
    PathParam  map[string]string  // 路径参数
    QueryParam map[string]string  // 查询参数
    Service    ServiceInfo        // 服务信息
    RandomID   string             // 随机 UUID
    Timestamp  string             // 当前时间戳
}
```

**示例**:
```yaml
response:
  body:
    id: "{{.PathParam.id}}"
    timestamp: "{{.Timestamp}}"
```

#### 3.2 Proxy Handler

**功能**: 转发请求到目标服务

**关键特性**:
1. **Header 透传**: 自动透传 `X-` 开头的 Header
2. **超时控制**: 每个代理配置独立超时
3. **错误传播**: 下游错误码原样返回

**实现**:
```go
func (h *ProxyHandler) Forward(c *gin.Context, config ProxyConfig) {
    // 创建新请求
    req, _ := http.NewRequest(c.Request.Method, config.Target, c.Request.Body)
    
    // 透传 Headers
    h.copyHeaders(c.Request.Header, req.Header, config.PreserveHeaders)
    
    // 发送请求
    client := &http.Client{Timeout: config.Timeout}
    resp, err := client.Do(req)
    
    // 返回响应
    c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}
```

---

### 4. Behavior Simulator 模块

**职责**: 模拟各种运行时行为（CPU、内存、崩溃等）

#### 4.1 CPU Worker

**原理**: 使用 busy loop 消耗 CPU

```go
func (w *CPUWorker) Simulate(targetPercent int, duration time.Duration) {
    numCPU := runtime.NumCPU()
    for i := 0; i < numCPU; i++ {
        go func() {
            start := time.Now()
            for time.Since(start) < duration {
                // Busy loop
                for j := 0; j < 1000000; j++ {
                    _ = j * j
                }
                // Sleep to control CPU usage
                time.Sleep(time.Millisecond * time.Duration(100-targetPercent))
            }
        }()
    }
}
```

**验证**: 通过 `runtime.ReadMemStats()` 或 `/proc/stat` 验证 CPU 占用率

#### 4.2 Memory Worker

**原理**: 分配大块字节数组并持有引用

```go
func (w *MemoryWorker) Simulate(targetMB int, duration time.Duration) {
    data := make([]byte, targetMB*1024*1024)
    // Fill with random data to prevent optimization
    for i := range data {
        data[i] = byte(i % 256)
    }
    
    // Hold for duration
    time.Sleep(duration)
    
    // Release
    runtime.GC()
}
```

#### 4.3 Crash Worker

**功能**: 延迟退出进程

```go
func (w *CrashWorker) Trigger(delay time.Duration) {
    go func() {
        time.Sleep(delay)
        os.Exit(1)  // 或 panic("simulated crash")
    }()
}
```

#### 4.4 Startup Check

**功能**: 启动时检查配置，模拟启动失败

```go
func (s *MockServer) startupCheck(config BehaviorConfig) error {
    if config.Startup.Fail {
        time.Sleep(config.Startup.Delay)
        return errors.New("simulated startup failure")
    }
    return nil
}
```

---

### 5. Proxy Handler 模块

**职责**: 实现链式代理功能

**关键设计**:

```go
type ProxyConfig struct {
    Target          string            `yaml:"target"`
    Path            string            `yaml:"path"`  // 可选: 重写路径
    PreserveHeaders bool              `yaml:"preserve_headers"`
    Timeout         time.Duration     `yaml:"timeout"`
    QueryParams     map[string]string `yaml:"query_params"`
}
```

**Header 透传策略**:
```go
func (h *ProxyHandler) copyHeaders(src, dst http.Header, preserveAll bool) {
    if preserveAll {
        // 复制所有 X- 开头的 Header
        for k, v := range src {
            if strings.HasPrefix(k, "X-") {
                dst[k] = v
            }
        }
    }
    // 始终透传这些 Header
    dst.Set("X-Request-ID", src.Get("X-Request-ID"))
    dst.Set("X-Traffic-Tag", src.Get("X-Traffic-Tag"))
}
```

**链式代理示例**:
```
Client → Service A (user-service:8080)
            ↓ proxy /api/v1/orders → http://order-service:8081
         Service B (order-service:8081)
            ↓ proxy /api/v1/payments → http://payment-service:8082
         Service C (payment-service:8082)
```

---

### 6. Logging 模块

**职责**: 提供结构化日志

**技术选型**: `go.uber.org/zap`

**日志格式**:
```json
{
  "level": "info",
  "ts": "2025-10-24T08:00:00Z",
  "msg": "request handled",
  "request_id": "abc-123",
  "method": "GET",
  "path": "/api/v1/users/1",
  "status": 200,
  "latency": "10ms",
  "proxy_target": "http://order-service:8081"
}
```

**日志级别**:
- `debug`: 详细的请求/响应数据
- `info`: 正常请求日志
- `warn`: 代理超时、下游错误
- `error`: 配置错误、启动失败

---

## 🚀 部署架构

### Docker 部署

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o mock-server cmd/mock-server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /build/mock-server .
COPY configs/ /app/configs/
EXPOSE 8080
CMD ["./mock-server", "--config", "/app/configs/default.yaml"]
```

**镜像优化**:
- 多阶段构建，减小镜像体积
- 使用 Alpine 作为基础镜像
- 预期镜像大小: ~20MB

---

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mock-user-service-v1-2
spec:
  replicas: 2
  template:
    spec:
      containers:
      - name: mock-server
        image: mock-service:v1.0
        ports:
        - containerPort: 8080
        volumeMounts:
        - name: config
          mountPath: /app/configs
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 3
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: mock-user-service-v1-2-config
```

---

## 📊 监控与可观测性

### 健康检查

**端点**: `GET /health`

**响应**:
```json
{
  "status": "healthy",
  "version": "v1.2.0",
  "uptime": "3m45s",
  "timestamp": "2025-10-24T08:00:00Z"
}
```

### 基础设施指标

**采集方式**: 由 node_exporter 采集容器/节点指标

**关键指标**:
- `node_cpu_seconds_total`: CPU 使用时间
- `node_memory_MemTotal_bytes`: 总内存
- `node_memory_MemAvailable_bytes`: 可用内存
- `container_cpu_usage_seconds_total`: 容器 CPU 使用
- `container_memory_usage_bytes`: 容器内存使用

**Prometheus 查询示例**:
```promql
# CPU 使用率
rate(container_cpu_usage_seconds_total{pod="mock-user-service-v1-2"}[1m])

# 内存使用量
container_memory_usage_bytes{pod="mock-user-service-v1-2"}
```

---

## 🔐 安全设计

### 配置安全

- ❌ 不支持动态代理到任意地址
- ✅ 代理目标必须在配置文件中显式声明
- ✅ 配置文件验证防止注入攻击

### 运行时安全

- 使用非 root 用户运行容器
- 只读文件系统（除日志目录）
- 资源限制（CPU/Memory limits）

---

## 🎯 技术选型总结

| 组件           | 技术选型                  | 理由                                      |
|----------------|---------------------------|-------------------------------------------|
| Web 框架       | Gin                       | 高性能、轻量、中间件生态丰富              |
| 配置解析       | gopkg.in/yaml.v3          | Go 官方推荐的 YAML 库                     |
| 日志           | uber-go/zap               | 高性能结构化日志                          |
| HTTP 客户端    | net/http (标准库)         | 无需额外依赖，性能足够                    |
| 模板引擎       | text/template (标准库)    | 简单场景足够，无需复杂模板                |
| 配置验证       | go-playground/validator   | 声明式验证，减少样板代码                  |

---

## 🔄 数据流图

### 静态响应流程

```
Client → Gin Router → Middleware Chain → Static Response Handler
                          ↓                        ↓
                      Request ID             Template Render
                      Logging                     ↓
                      Latency                Response Body
                          ↓                        ↓
                      Gin Context ← ──────────── JSON
```

### 代理请求流程

```
Client → Gin Router → Middleware Chain → Proxy Handler
                          ↓                   ↓
                      Request ID         Create Request
                      Logging            Copy Headers
                          ↓                   ↓
                      Timeout           HTTP Client.Do
                          ↓                   ↓
                      Response ← ────── Downstream Service
```

---

## 📦 项目结构

```
mock-service/
├── cmd/
│   └── mock-server/
│       └── main.go                 # 程序入口，初始化配置和服务
├── internal/
│   ├── server/
│   │   ├── server.go               # HTTP Server 实现
│   │   ├── router.go               # 路由注册
│   │   └── middleware/
│   │       ├── request_id.go       # Request ID 中间件
│   │       ├── logger.go           # 日志中间件
│   │       ├── latency.go          # 延迟注入
│   │       └── error_injection.go  # 错误注入
│   ├── handler/
│   │   ├── static.go               # 静态响应处理器
│   │   ├── proxy.go                # 代理处理器
│   │   └── health.go               # 健康检查
│   ├── behavior/
│   │   ├── simulator.go            # 行为模拟器协调
│   │   ├── cpu.go                  # CPU 负载
│   │   ├── memory.go               # 内存负载
│   │   ├── crash.go                # 崩溃模拟
│   │   └── startup.go              # 启动检查
│   └── template/
│       └── renderer.go             # 模板渲染
├── pkg/
│   ├── config/
│   │   ├── loader.go               # 配置加载器
│   │   ├── types.go                # 配置数据结构
│   │   └── validator.go            # 配置验证
│   └── logger/
│       └── logger.go               # Zap 日志封装
├── configs/
│   ├── user-service-v1.1.yaml      # 示例配置
│   ├── user-service-v1.2.yaml
│   └── order-service-v1.0.yaml
├── Dockerfile
├── docker-compose.yaml
├── k8s/
│   ├── deployment.yaml
│   ├── service.yaml
│   └── configmap.yaml
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

## 🔮 未来扩展

### 阶段 2 功能（可选）

1. **配置热加载**: 使用 `fsnotify` 监听配置文件变更
2. **Metrics 导出**: 可选的 Prometheus metrics（应用级指标）
3. **gRPC 支持**: 支持 gRPC 服务模拟
4. **数据库模拟**: 可选的内存数据存储（SQLite）

### 性能优化

1. **连接池**: HTTP 客户端连接复用
2. **响应缓存**: 静态响应缓存（可选）
3. **并发控制**: Goroutine 池管理

---

**最后更新**: 2025-10-24
