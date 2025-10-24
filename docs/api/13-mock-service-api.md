# Mock Service API 与配置设计文档

**Issue**: #13  
**版本**: v1.0  
**日期**: 2025-10-24  
**状态**: 设计阶段

---

## 📌 概述

本文档详细描述 Mock Service 的配置文件格式、HTTP API 端点规范和命令行接口。

---

## 🔧 配置文件规范

### 配置文件格式

**文件类型**: YAML  
**文件扩展名**: `.yaml` 或 `.yml`  
**编码**: UTF-8

### 完整配置示例

```yaml
# Mock Service 配置文件
# 文件名: user-service-v1.2.yaml

# ============================================
# 服务基本信息
# ============================================
service:
  name: "user-service"          # 服务名称
  version: "v1.2.0"              # 服务版本（遵循 SemVer）
  port: 8080                     # 监听端口

# ============================================
# 行为模拟配置
# ============================================
behaviors:
  # CPU 占用模拟
  cpu:
    enabled: true                # 是否启用
    target_percent: 80           # 目标 CPU 占用率 (0-100)
    trigger:
      type: "endpoint"           # 触发类型: startup | endpoint | periodic
      endpoint: "/api/v1/heavy-compute"  # 触发接口 (type=endpoint 时)
    duration: 30s                # 持续时间

  # 内存占用模拟
  memory:
    enabled: true
    target_mb: 512               # 目标内存占用 (MB)
    trigger:
      type: "endpoint"
      endpoint: "/api/v1/load-data"
    duration: 60s

  # 启动失败模拟
  startup:
    fail: false                  # 是否启动失败
    delay: 5s                    # 启动延迟
    exit_code: 1                 # 退出码 (0-255, 默认 1)

  # 进程崩溃模拟
  crash:
    enabled: false
    trigger:
      type: "endpoint"
      endpoint: "/api/v1/trigger-crash"
    delay: 2s                    # 延迟退出时间

# ============================================
# API 端点定义
# ============================================
endpoints:
  # 1. 简单响应端点
  - path: "/api/v1/health"
    method: "GET"
    response:
      status: 200
      headers:
        Content-Type: "application/json"
      body:
        status: "healthy"
        version: "v1.2.0"
    latency: 10ms                # 响应延迟

  # 2. 动态响应端点（路径参数）
  - path: "/api/v1/users/:id"
    method: "GET"
    response:
      status: 200
      body:
        id: "{{.PathParam.id}}"
        name: "User {{.PathParam.id}}"
        email: "user{{.PathParam.id}}@example.com"
        version: "{{.Service.Version}}"
        created_at: "{{.Timestamp}}"
    latency: 50ms

  # 3. 查询参数示例
  - path: "/api/v1/users"
    method: "GET"
    response:
      status: 200
      body:
        users:
          - id: "1"
            name: "Alice"
          - id: "2"
            name: "Bob"
        page: "{{.QueryParam.page}}"
        limit: "{{.QueryParam.limit}}"
    latency: 100ms

  # 4. POST 请求示例
  - path: "/api/v1/users"
    method: "POST"
    response:
      status: 201
      body:
        id: "{{.RandomID}}"
        created_at: "{{.Timestamp}}"
        message: "User created successfully"
    latency: 100ms

  # 5. 代理端点
  - path: "/api/v1/orders"
    method: "GET"
    proxy:
      target: "http://order-service:8081"    # 代理目标
      path: "/api/v1/orders"                 # 重写路径 (可选)
      preserve_headers: true                 # 保留 Headers
      timeout: 5s                            # 超时时间
      query_params:                          # 额外查询参数 (可选)
        user_id: "{{.PathParam.id}}"

  # 6. 错误注入端点
  - path: "/api/v1/error-test"
    method: "GET"
    response:
      status: 500
      body:
        error: "Internal Server Error"
        message: "This is a simulated error"
    error_rate: 0.5              # 50% 错误率 (0.0-1.0)
    latency: 100ms

  # 7. 随机响应
  - path: "/api/v1/random"
    method: "GET"
    responses:                   # 多个响应定义
      - weight: 70               # 权重 70%
        status: 200
        body:
          result: "success"
      - weight: 20               # 权重 20%
        status: 400
        body:
          error: "Bad Request"
      - weight: 10               # 权重 10%
        status: 500
        body:
          error: "Internal Server Error"

# ============================================
# 版本特定行为 (可选)
# ============================================
version_behaviors:
  "v1.1.0":
    endpoints:
      - path: "/api/v1/health"
        response:
          status: 200
          body:
            status: "healthy"
    behaviors:
      cpu:
        enabled: false
  
  "v1.2.0":
    endpoints:
      - path: "/api/v1/health"
        response:
          status: 503
          body:
            status: "unhealthy"
    behaviors:
      cpu:
        enabled: true
        target_percent: 90

# ============================================
# 可观测性配置
# ============================================
observability:
  health_check:
    enabled: true
    port: 8080                   # 健康检查端口（与服务端口相同）
    path: "/health"              # 健康检查路径
  
  logging:
    level: "info"                # debug | info | warn | error
    format: "json"               # json | console
    output: "stdout"             # stdout | file
    file_path: "/var/log/mock-service.log"  # 日志文件路径 (output=file 时)
```

