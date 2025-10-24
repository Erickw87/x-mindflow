# 拓扑可视化与发布管理架构设计文档

**关联 Issue**: #22  
**文档类型**: 架构设计  
**创建日期**: 2025-10-24  
**状态**: 设计阶段

---

## 📋 概述

本文档描述了拓扑可视化与发布管理模块的系统架构设计,包括系统分层、组件交互、数据流、技术选型等核心内容。

### 设计目标

1. **高可用性**: 系统核心组件支持高可用部署
2. **可扩展性**: 支持水平扩展,应对大规模服务拓扑
3. **松耦合**: 前后端分离,模块间低耦合
4. **易维护性**: 清晰的分层架构,便于理解和维护
5. **高性能**: 合理使用缓存,优化数据库查询

---

## 🏗️ 系统架构图

### 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                        用户层 (Users)                             │
│                     Web Browser / CLI Tool                       │
└───────────────────────────┬─────────────────────────────────────┘
                            │ HTTPS
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      前端层 (Frontend)                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ Vue 3 + TS   │  │    Pinia     │  │   AntV G6    │          │
│  │   Components │  │ State Manager│  │ Graph Engine │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│           │               │                   │                  │
│           └───────────────┴───────────────────┘                  │
│                           │ Axios HTTP Client                    │
└───────────────────────────┬─────────────────────────────────────┘
                            │ RESTful API (JSON)
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                   API 网关层 (API Gateway)                        │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Gin HTTP Server                                          │   │
│  │  - 路由管理  - 认证鉴权  - 限流熔断  - 日志记录          │   │
│  └──────────────────────────────────────────────────────────┘   │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    业务逻辑层 (Business Layer)                    │
│                                                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │ Topology Service│  │Deployment Service│  │ Instance Service│ │
│  │  拓扑管理服务    │  │  发布管理服务    │  │  实例管理服务    │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│           │                    │                     │           │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   LB Converter  │  │ Traffic Monitor │  │  Auth Service   │ │
│  │  LB 配置转换器   │  │   流量监控器    │  │   认证服务      │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                   数据访问层 (Data Access Layer)                  │
│                                                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │Topology Repo    │  │Deployment Repo  │  │ Instance Repo   │ │
│  │  拓扑仓储       │  │   发布仓储      │  │   实例仓储      │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
│           │                    │                     │           │
│           └────────────────────┴─────────────────────┘           │
│                              GORM ORM                            │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      数据层 (Data Layer)                          │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              PostgreSQL 18.0 Database                    │    │
│  │  - topologies  - topology_nodes  - topology_edges       │    │
│  │  - deployments - deployment_steps - service_instances   │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                 freecache (内存缓存)                      │    │
│  │  - 拓扑数据缓存  - 实例状态缓存  - 用户会话缓存          │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    外部系统 (External Systems)                    │
│                                                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │
│  │   K8s Cluster   │  │  Load Balancer  │  │ Monitoring      │ │
│  │  容器编排平台    │  │   负载均衡器    │  │   监控系统      │ │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🔄 系统分层架构

### 1. 前端层 (Frontend Layer)

**职责**: 用户交互、拓扑可视化、状态管理

**核心组件**:
- **Vue 3 组件**: 页面和业务组件
- **Pinia Store**: 全局状态管理
- **AntV G6**: 图可视化引擎
- **Axios**: HTTP 客户端

**关键特性**:
- 响应式 UI 设计
- 实时拓扑图渲染
- 本地状态缓存
- WebSocket 支持(未来)

---

### 2. API 网关层 (API Gateway Layer)

**职责**: 请求路由、认证鉴权、流量控制、日志记录

**核心组件**:
- **Gin HTTP Server**: Web 框架
- **JWT Middleware**: 认证中间件
- **Rate Limiter**: 限流中间件
- **Logger Middleware**: 日志中间件

**关键流程**:

