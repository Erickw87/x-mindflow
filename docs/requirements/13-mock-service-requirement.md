# Mock Service 需求分析文档

**Issue**: #13  
**作者**: LiusCraft  
**日期**: 2025-10-24  
**状态**: 已确认

---

## 📌 背景

x-mindflow 是一个企业级智能发布系统，支持多版本并行部署、统一多环境管理和自动化故障回滚。为了验证发布系统的核心功能（多版本并行发布、灰度发布、流量控制、故障回滚等），需要一个可配置的 Mock 服务程序来模拟真实业务服务的行为。

### 为什么需要 Mock Service?

1. **测试环境隔离**: 避免在真实业务服务上进行发布系统测试
2. **行为可控**: 能够精确模拟各种正常和异常场景
3. **成本优化**: 无需部署完整的业务服务栈
4. **快速迭代**: 通过配置文件快速调整测试场景

---

## 🎯 目标

开发一个轻量级、可配置的 Mock 服务程序，用于验证 x-mindflow 发布系统的以下核心能力：

- ✅ 多版本并行发布
- ✅ 灰度发布与流量控制
- ✅ 基于 LB 的流量标记传递
- ✅ 链式服务调用
- ✅ 异常场景下的故障检测与回滚

---

## 📋 功能需求

### 1. 可配置的 API 端点

**需求描述**:  
通过 YAML 配置文件声明 Mock 服务暴露的 HTTP 接口，无需编写代码即可定义 API 行为。

**用户故事**:  
> 作为测试工程师，我希望通过配置文件定义 Mock 服务的 API 端点，这样我可以快速模拟不同的服务接口，而无需修改代码。

**功能点**:
- 支持定义 RESTful API 端点（GET, POST, PUT, DELETE 等）
- 支持路径参数（如 `/users/:id`）
- 支持自定义响应状态码和响应体
- 支持响应延迟配置
- 支持动态模板渲染（如时间戳、随机 ID）

**示例配置**:
```yaml
endpoints:
  - path: "/api/v1/users/:id"
    method: "GET"
    response:
      status: 200
      body:
        id: "{{.PathParam.id}}"
        name: "User {{.PathParam.id}}"
    latency: 50ms
```

---

### 2. 链式代理功能

**需求描述**:  
Mock 服务能够将某些接口请求代理到其他 Mock 服务实例，形成链式调用（A → B → C），模拟微服务架构中的服务间调用。

**用户故事**:  
> 作为测试工程师，我希望 Mock 服务 A 能够将请求转发到 Mock 服务 B，服务 B 再转发到服务 C，这样我可以测试流量标记在服务链中的传递行为。

**功能点**:
- 支持 HTTP 请求转发到指定目标服务
- 自动透传 HTTP Headers（特别是 `X-Traffic-Tag` 等流量标记）
- 支持路径重写
- 支持超时控制
- 支持错误传播（下游错误原样返回）

**示例配置**:
```yaml
endpoints:
  - path: "/api/v1/orders"
    method: "GET"
    proxy:
      target: "http://order-service:8081"
      preserve_headers: true
      timeout: 5s
```

**验收标准**:
- [ ] 支持至少 3 层链式代理（A → B → C）
- [ ] `X-Traffic-Tag` header 能够在整个调用链中传递
- [ ] 代理超时后返回 504 Gateway Timeout
- [ ] 下游服务返回的错误码能原样传递给上游

---

### 3. 行为模拟

**需求描述**:  
根据配置模拟各种服务行为，包括正常运行和异常场景，用于测试发布系统的监控和回滚能力。

#### 3.1 CPU 负载模拟

**用户故事**:  
> 作为 SRE 工程师，我希望模拟某个版本的服务在运行时 CPU 占用率异常升高，这样我可以验证发布系统是否能检测到异常并触发回滚。

**功能点**:
- 支持指定目标 CPU 占用率（如 80%）
- 支持触发方式：启动时、访问特定接口时、周期性
- 支持持续时间配置

**示例配置**:
```yaml
behaviors:
  cpu:
    enabled: true
    target_percent: 80
    trigger:
      type: "endpoint"
      endpoint: "/api/v1/heavy-compute"
    duration: 30s
```

#### 3.2 内存负载模拟

**用户故事**:  
> 作为 SRE 工程师，我希望模拟某个版本的服务内存泄漏，这样我可以验证发布系统是否能检测到内存异常。

**功能点**:
- 支持指定目标内存占用（MB）
- 支持触发方式：启动时、访问特定接口时
- 支持持续时间配置

**示例配置**:
```yaml
behaviors:
  memory:
    enabled: true
    target_mb: 512
    trigger:
      type: "endpoint"
      endpoint: "/api/v1/load-data"
    duration: 60s
```

#### 3.3 进程崩溃模拟

**用户故事**:  
> 作为 SRE 工程师，我希望模拟某个版本的服务在运行过程中崩溃退出，这样我可以验证发布系统的健康检查和自动重启机制。

**功能点**:
- 支持延迟退出（如 2 秒后退出）
- 支持通过接口触发崩溃
- 支持配置退出码

**示例配置**:
```yaml
behaviors:
  crash:
    enabled: true
    trigger:
      type: "endpoint"
      endpoint: "/api/v1/trigger-crash"
    delay: 2s
```

#### 3.4 启动失败模拟

**用户故事**:  
> 作为 SRE 工程师，我希望模拟某个版本的服务启动失败（如依赖检查失败），这样我可以验证发布系统是否能阻止失败版本的部署。

**功能点**:
- 支持配置启动失败
- 支持启动延迟
- 支持配置退出码

