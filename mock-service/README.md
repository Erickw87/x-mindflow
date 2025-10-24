# Mock Service

Mock Service 是一个轻量级、可配置的 HTTP 服务模拟器，用于验证 [x-mindflow](../) 智能发布系统的核心功能。

## 📌 功能特性

- ✅ **可配置的 API 端点**: 通过 YAML 配置文件定义 HTTP 接口，无需编写代码
- ✅ **链式代理**: 支持 A → B → C 多层代理，自动透传流量标记 (`X-Traffic-Tag`)
- ✅ **行为模拟**: 模拟 CPU/内存负载、进程崩溃、启动失败等异常场景
- ✅ **模板渲染**: 支持动态响应，内置路径参数、查询参数、时间戳等变量
- ✅ **多部署方式**: 支持本地运行、Docker、Kubernetes 部署
- ✅ **结构化日志**: 基于 zap 的 JSON 格式日志，支持请求链路追踪

## 🚀 快速开始

### 前置要求

- Go 1.24+
- Docker (可选)
- Kubernetes (可选)

### 本地运行

1. **初始化项目**

```bash
make init
```

2. **构建二进制**

```bash
make build
```

3. **运行服务**

```bash
./bin/mock-server --config configs/user-service-v1.1.yaml
```

4. **测试接口**

```bash
# 健康检查
curl http://localhost:8081/health

# 获取用户信息
curl http://localhost:8081/api/v1/users/123
```

### Docker 运行

```bash
# 构建镜像
make docker-build

# 运行容器
docker run -p 8080:8081 \
  -v $(pwd)/configs:/app/configs \
  mock-service:latest \
  --config /app/configs/user-service-v1.1.yaml
```

### Docker Compose 多实例

```bash
# 启动所有服务 (v1.1, v1.2)
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## 📄 配置文件

### 最小配置示例

```yaml
service:
  name: "user-service"
  version: "v1.0.0"
  port: 8080

endpoints:
  - path: "/health"
    method: "GET"
    response:
      status: 200
      body:
        status: "healthy"

observability:
  logging:
    level: "info"
    format: "json"
```

### 完整配置示例

参见 [configs/user-service-v1.2.yaml](configs/user-service-v1.2.yaml) 和 [API 设计文档](../docs/api/13-mock-service-api.md)。

## 🎯 使用场景

### 场景 1: 模拟多版本并行发布

启动两个版本的服务：

```bash
# v1.1.0 (稳定版本)
./bin/mock-server --config configs/user-service-v1.1.yaml &

# v1.2.0 (高 CPU 占用版本)
./bin/mock-server --config configs/user-service-v1.2.yaml &
```

x-mindflow 发布系统可以：
- 配置 LB 将 10% 流量导向 v1.2.0
- 监控 v1.2.0 的 CPU 指标 (via node_exporter)
- 发现异常后自动回滚到 v1.1.0

### 场景 2: 测试链式代理与流量标记

配置文件 `user-service.yaml`:

```yaml
endpoints:
  - path: "/api/v1/users/:id/orders"
    method: "GET"
    proxy:
      target: "http://order-service:8081"
      preserve_headers: true
```

发送带流量标记的请求：

```bash
curl -H "X-Traffic-Tag: canary-v1.2" \
     http://localhost:8080/api/v1/users/123/orders
```

验证 `X-Traffic-Tag` header 在整个调用链中传递。

### 场景 3: 模拟服务启动失败

配置文件 `crash-service.yaml`:

```yaml
behaviors:
  startup:
    fail: true
    delay: 5s
    exit_code: 1
```

测试 x-mindflow 的故障检测与回滚机制。

## ☸️ Kubernetes 部署

1. **创建 ConfigMap**

```bash
kubectl apply -f k8s/configmap.yaml
```

2. **部署服务**

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
```

3. **验证部署**

```bash
# 查看 Pod
kubectl get pods -l app=mock-user-service

# 查看服务
kubectl get svc -l app=mock-user-service

# 测试服务
kubectl port-forward svc/mock-user-service 8080:80
curl http://localhost:8080/health
```

## 🧪 测试

```bash
# 运行测试
make test

# 运行 lint
make lint
```

## 📊 监控

Mock Service 不导出应用级 Prometheus metrics，依赖 node_exporter 采集基础设施指标：

- `container_cpu_usage_seconds_total`: 容器 CPU 使用
- `container_memory_usage_bytes`: 容器内存使用

Prometheus 查询示例：

```promql
# CPU 使用率
rate(container_cpu_usage_seconds_total{pod=~"mock-user-service.*"}[1m])

# 内存使用量
container_memory_usage_bytes{pod=~"mock-user-service.*"}
```

## 🔧 命令行参数

```bash
mock-server [flags]

Flags:
  --config string      配置文件路径 (必填)
  --port int          覆盖配置文件中的端口
  --log-level string  覆盖日志级别 (debug, info, warn, error)
  --version           显示版本信息
  --help              显示帮助信息
```

## 📚 文档

- [需求分析文档](../docs/requirements/13-mock-service-requirement.md)
- [架构设计文档](../docs/architecture/13-mock-service-architecture.md)
- [API/配置设计文档](../docs/api/13-mock-service-api.md)
- [模块设计文档](../docs/modules/13-mock-service-modules.md)

## 🛠️ 开发

### 项目结构

```
mock-service/
├── cmd/
│   └── mock-server/       # 程序入口
├── internal/
│   ├── server/            # HTTP 服务器
│   ├── handler/           # 请求处理器
│   ├── behavior/          # 行为模拟器
│   └── template/          # 模板渲染
├── pkg/
│   ├── config/            # 配置管理
│   └── logger/            # 日志
├── configs/               # 示例配置
├── k8s/                   # Kubernetes manifests
└── Dockerfile
```

### 技术栈

- **Web 框架**: Gin
- **日志**: uber-go/zap
- **配置解析**: gopkg.in/yaml.v3
- **验证**: go-playground/validator

## 🤝 贡献

参见 [CLAUDE.md](../CLAUDE.md) 了解开发规范。

## 📄 许可证

[MIT License](LICENSE)

---

**相关链接**:
- [x-mindflow 主项目](../)
- [Issue #13](https://github.com/LiusCraft/x-mindflow/issues/13)
- [设计文档 PR #15](https://github.com/LiusCraft/x-mindflow/pull/15)
