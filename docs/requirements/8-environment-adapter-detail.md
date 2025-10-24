# 环境适配层详细设计

## 文档概述

本文档详细说明环境适配层的技术实现方案,包括 Kubernetes 适配器和 BareMetal 适配器的设计细节。

## 架构定位

环境适配层是逻辑层与执行层之间的"桥梁",统一实现同一接口定义:

```go
type EnvironmentAdapter interface {
    CreateInstances(spec TemplateSpec, count int) ([]InstanceInfo, error)
    DeleteInstances(selector map[string]string) error
    GetInstanceStatus(selector map[string]string) ([]InstanceStatus, error)
    UpdateTraffic(serviceName string, weights map[string]float64) error
    HealthCheck(serviceName string) (HealthSummary, error)
}
```

## 一、Kubernetes 适配器详细设计

### 1.1 核心能力

#### 实例管理
- 通过 Deployment 管理无状态服务
- 通过 StatefulSet 管理有状态服务
- 支持动态扩缩容

#### 流量控制
- 通过 Service 实现服务发现
- 通过 Ingress 实现外部流量接入
- 通过权重注解控制流量分配
- 集成 Argo Rollouts 或自定义 CRD 实现动态权重调整

#### 状态监控
- 读取 Pod 状态实时上报
- 集成健康探针(Liveness/Readiness)
- 监控 Pod 启动失败、重启等异常事件

### 1.2 实现方式

#### 技术选型
- **SDK**: 使用 `client-go` 或 Kubernetes REST API
- **控制策略**: 所有复杂控制交给 K8s 自身
  - Scheduler 负责调度
  - Controller 负责状态协调
  - Network 负责网络管理
- **发布系统职责**: 仅负责状态对齐和回滚控制

#### 关键功能实现

**YAML 模板管理**:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .ServiceName }}-{{ .Version }}
  labels:
    app: {{ .ServiceName }}
    version: {{ .Version }}
spec:
  replicas: {{ .ReplicaCount }}
  selector:
    matchLabels:
      app: {{ .ServiceName }}
      version: {{ .Version }}
  template:
    metadata:
      labels:
        app: {{ .ServiceName }}
        version: {{ .Version }}
    spec:
      containers:
      - name: {{ .ServiceName }}
        image: {{ .Image }}
        ports:
        - containerPort: {{ .Port }}
        livenessProbe:
          httpGet:
            path: {{ .HealthCheckPath }}
            port: {{ .Port }}
          initialDelaySeconds: 30
          periodSeconds: 10
```

**动态权重调整示例**:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: {{ .ServiceName }}
  annotations:
    traffic.weight.v1: "90"
    traffic.weight.v2: "10"
```

**Pod 健康探针集成**:
- Liveness Probe: 检测 Pod 是否存活,失败则重启
- Readiness Probe: 检测 Pod 是否就绪,未就绪则从 Service 摘除
- Startup Probe: 检测应用启动完成,避免启动慢的应用被误杀

### 1.3 支持的资源类型

| 资源类型 | 用途 | 是否支持 |
|---------|------|---------|
| Deployment | 无状态应用部署 | ✅ 支持 |
| StatefulSet | 有状态应用部署 | ✅ 支持 |
| Service | 服务发现和负载均衡 | ✅ 支持 |
| Ingress | 外部流量接入 | ✅ 支持 |
| ConfigMap | 配置管理 | ✅ 支持 |
| Secret | 密钥管理 | ✅ 支持 |

---

## 二、BareMetal 适配器详细设计

### 2.1 架构概览

BareMetal 适配器需要三个核心组件协同工作:

```
┌─────────────────────────────────────────┐
│        统一发布抽象层 (Logic Layer)        │
└────────────────┬────────────────────────┘
                 │ EnvironmentAdapter Interface
                 │
            ┌────▼──────┐
            │ BareMetal │
            │  Adapter  │
            └─────┬─────┘
                  │
     ┌────────────┼────────────┐
     │            │            │
┌────▼─────┐ ┌───▼──────┐ ┌──▼────────┐
│  Agent   │ │ Registry │ │Edge Proxy │
│(Process) │ │(Service  │ │(Load      │
│          │ │Discovery)│ │Balancer)  │
└──────────┘ └──────────┘ └───────────┘
     │
┌────▼──────────────────────────┐
│  Physical Servers (Processes) │
└───────────────────────────────┘
```

### 2.2 Agent 组件设计

#### 职责定义
1. **进程管理**:
   - 接收创建/停止/重启服务命令
   - 启动和管理应用进程(非容器化)
   - 监控进程状态,上报异常退出

2. **健康检查**:
   - 执行 HTTP 健康探针
   - 执行 TCP 连通性检查
   - 上报健康状态到 Registry

3. **状态上报**:
   - 定期上报心跳
   - 上报实例状态(运行/停止/异常)
   - 上报资源使用情况(CPU/内存)

#### 部署方式决策

**✅ 最终方案: Supervisor 进程管理**