**示例配置**:
```yaml
behaviors:
  startup:
    fail: true
    delay: 5s
    exit_code: 1
```

**验收标准**:
- [ ] CPU 模拟能让 node_exporter 监测到 CPU 占用率达到目标值 ±5%
- [ ] 内存模拟能让 node_exporter 监测到内存占用达到目标值 ±10MB
- [ ] 进程崩溃能被 Kubernetes/Docker 检测到容器退出
- [ ] 启动失败能阻止容器进入 Ready 状态

---

### 4. 版本特定行为

**需求描述**:  
支持为不同版本的 Mock 服务配置不同的行为，模拟版本升级时的差异。

**用户故事**:  
> 作为测试工程师，我希望 v1.1.0 版本健康检查返回 200，而 v1.2.0 版本健康检查返回 503，这样我可以测试灰度发布时的流量切换逻辑。

**功能点**:
- 支持基于版本号覆盖端点配置
- 支持基于版本号覆盖行为配置

**示例配置**:
```yaml
service:
  version: "v1.2.0"

version_behaviors:
  "v1.1.0":
    endpoints:
      - path: "/api/v1/health"
        response:
          status: 200
  
  "v1.2.0":
    endpoints:
      - path: "/api/v1/health"
        response:
          status: 503
```

---

### 5. 健康检查端点

**需求描述**:  
提供标准的健康检查端点，供 Kubernetes/Docker 进行存活探测和就绪探测。

**功能点**:
- 提供 `/health` 端点
- 支持配置返回状态码（用于模拟不健康状态）

**示例响应**:
```json
{
  "status": "healthy",
  "version": "v1.2.0",
  "timestamp": "2025-10-24T08:00:00Z"
}
```

---

## 🚫 非功能需求

### 性能要求

- **启动时间**: ≤ 3 秒
- **内存占用**: 基础状态 ≤ 50MB（不含行为模拟）
- **响应延迟**: 基础响应延迟 ≤ 10ms（不含配置的 latency）
- **并发能力**: 单实例支持 ≥ 1000 QPS

### 可靠性

- **配置错误处理**: 配置文件语法错误时，应输出清晰的错误信息并拒绝启动
- **代理容错**: 下游服务不可达时，应返回明确的错误信息，而不是崩溃

### 可维护性

- **日志**: 使用结构化日志（JSON 格式），包含 Request ID 追踪
- **配置热加载**: 支持运行时重新加载配置文件（可选，非 MVP）

### 部署要求

- **部署方式**: 支持 Docker 容器部署、Kubernetes Deployment、本地二进制运行
- **镜像大小**: Docker 镜像 ≤ 100MB
- **平台支持**: Linux AMD64/ARM64

---

## 🔒 安全需求

- **无敏感信息**: 配置文件中不应包含任何敏感信息（密钥、密码等）
- **代理限制**: 代理目标必须在配置文件中显式声明，不支持动态代理到任意地址

---

## 📊 验收标准

### 场景 1: 多版本并行发布

**前置条件**:
- 启动 2 个 Mock 服务实例（v1.1.0 和 v1.2.0）
- v1.1.0 CPU 占用正常（< 20%）
- v1.2.0 CPU 占用异常（> 80%）

**验证步骤**:
1. x-mindflow 配置 LB 将 10% 流量导向 v1.2.0
2. 监控系统（node_exporter + Prometheus）采集 CPU 指标
3. x-mindflow 检测到 v1.2.0 异常
4. x-mindflow 自动回滚流量到 v1.1.0

**预期结果**:
- [ ] node_exporter 能采集到 v1.2.0 的 CPU 占用率 > 80%
- [ ] x-mindflow 能检测到异常并触发回滚
- [ ] 回滚后所有流量导向 v1.1.0

---

### 场景 2: 流量标记与链式代理

**前置条件**:
- 启动 3 个 Mock 服务（user-service → order-service → payment-service）
- 配置链式代理关系

**验证步骤**:
1. 客户端发送请求到 user-service，携带 `X-Traffic-Tag: canary-v1.2`
2. user-service 代理到 order-service
3. order-service 代理到 payment-service

**预期结果**:
- [ ] payment-service 能接收到 `X-Traffic-Tag: canary-v1.2` header
- [ ] 所有服务的日志中都记录了相同的 Request ID
- [ ] 链路追踪完整

---

### 场景 3: 启动失败检测

**前置条件**:
- Mock 服务 v1.3.0 配置 `startup.fail: true`

**验证步骤**:
1. 尝试启动 v1.3.0 实例
2. Kubernetes 进行健康检查

**预期结果**:
- [ ] 容器启动后立即退出（或健康检查失败）
- [ ] Kubernetes 标记 Pod 为 NotReady
- [ ] x-mindflow 检测到启动失败，阻止流量导入

---

## 📦 交付物

### 第一阶段（MVP）
- [ ] 配置文件解析功能
- [ ] HTTP Server 与动态路由注册
- [ ] 基本 JSON 响应
- [ ] 健康检查端点
- [ ] Dockerfile

### 第二阶段
- [ ] CPU/内存负载模拟
- [ ] 进程崩溃/启动失败模拟
- [ ] 基础代理功能

### 第三阶段
- [ ] 链式代理（3 层）
- [ ] Header 透传
- [ ] Kubernetes 部署配置

---

## 📚 参考资料

- [x-mindflow README](../../README.md)
- [x-mindflow 架构设计](../architecture/)
- [CLAUDE.md 开发规范](../../CLAUDE.md)

---

**最后更新**: 2025-10-24