```
HTTP 请求 → 日志记录 → 认证鉴权 → 限流检查 → 路由分发 → 业务处理
                ↓
           响应返回 ← 错误处理 ← 数据序列化 ← 业务结果
```

**伪代码示例**:

```go
func SetupRouter(r *gin.Engine) {
    // 全局中间件
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    r.Use(middleware.CORS())
    
    // API 版本组
    v1 := r.Group("/api/v1")
    
    // 公开接口
    v1.POST("/auth/login", handler.Login)
    
    // 需要认证的接口
    authorized := v1.Group("")
    authorized.Use(middleware.JWTAuth())
    authorized.Use(middleware.RateLimit(100, time.Minute))
    {
        // 拓扑管理
        topologies := authorized.Group("/topologies")
        topologies.GET("", handler.ListTopologies)
        topologies.POST("", handler.CreateTopology)
        
        // 发布管理
        deployments := authorized.Group("/deployments")
        deployments.GET("", handler.ListDeployments)
        deployments.POST("", handler.CreateDeployment)
    }
}
```

---

### 3. 业务逻辑层 (Business Layer)

**职责**: 核心业务逻辑、数据处理、规则校验

#### 3.1 Topology Service (拓扑管理服务)

**职责**:
- 拓扑 CRUD 操作
- 节点和边的管理
- 拓扑克隆逻辑
- 拓扑版本管理

**核心方法**:

```go
type TopologyService interface {
    // 拓扑管理
    CreateTopology(ctx context.Context, req *CreateTopologyRequest) (*Topology, error)
    GetTopology(ctx context.Context, id string) (*TopologyDetail, error)
    ListTopologies(ctx context.Context, filter *TopologyFilter) (*TopologyList, error)
    UpdateTopology(ctx context.Context, id string, req *UpdateTopologyRequest) error
    DeleteTopology(ctx context.Context, id string) error
    
    // 节点管理
    AddNode(ctx context.Context, topologyID string, node *NodeRequest) (*Node, error)
    UpdateNode(ctx context.Context, topologyID, nodeID string, req *UpdateNodeRequest) error
    DeleteNode(ctx context.Context, topologyID, nodeID string) error
    
    // 边管理
    AddEdge(ctx context.Context, topologyID string, edge *EdgeRequest) (*Edge, error)
    UpdateEdge(ctx context.Context, topologyID, edgeID string, req *UpdateEdgeRequest) error
    DeleteEdge(ctx context.Context, topologyID, edgeID string) error
    
    // 配置转换
    ExportToLBConfig(ctx context.Context, topologyID string) (*LBConfig, error)
    ImportFromLBConfig(ctx context.Context, config *LBConfig) (*Topology, error)
}
```

---

#### 3.2 Deployment Service (发布管理服务)

**职责**:
- 发布任务管理
- 发布流程编排
- 实例和流量控制
- 回滚操作

**核心方法**:

```go
type DeploymentService interface {
    // 发布任务管理
    CreateDeployment(ctx context.Context, req *CreateDeploymentRequest) (*Deployment, error)
    GetDeployment(ctx context.Context, id string) (*DeploymentDetail, error)
    ListDeployments(ctx context.Context, filter *DeploymentFilter) (*DeploymentList, error)
    
    // 发布控制
    StartDeployment(ctx context.Context, id string) error
    PauseDeployment(ctx context.Context, id string) error
    ContinueDeployment(ctx context.Context, id string) error
    RollbackDeployment(ctx context.Context, id string, reason string) error
    CancelDeployment(ctx context.Context, id string) error
    
    // 步骤执行
    ExecuteStep(ctx context.Context, deploymentID string, stepNumber int) error
    ValidateStep(ctx context.Context, deploymentID string, stepNumber int) error
}
```

**发布流程状态机**:

```
         ┌─────────┐
    ┌────│ pending │────┐
    │    └─────────┘    │
    │                   │
    │ start()           │ cancel()
    │                   │
    ▼                   ▼
┌─────────┐       ┌───────────┐
│ running │◄──────│ cancelled │
└─────────┘       └───────────┘
    │   │
    │   │ pause()
    │   │
    │   ▼
    │ ┌────────┐  continue()
    │ │ paused │──────────┘
    │ └────────┘
    │
    │ complete() / fail() / rollback()
    │
    ▼
┌───────────┐   ┌────────┐   ┌──────────────┐
│ completed │   │ failed │   │ rolled_back  │
└───────────┘   └────────┘   └──────────────┘
```

---

#### 3.3 Instance Service (实例管理服务)

**职责**:
- 服务实例监控
- 实例健康检查
- 实例生命周期管理

**核心方法**:

```go
type InstanceService interface {
    // 实例查询
    GetInstance(ctx context.Context, id string) (*Instance, error)
    ListInstances(ctx context.Context, filter *InstanceFilter) (*InstanceList, error)
    
    // 实例控制
    CreateInstances(ctx context.Context, serviceName, version string, count int) error
    TerminateInstances(ctx context.Context, instanceIDs []string) error
    
    // 健康检查
    CheckHealth(ctx context.Context, instanceID string) (*HealthStatus, error)
    UpdateHealthStatus(ctx context.Context, instanceID string, status *HealthStatus) error
}
```

---

#### 3.4 LB Converter (LB 配置转换器)

**职责**: 拓扑图与 LB 配置的双向转换

**转换逻辑**:

```go
// 拓扑 → LB 配置
func TopologyToLBConfig(topology *TopologyDetail) (*LBConfig, error) {
    lbConfig := &LBConfig{
        ServiceName: topology.ServiceName,
        Rules:       []LBRule{},
    }
    
    // 遍历 LB 节点
    for _, node := range topology.Nodes {
        if node.Type == "lb" {
            // 提取 LB 规则
            for _, rule := range node.LBRules {
                lbConfig.Rules = append(lbConfig.Rules, rule)
            }
        }
    }
    
    return lbConfig, nil
}

// LB 配置 → 拓扑
func LBConfigToTopology(config *LBConfig) (*TopologyDetail, error) {
    topology := &TopologyDetail{
        ServiceName: config.ServiceName,
        Nodes:       []Node{},
        Edges:       []Edge{},
    }
    
    // 根据 LB 规则构建节点和边
    for _, rule := range config.Rules {
        // 创建 LB 节点
        lbNode := createLBNode(rule)
        topology.Nodes = append(topology.Nodes, lbNode)
        
        // 创建目标服务节点
        for _, target := range rule.Targets {
            serviceNode := createServiceNode(target)
            topology.Nodes = append(topology.Nodes, serviceNode)
            
            // 创建连接边
            edge := createEdge(lbNode.ID, serviceNode.ID, target.Weight)
            topology.Edges = append(topology.Edges, edge)
        }
    }
    
    return topology, nil
}
```

---

#### 3.5 Traffic Monitor (流量监控器)

**职责**: 实时监控流量分布和实例健康状态

**核心方法**:

```go
type TrafficMonitor interface {
    // 流量统计
    GetTrafficStats(ctx context.Context, topologyID string, timeRange string) (*TrafficStats, error)
    GetNodeTraffic(ctx context.Context, nodeID string) (*NodeTrafficStats, error)
    
    // 实时监控
    MonitorDeployment(ctx context.Context, deploymentID string) (<-chan *DeploymentStatus, error)
}
```

---

### 4. 数据访问层 (Data Access Layer)

**职责**: 数据持久化、缓存管理、数据库操作

**Repository 模式**:

```go
// 拓扑仓储接口
type TopologyRepository interface {
    Create(ctx context.Context, topology *model.Topology) error
    FindByID(ctx context.Context, id string) (*model.Topology, error)
    FindAll(ctx context.Context, filter *TopologyFilter) ([]*model.Topology, error)
    Update(ctx context.Context, topology *model.Topology) error
    Delete(ctx context.Context, id string) error
    
    // 带缓存的查询
    FindByIDWithCache(ctx context.Context, id string) (*model.Topology, error)
}

// 实现示例
type topologyRepositoryImpl struct {
    db    *gorm.DB
    cache *freecache.Cache
}

func (r *topologyRepositoryImpl) FindByIDWithCache(ctx context.Context, id string) (*model.Topology, error) {
    // 1. 尝试从缓存获取
    cacheKey := fmt.Sprintf("topology:%s", id)
    if data, err := r.cache.Get([]byte(cacheKey)); err == nil {
        var topology model.Topology
        if err := json.Unmarshal(data, &topology); err == nil {
            return &topology, nil
        }
    }
    
    // 2. 缓存未命中,从数据库查询
    var topology model.Topology
    if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&topology).Error; err != nil {
        return nil, err
    }
    
    // 3. 写入缓存
    if data, err := json.Marshal(topology); err == nil {
        r.cache.Set([]byte(cacheKey), data, 300) // 缓存 5 分钟
    }
    
    return &topology, nil
}
```

---

### 5. 数据层 (Data Layer)

#### 5.1 PostgreSQL 数据库

**特性**:
- 支持 JSONB 类型(存储灵活配置)
- 支持事务(保证数据一致性)
- 丰富的索引类型(B-Tree, GIN, GiST)
- 支持分区表(大数据量优化)

**连接池配置**:

```go
func InitDB(config *DBConfig) (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        config.Host, config.Port, config.User, config.Password, config.DBName)
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, err
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    
    // 连接池配置
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    return db, nil
}
```

---

#### 5.2 freecache 内存缓存

**用途**:
- 热点数据缓存(拓扑、实例状态)
- 减少数据库查询压力
- 提升响应速度

**缓存策略**:

| 数据类型 | TTL | 淘汰策略 |
|---------|-----|---------|
| 拓扑详情 | 5 分钟 | LRU |
| 实例状态 | 30 秒 | LRU |
| 用户会话 | 1 小时 | LRU |

**初始化配置**:

```go
import "github.com/coocood/freecache"

func InitCache() *freecache.Cache {
    // 分配 100MB 内存
    cache := freecache.NewCache(100 * 1024 * 1024)
    return cache
}
```

---

## 📈 数据流图

### 拓扑查询流程

```
用户请求 → Frontend (Vue)
             ↓
          Axios GET /api/v1/topologies/:id
             ↓
          API Gateway (Gin)
             ↓ JWT 验证
          Topology Handler
             ↓
          Topology Service
             ↓
          Topology Repository
             ↓
       检查 freecache 缓存
             ↓ (miss)
          PostgreSQL 查询
             ↓
          GORM ORM 映射
             ↓
          写入缓存 + 返回结果
             ↓
          JSON 序列化
             ↓
          HTTP 响应 → Frontend
             ↓
          Pinia Store 更新
             ↓
          AntV G6 渲染拓扑图
```

---

### 发布流程数据流

```
用户创建发布任务
     ↓
Frontend 提交 POST /api/v1/deployments
     ↓
Deployment Service 创建任务(状态: pending)
     ↓
保存到 deployments 表
     ↓
用户点击"启动发布"
     ↓
Deployment Service.StartDeployment()
     ↓
更新任务状态为 running
     ↓
执行第 1 步:
  ├─ 计算实例数: 新 1 个, 旧 9 个
  ├─ 调用 Instance Service 创建新实例
  ├─ 调用 K8s API 创建 Pod
  ├─ 更新 LB 配置(流量权重 10%:90%)
  └─ 等待手动确认或自动暂停
     ↓
用户确认 / 超时自动继续
     ↓
执行第 2 步:
  ├─ 计算实例数: 新 3 个, 旧 7 个
  ├─ 调用 Instance Service 扩容新实例
  ├─ 调用 Instance Service 缩容旧实例
  ├─ 更新 LB 配置(流量权重 30%:70%)
  └─ 等待 5 分钟
     ↓
...继续执行后续步骤...
     ↓
所有步骤完成
     ↓
更新任务状态为 completed
     ↓
发送完成通知
```