**实现方式**:
- 每台物理机部署 Agent 程序
- 使用 Supervisor 或 systemd 实现进程守护
- Agent 以 **root 权限**运行

**Supervisor 配置示例**:
```ini
[program:x-mindflow-agent]
command=/usr/local/bin/x-mindflow-agent --config=/etc/x-mindflow/agent.conf
directory=/var/lib/x-mindflow
autostart=true
autorestart=true
startretries=3
user=root
redirect_stderr=true
stdout_logfile=/var/log/x-mindflow/agent.log
```

**systemd 配置示例**:
```ini
[Unit]
Description=x-mindflow Agent
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/x-mindflow-agent --config=/etc/x-mindflow/agent.conf
Restart=always
RestartSec=10s

[Install]
WantedBy=multi-user.target
```

#### 权限说明

Agent 需要 root 权限以执行以下操作:
- 启动/停止/重启应用进程
- 绑定特权端口(< 1024)
- 监控系统资源
- 执行健康检查脚本

#### Agent API 接口

```go
// Agent HTTP API
type AgentAPI interface {
    // 启动服务实例
    StartService(req StartServiceRequest) error
    
    // 停止服务实例
    StopService(instanceID string) error
    
    // 重启服务实例
    RestartService(instanceID string) error
    
    // 查询实例状态
    GetInstanceStatus(instanceID string) (InstanceStatus, error)
    
    // 健康检查
    HealthCheck() (HealthStatus, error)
}

type StartServiceRequest struct {
    InstanceID   string            // 实例唯一标识
    ServiceName  string            // 服务名称
    Version      string            // 版本号
    Binary       string            // 可执行文件路径
    Args         []string          // 启动参数
    Env          map[string]string // 环境变量
    WorkDir      string            // 工作目录
    Port         int               // 服务端口
    HealthCheck  HealthCheckConfig // 健康检查配置
}
```

### 2.3 Registry 注册中心设计

#### 职责定义
1. **服务注册**:
   - Agent 启动实例后注册到 Registry
   - 存储实例地址、端口、健康状态

2. **服务发现**:
   - 提供查询接口,获取服务所有实例
   - 按版本、状态等条件过滤实例

3. **健康状态管理**:
   - 接收 Agent 上报的健康状态
   - 标记不健康实例,从可用列表移除

4. **动态更新 Edge Proxy**:
   - 实例变化时通知 Edge Proxy
   - 触发 Proxy 配置热更新

#### 数据模型

```go
type ServiceInstance struct {
    InstanceID   string    // 实例唯一标识
    ServiceName  string    // 服务名称
    Version      string    // 版本号
    IPAddress    string    // IP 地址
    Port         int       // 端口
    Status       string    // 状态: running/stopped/unhealthy
    LastHeartbeat time.Time // 最后心跳时间
    Metadata     map[string]string // 元数据
}
```

#### Registry API 接口

```go
type RegistryAPI interface {
    // 注册实例
    RegisterInstance(instance ServiceInstance) error
    
    // 注销实例
    DeregisterInstance(instanceID string) error
    
    // 更新实例健康状态
    UpdateHealth(instanceID string, status HealthStatus) error
    
    // 心跳
    Heartbeat(instanceID string) error
    
    // 查询服务所有实例
    GetServiceInstances(serviceName string) ([]ServiceInstance, error)
    
    // 查询指定版本实例
    GetInstancesByVersion(serviceName, version string) ([]ServiceInstance, error)
    
    // 查询健康实例
    GetHealthyInstances(serviceName string) ([]ServiceInstance, error)
}
```

### 2.4 Edge Proxy 流量控制设计

#### 职责定义
1. **统一流量入口**:
   - 接收所有外部请求
   - 根据服务名路由到对应实例

2. **动态权重调整**:
   - 接收逻辑层下发的权重配置
   - 按权重分配流量到不同版本

3. **负载均衡**:
   - 支持轮询(Round Robin)
   - 支持加权轮询(Weighted Round Robin)
   - 支持最少连接(Least Connections)

4. **配置热更新**:
   - 接收 Registry 推送的实例变更
   - 动态更新路由表,无需重启

#### 技术选型

**最终决策**（基于 #12 讨论）:

**MVP 阶段（2天）**: **Nginx + 动态配置生成**
- 通过 Go 后端 API 自动生成 Nginx 配置文件
- 自动触发 `nginx -s reload` 实现配置热更新
- 完全自动化，无需人工介入
- 支持基于模板的权重调整

**长期方案**: **Envoy + xDS API**
- 真正的动态配置能力（无需 reload）
- 零中断流量切换
- 强大的可观测性
- 与 K8s 生态深度集成

**技术实现**:
```go
// 示例：Nginx 配置自动生成
func UpdateTrafficWeight(v1Weight, v2Weight int) error {
    // 1. 基于模板生成配置
    tmpl := template.Must(template.New("nginx").Parse(nginxTemplate))
    f, _ := os.Create("/etc/nginx/conf.d/lb.conf")
    defer f.Close()
    tmpl.Execute(f, map[string]int{"V1Weight": v1Weight, "V2Weight": v2Weight})
    
    // 2. 自动 reload
    return exec.Command("nginx", "-s", "reload").Run()
}
```