---

## 📖 配置字段详解

### 1. `service` - 服务基本信息

| 字段      | 类型   | 必填 | 说明                               |
|-----------|--------|------|------------------------------------|
| `name`    | string | 是   | 服务名称                           |
| `version` | string | 是   | 服务版本（遵循 SemVer 2.0）        |
| `port`    | int    | 是   | 监听端口 (1-65535)                 |

### 2. `behaviors` - 行为模拟配置

#### 2.1 `cpu` - CPU 负载模拟

| 字段             | 类型   | 必填 | 说明                                |
|------------------|--------|------|-------------------------------------|
| `enabled`        | bool   | 是   | 是否启用                            |
| `target_percent` | int    | 是   | 目标 CPU 占用率 (0-100)             |
| `trigger.type`   | string | 是   | 触发类型: `startup` / `endpoint` / `periodic` |
| `trigger.endpoint` | string | 否 | 触发接口 (type=endpoint 时必填)     |
| `duration`       | duration | 是 | 持续时间 (如 `30s`, `5m`)          |

#### 2.2 `memory` - 内存负载模拟

| 字段             | 类型   | 必填 | 说明                                |
|------------------|--------|------|-------------------------------------|
| `enabled`        | bool   | 是   | 是否启用                            |
| `target_mb`      | int    | 是   | 目标内存占用 (MB)                   |
| `trigger.type`   | string | 是   | 触发类型                            |
| `trigger.endpoint` | string | 否 | 触发接口                            |
| `duration`       | duration | 是 | 持续时间                            |

#### 2.3 `startup` - 启动行为

| 字段        | 类型     | 必填 | 说明                     |
|-------------|----------|------|--------------------------|
| `fail`      | bool     | 否   | 是否启动失败 (默认 false)     |
| `delay`     | duration | 否   | 启动延迟 (默认 0s)       |
| `exit_code` | int      | 否   | 退出码 (0-255, 默认 1) |

#### 2.4 `crash` - 崩溃模拟

| 字段             | 类型   | 必填 | 说明                                |
|------------------|--------|------|-------------------------------------|
| `enabled`        | bool   | 是   | 是否启用                            |
| `trigger.type`   | string | 是   | 触发类型                            |
| `trigger.endpoint` | string | 否 | 触发接口                            |
| `delay`          | duration | 是 | 延迟退出时间                        |

### 3. `endpoints` - API 端点定义

#### 3.1 基础字段

| 字段     | 类型   | 必填 | 说明                          |
|----------|--------|------|-------------------------------|
| `path`   | string | 是   | 接口路径（支持路径参数）      |
| `method` | string | 是   | HTTP 方法（GET/POST/PUT/DELETE）|
| `latency`| duration | 否 | 响应延迟 (默认 0ms)           |

#### 3.2 静态响应配置

| 字段              | 类型   | 必填 | 说明                          |
|-------------------|--------|------|-------------------------------|
| `response.status` | int    | 是   | HTTP 状态码                   |
| `response.headers`| map    | 否   | 响应 Headers                  |
| `response.body`   | any    | 否   | 响应体（JSON 对象）           |
| `error_rate`      | float  | 否   | 错误率 (0.0-1.0, 默认 0)      |

#### 3.3 代理配置

| 字段                | 类型   | 必填 | 说明                          |
|---------------------|--------|------|-------------------------------|
| `proxy.target`      | string | 是   | 代理目标地址                  |
| `proxy.path`        | string | 否   | 重写路径                      |
| `proxy.preserve_headers` | bool | 否 | 是否保留 Headers (默认 true) |
| `proxy.timeout`     | duration | 否 | 超时时间 (默认 5s)           |
| `proxy.query_params`| map    | 否   | 额外查询参数                  |