---

## 🔌 外部系统集成

### 1. Kubernetes 集成

**集成方式**: 通过 K8s Client-Go SDK

**核心操作**:

```go
import "k8s.io/client-go/kubernetes"

type K8sAdapter struct {
    clientset *kubernetes.Clientset
}

// 创建 Deployment
func (k *K8sAdapter) CreateDeployment(ctx context.Context, namespace, name, image string, replicas int32) error {
    deployment := &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name: name,
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{"app": name},
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{"app": name},
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  name,
                            Image: image,
                        },
                    },
                },
            },
        },
    }
    
    _, err := k.clientset.AppsV1().Deployments(namespace).Create(ctx, deployment, metav1.CreateOptions{})
    return err
}

// 扩缩容
func (k *K8sAdapter) ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error {
    scale := &autoscalingv1.Scale{
        Spec: autoscalingv1.ScaleSpec{
            Replicas: replicas,
        },
    }
    
    _, err := k.clientset.AppsV1().Deployments(namespace).UpdateScale(ctx, name, scale, metav1.UpdateOptions{})
    return err
}
```

---

### 2. Load Balancer 集成

**集成方式**: 通过 HTTP API 或配置文件

**LB 配置更新流程**:

```go
type LBAdapter interface {
    UpdateRules(ctx context.Context, serviceName string, rules []LBRule) error
    GetCurrentRules(ctx context.Context, serviceName string) ([]LBRule, error)
}

// Nginx 实现示例
type NginxAdapter struct {
    configPath string
}

func (n *NginxAdapter) UpdateRules(ctx context.Context, serviceName string, rules []LBRule) error {
    // 1. 生成 Nginx 配置文件
    config := generateNginxConfig(serviceName, rules)
    
    // 2. 写入配置文件
    if err := ioutil.WriteFile(n.configPath, []byte(config), 0644); err != nil {
        return err
    }
    
    // 3. 重载 Nginx
    cmd := exec.Command("nginx", "-s", "reload")
    return cmd.Run()
}

func generateNginxConfig(serviceName string, rules []LBRule) string {
    config := fmt.Sprintf("upstream %s {\n", serviceName)
    
    for _, rule := range rules {
        for _, target := range rule.Targets {
            config += fmt.Sprintf("    server %s weight=%d;\n", target.Address, target.Weight)
        }
    }
    
    config += "}\n"
    return config
}
```

---

### 3. 监控系统集成

**集成方式**: Prometheus + Grafana

**指标暴露**:

```go
import "github.com/prometheus/client_golang/prometheus"

var (
    deploymentDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "deployment_duration_seconds",
            Help: "Deployment duration in seconds",
        },
        []string{"service", "version", "status"},
    )
    
    instanceCount = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "service_instance_count",
            Help: "Number of service instances",
        },
        []string{"service", "version"},
    )
)

func init() {
    prometheus.MustRegister(deploymentDuration)
    prometheus.MustRegister(instanceCount)
}
```

---

## 🔒 安全架构

### 认证与授权

**JWT Token 认证流程**:

```
1. 用户登录 → 验证用户名密码
2. 生成 JWT Token(包含用户 ID, 角色, 过期时间)
3. 返回 Token 给客户端
4. 客户端在请求头携带 Token: Authorization: Bearer <token>
5. API 网关验证 Token 有效性
6. 解析 Token 获取用户信息
7. 基于角色进行权限校验
8. 放行或拒绝请求
```