**备选方案评估**:
- **HAProxy**: 性能优秀但动态配置支持较弱
- **Traefik**: 适合 K8s 但 BareMetal 场景支持有限
- **自研轻量 Proxy**: 开发成本高，不符合"最少功能"原则

#### Edge Proxy API 接口

```go
type EdgeProxyAPI interface {
    // 更新流量权重
    UpdateTrafficWeights(serviceName string, weights map[string]float64) error
    
    // 更新实例列表
    UpdateInstances(serviceName string, instances []ServiceInstance) error
    
    // 查询当前配置
    GetCurrentConfig(serviceName string) (ProxyConfig, error)
    
    // 健康检查
    HealthCheck() (HealthStatus, error)
}

type ProxyConfig struct {
    ServiceName string
    Weights     map[string]float64 // version -> weight
    Instances   []ServiceInstance
    LBAlgorithm string // 负载均衡算法
}
```

### 2.5 进程部署流程

```
1. 逻辑层调用 BareMetal Adapter CreateInstances()
         ↓
2. Adapter 选择目标物理机,调用 Agent StartService()
         ↓
3. Agent 启动应用进程,注册到 Supervisor
         ↓
4. Agent 向 Registry 注册实例信息
         ↓
5. Registry 通知 Edge Proxy 更新实例列表
         ↓
6. Edge Proxy 热更新配置,开始接收流量
         ↓
7. Agent 定期上报心跳和健康状态
```

### 2.6 容器运行时决策

**❌ 不使用容器运行时**

**决策理由**:
- 简化部署架构,避免容器层开销
- 更直接的资源控制和进程管理
- 降低运维复杂度

**架构对比**:
```
K8s 环境:     应用 → Container → Pod → Node
物理机环境:   应用 → Process → Physical Server ✓
```

### 2.7 限制说明

**⚠️ BareMetal 环境不支持的功能**:

1. **StatefulSet 有状态服务**:
   - BareMetal 环境仅支持无状态服务部署
   - 有状态服务建议部署到 K8s 环境

2. **自动扩缩容**:
   - MVP 版本不支持自动扩缩容
   - 需手动调整实例数量

3. **资源隔离**:
   - 进程级部署缺乏容器级别的资源隔离
   - 依赖操作系统调度和资源限制

---

## 三、进一步讨论的问题

### 3.1 进程资源限制

**问题**: 物理机环境如何限制单个服务的 CPU/内存使用?

**候选方案**:
- **使用 cgroups**: Linux 原生资源限制机制
- **仅依赖 OS 调度**: 依赖操作系统自动管理

**建议**: 使用 cgroups 实现基础资源限制,避免单个服务占用过多资源。

### 3.2 日志收集

**问题**: Agent 和应用进程产生的日志如何统一收集?

**候选方案**:
- **本地文件 + 日志采集器**: 应用输出到本地文件,使用 Filebeat/Fluentd 采集
- **直接上报**: 应用直接将日志推送到中心日志系统

**建议**: 本地文件 + 日志采集器,保证日志可靠性。

### 3.3 配置管理

**问题**: 服务的配置文件如何分发和更新?

**候选方案**:
- **Agent 拉取**: Agent 从 Registry 或配置中心拉取配置
- **推送模式**: 发布系统主动推送配置到 Agent

**建议**: Agent 定期拉取 + 推送触发更新,结合两种模式优势。

---

## 四、后续行动项

基于本次讨论结果,建议拆分以下 sub-issues:

1. **[模块设计] BareMetal Agent 详细设计**
   - Agent 进程管理机制
   - Supervisor 配置模板
   - 健康检查实现方案
   - 进程启动/停止协议

2. **[模块设计] Registry 注册中心设计**
   - 服务注册/注销接口
   - 健康状态存储方案
   - 服务发现 API 定义

3. **[模块设计] Edge Proxy 流量控制设计**
   - 动态权重更新机制
   - 负载均衡算法选择
   - 配置热更新方案

4. **[文档] 环境能力边界说明文档**
   - K8s vs BareMetal 功能对比表
   - StatefulSet 限制说明
   - 最佳实践建议

---

## 五、设计原则

- ✅ 最少组件支撑逻辑层操作
- ✅ 不追求实现"完整 K8s-like 系统"
- ✅ 复用现有成熟组件
- ✅ 进程级部署,简化架构
- ✅ root 权限运行,确保功能完整性

---

## 相关 Issue

- 主 Issue: [#3 智能发布系统](https://github.com/LiusCraft/x-mindflow/issues/3)
- 统一发布抽象层: [#7 统一发布抽象层设计](https://github.com/LiusCraft/x-mindflow/issues/7)
- 本 Issue: [#8 环境适配层设计](https://github.com/LiusCraft/x-mindflow/issues/8)