#### 3.4 随机响应配置

| 字段          | 类型  | 必填 | 说明                          |
|---------------|-------|------|-------------------------------|
| `responses`   | array | 是   | 多个响应定义                  |
| `responses[].weight` | int | 是 | 权重（总和必须为 100）          |
| `responses[].status` | int | 是 | HTTP 状态码                   |
| `responses[].body` | any | 否 | 响应体                        |

**注意**: 所有响应的权重总和必须等于 100。

### 4. `version_behaviors` - 版本特定行为

**说明**: 为不同版本覆盖配置

**示例**:
```yaml
version_behaviors:
  "v1.1.0":
    endpoints:
      - path: "/api/v1/health"
        response:
          status: 200
```

### 5. `observability` - 可观测性配置

#### 5.1 `health_check`

| 字段      | 类型   | 必填 | 说明                          |
|-----------|--------|------|-------------------------------|
| `enabled` | bool   | 否   | 是否启用 (默认 true)          |
| `port`    | int    | 否   | 端口 (默认同 service.port)    |
| `path`    | string | 否   | 路径 (默认 `/health`)         |

#### 5.2 `logging`

| 字段        | 类型   | 必填 | 说明                          |
|-------------|--------|------|-------------------------------|
| `level`     | string | 否   | 日志级别 (默认 `info`)        |
| `format`    | string | 否   | 格式 (默认 `json`)            |
| `output`    | string | 否   | 输出 (默认 `stdout`)          |
| `file_path` | string | 否   | 文件路径 (output=file 时)     |

---

## 🎨 模板语法

Mock Service 支持在响应体中使用 Go `text/template` 语法。

### 可用变量

| 变量                 | 类型   | 说明                 | 示例                      |
|----------------------|--------|----------------------|---------------------------|
| `{{.PathParam.xxx}}` | string | 路径参数             | `{{.PathParam.id}}`       |
| `{{.QueryParam.xxx}}`| string | 查询参数             | `{{.QueryParam.page}}`    |
| `{{.Service.Name}}`  | string | 服务名称             | `user-service`            |
| `{{.Service.Version}}`| string | 服务版本            | `v1.2.0`                  |
| `{{.RandomID}}`      | string | 随机 UUID            | `a1b2c3d4-...`            |
| `{{.Timestamp}}`     | string | 当前时间戳 (ISO 8601)| `2025-10-24T08:00:00Z`    |

### 示例

```yaml
response:
  body:
    user_id: "{{.PathParam.id}}"
    service: "{{.Service.Name}}"
    version: "{{.Service.Version}}"
    request_id: "{{.RandomID}}"
    created_at: "{{.Timestamp}}"
```

**响应示例**:
```json
{
  "user_id": "123",
  "service": "user-service",
  "version": "v1.2.0",
  "request_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "created_at": "2025-10-24T08:00:00Z"
}
```

---

## 🌐 HTTP API 端点

### 1. 健康检查

**端点**: `GET /health`

**描述**: 返回服务健康状态

**响应**:
```json
{
  "status": "healthy",
  "version": "v1.2.0",
  "uptime": "3m45s",
  "timestamp": "2025-10-24T08:00:00Z"
}
```

**状态码**:
- `200 OK`: 服务健康
- `503 Service Unavailable`: 服务不健康（由配置控制）

---

### 2. 配置的动态端点

根据配置文件动态注册的端点，具体行为由配置决定。

**示例**: `GET /api/v1/users/:id`

**请求**:
```bash
curl http://localhost:8080/api/v1/users/123
```

**响应**:
```json
{
  "id": "123",
  "name": "User 123",
  "email": "user123@example.com",
  "version": "v1.2.0",
  "created_at": "2025-10-24T08:00:00Z"
}
```

---

## 🖥️ 命令行接口

### 启动命令

```bash
mock-server [flags]
```

### 参数说明

| 参数               | 短参数 | 类型   | 必填 | 默认值            | 说明              |
|--------------------|--------|--------|------|-------------------|-------------------|
| `--config`         | `-c`   | string | 是   | -                 | 配置文件路径      |
| `--port`           | `-p`   | int    | 否   | 配置文件中的值    | 覆盖监听端口      |
| `--log-level`      | `-l`   | string | 否   | 配置文件中的值    | 覆盖日志级别      |
| `--version`        | `-v`   | bool   | 否   | false             | 显示版本信息      |
| `--help`           | `-h`   | bool   | 否   | false             | 显示帮助信息      |

### 使用示例