**Token 结构**:

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "user_id": "uuid-1234",
    "username": "admin",
    "role": "admin",
    "exp": 1735123456
  },
  "signature": "..."
}
```

**权限控制**:

| 角色 | 权限 |
|------|------|
| admin | 所有操作(CRUD 拓扑、发布、实例) |
| user | 只读(查看拓扑、发布状态) |

---

### 数据安全

1. **密码加密**: 使用 bcrypt 加密存储
2. **SQL 注入防护**: 使用 GORM 参数化查询
3. **XSS 防护**: 前端输入验证 + 后端转义
4. **HTTPS**: 生产环境强制 HTTPS
5. **敏感信息**: 使用环境变量,不硬编码

---

## 🚀 性能优化策略

### 1. 缓存策略

- **L1 缓存**: freecache 内存缓存(热点数据)
- **L2 缓存**: Redis(分布式缓存,未来扩展)
- **缓存失效**: TTL + 主动失效

### 2. 数据库优化

- **索引优化**: 为常用查询字段创建索引
- **连接池**: 合理配置连接池大小
- **慢查询分析**: 定期分析慢查询日志
- **读写分离**: 主从复制(高级特性)

### 3. API 优化

- **分页**: 大列表强制分页
- **字段过滤**: 支持 `fields` 参数减少数据传输
- **批量操作**: 提供批量 API 减少请求次数
- **压缩**: 启用 Gzip 压缩

### 4. 前端优化

- **懒加载**: 路由和组件懒加载
- **虚拟滚动**: 大列表使用虚拟滚动
- **请求去重**: 相同请求去重
- **本地缓存**: Pinia 持久化

---

## 📊 可观测性

### 日志

**日志级别**: DEBUG < INFO < WARN < ERROR < FATAL

**日志格式**(JSON):

```json
{
  "timestamp": "2025-10-24T12:00:00Z",
  "level": "INFO",
  "service": "topology-service",
  "trace_id": "abc123",
  "user_id": "user-001",
  "message": "Topology created successfully",
  "data": {
    "topology_id": "topo-001",
    "service_name": "user-service"
  }
}
```

**日志收集**: uber-go/zap → 文件 → Filebeat → Elasticsearch → Kibana

---

### 监控

**监控指标**:

- **系统指标**: CPU、内存、磁盘、网络
- **应用指标**: QPS、响应时间、错误率
- **业务指标**: 发布成功率、实例健康率、拓扑数量

**监控工具**: Prometheus + Grafana

---

### 链路追踪

**工具**: OpenTelemetry (未来扩展)

**追踪流程**:

```
请求 → API Gateway(生成 trace_id) → Service A → Service B → 数据库
             ↓                          ↓           ↓
          记录 span                 记录 span    记录 span
             ↓                          ↓           ↓
                    统一上报到 Jaeger / Zipkin
```

---

## 🔧 部署架构

### 开发环境

```yaml
# docker-compose.yml
version: '3.8'
services:
  postgres:
    image: postgres:18.0
    environment:
      POSTGRES_DB: xmindflow
      POSTGRES_USER: admin
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
  
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    depends_on:
      - postgres
    environment:
      DB_HOST: postgres
      DB_PORT: 5432
  
  frontend:
    build: ./frontend
    ports:
      - "5173:5173"
    depends_on:
      - backend
```

---

### 生产环境

**部署方式**: Kubernetes

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: xmindflow-backend
spec:
  replicas: 3
  selector:
    matchLabels:
      app: xmindflow-backend
  template:
    metadata:
      labels:
        app: xmindflow-backend
    spec:
      containers:
      - name: backend
        image: xmindflow/backend:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: DB_HOST
          value: postgres-service
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
```

---

## 📚 相关文档

- **API 设计**: `docs/api/22-topology-api-design.md`
- **数据库设计**: `docs/database/22-topology-schema-design.md`
- **模块设计**: `docs/modules/22-topology-module-design.md`
- **原型文档**: `docs/prototypes/11-topology-visualization-prototype.md`

---

**文档版本**: v1.0  
**最后更新**: 2025-10-24  
**维护人**: xgopilot (Claude Code AI Assistant)