```bash
# 1. 基本用法
./mock-server --config configs/user-service-v1.2.yaml

# 2. 覆盖端口
./mock-server -c configs/user-service-v1.2.yaml -p 8081

# 3. 设置日志级别
./mock-server -c configs/user-service-v1.2.yaml -l debug

# 4. 查看版本
./mock-server --version

# 5. 查看帮助
./mock-server --help
```

---

## 🐳 Docker 运行

```bash
# 使用默认配置
docker run -p 8080:8080 \
  -v $(pwd)/configs:/app/configs \
  mock-service:v1.0 \
  --config /app/configs/user-service-v1.2.yaml

# 覆盖端口
docker run -p 9090:9090 \
  -v $(pwd)/configs:/app/configs \
  mock-service:v1.0 \
  --config /app/configs/user-service-v1.2.yaml \
  --port 9090
```

---

## ☸️ Kubernetes 部署

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: mock-user-service-v1-2-config
data:
  config.yaml: |
    service:
      name: "user-service"
      version: "v1.2.0"
      port: 8080
    behaviors:
      cpu:
        enabled: true
        target_percent: 80
        trigger:
          type: "startup"
        duration: 300s
    endpoints:
      - path: "/api/v1/health"
        method: "GET"
        response:
          status: 200
          body:
            status: "healthy"
```

### Deployment

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
        args:
          - "--config"
          - "/app/configs/config.yaml"
          - "--log-level"
          - "info"
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
        resources:
          limits:
            cpu: "1"
            memory: "512Mi"
          requests:
            cpu: "100m"
            memory: "128Mi"
      volumes:
      - name: config
        configMap:
          name: mock-user-service-v1-2-config
```

---

## 🧪 配置验证

### 验证规则

1. **服务信息**:
   - `name`: 非空字符串
   - `version`: 符合 SemVer 2.0 规范
   - `port`: 1-65535

2. **端点定义**:
   - `path`: 必须以 `/` 开头
   - `method`: 必须是有效的 HTTP 方法
   - `response` 和 `proxy` 二选一

3. **代理配置**:
   - `target`: 必须是有效的 URL
   - `timeout`: 必须 > 0

4. **行为模拟**:
   - `cpu.target_percent`: 0-100
   - `memory.target_mb`: > 0
   - `duration`: 必须可解析为 time.Duration

### 验证工具

```bash
# 验证配置文件
./mock-server validate --config configs/user-service-v1.2.yaml
```

**成功输出**:
```
✓ Configuration is valid
  Service: user-service v1.2.0
  Endpoints: 5
  Behaviors: cpu, memory
```

**失败输出**:
```
✗ Configuration validation failed
  - service.port: must be between 1 and 65535
  - endpoints[0].path: must start with '/'
  - behaviors.cpu.target_percent: must be between 0 and 100
```

---

## 📊 完整配置示例集

### 示例 1: 稳定版本 (v1.1.0)

```yaml
service:
  name: "user-service"
  version: "v1.1.0"
  port: 8081

endpoints:
  - path: "/api/v1/health"
    method: "GET"
    response:
      status: 200
      body:
        status: "healthy"
  
  - path: "/api/v1/users/:id"
    method: "GET"
    response:
      status: 200
      body:
        id: "{{.PathParam.id}}"
        name: "User {{.PathParam.id}}"
    latency: 50ms

observability:
  logging:
    level: "info"
    format: "json"
```

### 示例 2: 高负载版本 (v1.2.0)

```yaml
service:
  name: "user-service"
  version: "v1.2.0"
  port: 8082

behaviors:
  cpu:
    enabled: true
    target_percent: 90
    trigger:
      type: "startup"
    duration: 600s

endpoints:
  - path: "/api/v1/health"
    method: "GET"
    response:
      status: 200
      body:
        status: "healthy"

observability:
  logging:
    level: "warn"
```

### 示例 3: 链式代理配置

```yaml
# User Service
service:
  name: "user-service"
  port: 8080

endpoints:
  - path: "/api/v1/users/:id/orders"
    method: "GET"
    proxy:
      target: "http://order-service:8081"
      path: "/api/v1/orders"
      preserve_headers: true
      query_params:
        user_id: "{{.PathParam.id}}"
```

```yaml
# Order Service
service:
  name: "order-service"
  port: 8081

endpoints:
  - path: "/api/v1/orders"
    method: "GET"
    proxy:
      target: "http://payment-service:8082"
      path: "/api/v1/payments/verify"
      preserve_headers: true
```

---

**最后更新**: 2025-10-24
