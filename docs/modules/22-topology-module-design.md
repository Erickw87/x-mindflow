# 拓扑可视化与发布管理模块设计文档

**关联 Issue**: #22  
**文档类型**: 模块设计  
**创建日期**: 2025-10-24  
**状态**: 设计阶段

---

## 📋 概述

本文档详细定义了拓扑可视化与发布管理系统的前后端模块划分、模块职责、接口定义和模块间交互关系,为后续开发实现提供清晰的技术规范。

### 设计原则

1. **单一职责**: 每个模块只负责一个明确的功能领域
2. **高内聚低耦合**: 模块内部紧密相关,模块间松散耦合
3. **接口清晰**: 模块间通过明确定义的接口交互
4. **可测试性**: 模块设计便于单元测试和集成测试
5. **可扩展性**: 预留扩展点,支持未来功能增强

---

## 🏗️ 模块总览

### 后端模块结构

```
backend/
├── cmd/
│   └── server/                # 应用入口
│       └── main.go
├── internal/                  # 私有代码
│   ├── api/                  # API 层
│   │   ├── handler/          # HTTP 处理器
│   │   ├── middleware/       # 中间件
│   │   └── router/           # 路由配置
│   ├── service/              # 业务逻辑层
│   ├── repository/           # 数据访问层
│   ├── model/                # 数据模型
│   ├── adapter/              # 外部系统适配器
│   └── config/               # 配置管理
├── pkg/                      # 公共库
│   ├── errors/               # 错误定义
│   ├── logger/               # 日志工具
│   ├── cache/                # 缓存工具
│   └── utils/                # 工具函数
└── test/                     # 测试
```

### 前端模块结构

```
frontend/src/
├── api/                      # API 接口封装
├── components/               # 组件
│   ├── common/              # 通用组件
│   └── business/            # 业务组件
├── views/                    # 页面组件
├── stores/                   # Pinia 状态管理
├── router/                   # 路由配置
├── composables/              # 组合式函数
├── types/                    # TypeScript 类型
├── utils/                    # 工具函数
└── styles/                   # 样式
```

---

## 🔧 后端模块设计

### 1. API 层 (Handler + Middleware)

#### 1.1 Handler 模块

**职责**: 处理 HTTP 请求,参数验证,调用 Service,返回响应

**模块列表**:

| 模块名 | 文件路径 | 职责 |
|--------|---------|------|
| TopologyHandler | `internal/api/handler/topology_handler.go` | 拓扑管理相关接口 |
| NodeHandler | `internal/api/handler/node_handler.go` | 节点管理相关接口 |
| EdgeHandler | `internal/api/handler/edge_handler.go` | 连接边管理相关接口 |
| DeploymentHandler | `internal/api/handler/deployment_handler.go` | 发布任务相关接口 |
| InstanceHandler | `internal/api/handler/instance_handler.go` | 实例管理相关接口 |
| AuthHandler | `internal/api/handler/auth_handler.go` | 认证相关接口 |

**接口设计示例**: `TopologyHandler`

```go
// internal/api/handler/topology_handler.go
package handler

import (
    "github.com/gin-gonic/gin"
    "x-mindflow/internal/service"
)

type TopologyHandler struct {
    topologyService service.TopologyService
}

func NewTopologyHandler(topologyService service.TopologyService) *TopologyHandler {
    return &TopologyHandler{
        topologyService: topologyService,
    }
}

// ListTopologies 获取拓扑列表
func (h *TopologyHandler) ListTopologies(c *gin.Context) {
    // 1. 解析查询参数
    var req ListTopologiesRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        c.JSON(400, gin.H{"code": 1001, "message": "参数错误", "details": err.Error()})
        return
    }
    
    // 2. 调用 Service
    result, err := h.topologyService.ListTopologies(c.Request.Context(), &req)
    if err != nil {
        c.JSON(500, gin.H{"code": 1000, "message": "查询失败", "details": err.Error()})
        return
    }
    
    // 3. 返回响应
    c.JSON(200, gin.H{"code": 200, "message": "success", "data": result})
}

// CreateTopology 创建拓扑
func (h *TopologyHandler) CreateTopology(c *gin.Context) {
    var req CreateTopologyRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"code": 1001, "message": "参数错误", "details": err.Error()})
        return
    }
    
    // 获取当前用户
    userID := c.GetString("user_id")
    
    topology, err := h.topologyService.CreateTopology(c.Request.Context(), &req, userID)
    if err != nil {
        c.JSON(500, gin.H{"code": 1000, "message": "创建失败", "details": err.Error()})
        return
    }
    
    c.JSON(201, gin.H{"code": 201, "message": "创建成功", "data": topology})
}

// GetTopology 获取拓扑详情
func (h *TopologyHandler) GetTopology(c *gin.Context) {
    id := c.Param("id")
    
    topology, err := h.topologyService.GetTopologyDetail(c.Request.Context(), id)
    if err != nil {
        c.JSON(404, gin.H{"code": 2001, "message": "拓扑不存在", "details": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"code": 200, "message": "success", "data": topology})
}

// UpdateTopology 更新拓扑
func (h *TopologyHandler) UpdateTopology(c *gin.Context) {
    id := c.Param("id")
    var req UpdateTopologyRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"code": 1001, "message": "参数错误", "details": err.Error()})
        return
    }
    
    if err := h.topologyService.UpdateTopology(c.Request.Context(), id, &req); err != nil {
        c.JSON(500, gin.H{"code": 1000, "message": "更新失败", "details": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"code": 200, "message": "更新成功"})
}

// DeleteTopology 删除拓扑
func (h *TopologyHandler) DeleteTopology(c *gin.Context) {
    id := c.Param("id")
    
    if err := h.topologyService.DeleteTopology(c.Request.Context(), id); err != nil {
        c.JSON(500, gin.H{"code": 1000, "message": "删除失败", "details": err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"code": 200, "message": "删除成功"})
}
```

---

#### 1.2 Middleware 模块

**职责**: 请求拦截、认证鉴权、日志记录、错误处理

**模块列表**:

| 模块名 | 文件路径 | 职责 |
|--------|---------|------|
| AuthMiddleware | `internal/api/middleware/auth.go` | JWT 认证 |
| LoggerMiddleware | `internal/api/middleware/logger.go` | 请求日志记录 |
| RecoveryMiddleware | `internal/api/middleware/recovery.go` | Panic 恢复 |
| CORSMiddleware | `internal/api/middleware/cors.go` | CORS 跨域处理 |
| RateLimitMiddleware | `internal/api/middleware/ratelimit.go` | 限流 |

**接口设计示例**: `AuthMiddleware`

```go
// internal/api/middleware/auth.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
    "strings"
)

func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 从 Header 获取 Token
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(401, gin.H{"code": 1004, "message": "未授权:缺少 Token"})
            c.Abort()
            return
        }
        
        // 2. 解析 Bearer Token
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(401, gin.H{"code": 1004, "message": "未授权:Token 格式错误"})
            c.Abort()
            return
        }
        
        // 3. 验证 Token
        token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
            return []byte("your-secret-key"), nil
        })
        
        if err != nil || !token.Valid {
            c.JSON(401, gin.H{"code": 1004, "message": "未授权:Token 无效或已过期"})
            c.Abort()
            return
        }
        
        // 4. 提取用户信息
        if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set("user_id", claims["user_id"])
            c.Set("username", claims["username"])
            c.Set("role", claims["role"])
        }
        
        c.Next()
    }
}

func RoleCheck(requiredRole string) gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists || role != requiredRole {
            c.JSON(403, gin.H{"code": 1005, "message": "权限不足"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

#### 1.3 Router 模块

**职责**: 路由配置、中间件绑定、Handler 注册

```go
// internal/api/router/router.go
package router

import (
    "github.com/gin-gonic/gin"
    "x-mindflow/internal/api/handler"
    "x-mindflow/internal/api/middleware"
)

func SetupRouter(
    topologyHandler *handler.TopologyHandler,
    deploymentHandler *handler.DeploymentHandler,
    authHandler *handler.AuthHandler,
) *gin.Engine {
    r := gin.New()
    
    // 全局中间件
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())
    r.Use(middleware.CORS())
    
    // API v1
    v1 := r.Group("/api/v1")
    {
        // 认证接口(无需 Token)
        auth := v1.Group("/auth")
        {
            auth.POST("/login", authHandler.Login)
            auth.POST("/refresh", authHandler.RefreshToken)
        }
        
        // 需要认证的接口
        authorized := v1.Group("")
        authorized.Use(middleware.JWTAuth())
        {
            // 拓扑管理
            topologies := authorized.Group("/topologies")
            {
                topologies.GET("", topologyHandler.ListTopologies)
                topologies.POST("", topologyHandler.CreateTopology)
                topologies.GET("/:id", topologyHandler.GetTopology)
                topologies.PUT("/:id", topologyHandler.UpdateTopology)
                topologies.DELETE("/:id", topologyHandler.DeleteTopology)
                
                // 节点管理
                topologies.POST("/:topology_id/nodes", topologyHandler.AddNode)
                topologies.PUT("/:topology_id/nodes/:node_id", topologyHandler.UpdateNode)
                topologies.DELETE("/:topology_id/nodes/:node_id", topologyHandler.DeleteNode)
                
                // 边管理
                topologies.POST("/:topology_id/edges", topologyHandler.AddEdge)
                topologies.PUT("/:topology_id/edges/:edge_id", topologyHandler.UpdateEdge)
                topologies.DELETE("/:topology_id/edges/:edge_id", topologyHandler.DeleteEdge)
                
                // LB 配置导出
                topologies.GET("/:id/lb-config", topologyHandler.ExportLBConfig)
            }
            
            // 发布管理
            deployments := authorized.Group("/deployments")
            {
                deployments.GET("", deploymentHandler.ListDeployments)
                deployments.POST("", deploymentHandler.CreateDeployment)
                deployments.GET("/:id", deploymentHandler.GetDeployment)
                deployments.POST("/:id/start", deploymentHandler.StartDeployment)
                deployments.POST("/:id/pause", deploymentHandler.PauseDeployment)
                deployments.POST("/:id/continue", deploymentHandler.ContinueDeployment)
                deployments.POST("/:id/rollback", deploymentHandler.RollbackDeployment)
                deployments.DELETE("/:id", deploymentHandler.CancelDeployment)
            }
        }
    }
    
    return r
}
```

---

### 2. Service 层 (业务逻辑)

#### 2.1 TopologyService

**职责**: 拓扑树的 CRUD、节点和边管理、配置转换

```go
// internal/service/topology_service.go
package service

import (
    "context"
    "x-mindflow/internal/model"
    "x-mindflow/internal/repository"
)

type TopologyService interface {
    // 拓扑管理
    CreateTopology(ctx context.Context, req *CreateTopologyRequest, userID string) (*model.Topology, error)
    GetTopologyDetail(ctx context.Context, id string) (*TopologyDetailResponse, error)
    ListTopologies(ctx context.Context, filter *ListTopologiesRequest) (*TopologyListResponse, error)
    UpdateTopology(ctx context.Context, id string, req *UpdateTopologyRequest) error
    DeleteTopology(ctx context.Context, id string) error
    
    // 节点管理
    AddNode(ctx context.Context, topologyID string, req *AddNodeRequest) (*model.TopologyNode, error)
    UpdateNode(ctx context.Context, topologyID, nodeID string, req *UpdateNodeRequest) error
    DeleteNode(ctx context.Context, topologyID, nodeID string) error
    
    // 边管理
    AddEdge(ctx context.Context, topologyID string, req *AddEdgeRequest) (*model.TopologyEdge, error)
    UpdateEdge(ctx context.Context, topologyID, edgeID string, req *UpdateEdgeRequest) error
    DeleteEdge(ctx context.Context, topologyID, edgeID string) error
    
    // 配置转换
    ExportToLBConfig(ctx context.Context, topologyID string) (*LBConfigResponse, error)
}

type topologyServiceImpl struct {
    topologyRepo repository.TopologyRepository
    nodeRepo     repository.TopologyNodeRepository
    edgeRepo     repository.TopologyEdgeRepository
}

func NewTopologyService(
    topologyRepo repository.TopologyRepository,
    nodeRepo repository.TopologyNodeRepository,
    edgeRepo repository.TopologyEdgeRepository,
) TopologyService {
    return &topologyServiceImpl{
        topologyRepo: topologyRepo,
        nodeRepo:     nodeRepo,
        edgeRepo:     edgeRepo,
    }
}

func (s *topologyServiceImpl) CreateTopology(ctx context.Context, req *CreateTopologyRequest, userID string) (*model.Topology, error) {
    // 1. 参数校验
    if req.Name == "" || req.ServiceName == "" {
        return nil, errors.New("name and service_name are required")
    }
    
    // 2. 检查名称是否已存在
    existing, _ := s.topologyRepo.FindByName(ctx, req.Name)
    if existing != nil {
        return nil, errors.New("topology name already exists")
    }
    
    // 3. 创建拓扑对象
    topology := &model.Topology{
        Name:        req.Name,
        Description: req.Description,
        ServiceName: req.ServiceName,
        Status:      "active",
        Version:     1,
        CreatedBy:   userID,
    }
    
    // 4. 如果是克隆,复制节点和边
    if req.ClonedFrom != "" {
        if err := s.cloneTopology(ctx, req.ClonedFrom, topology); err != nil {
            return nil, err
        }
    }
    
    // 5. 保存到数据库
    if err := s.topologyRepo.Create(ctx, topology); err != nil {
        return nil, err
    }
    
    return topology, nil
}

func (s *topologyServiceImpl) cloneTopology(ctx context.Context, sourceID string, target *model.Topology) error {
    // 1. 获取源拓扑
    source, err := s.topologyRepo.FindByID(ctx, sourceID)
    if err != nil {
        return err
    }
    
    // 2. 克隆节点
    nodes, err := s.nodeRepo.FindByTopologyID(ctx, sourceID)
    if err != nil {
        return err
    }
    
    nodeIDMap := make(map[string]string) // 旧 ID → 新 ID
    for _, node := range nodes {
        newNode := &model.TopologyNode{
            TopologyID:             target.ID,
            NodeType:               node.NodeType,
            Name:                   node.Name,
            Branch:                 node.Branch,
            TotalTrafficPercentage: node.TotalTrafficPercentage,
            TotalInstances:         node.TotalInstances,
            LBRules:                node.LBRules,
        }
        if err := s.nodeRepo.Create(ctx, newNode); err != nil {
            return err
        }
        nodeIDMap[node.ID] = newNode.ID
    }
    
    // 3. 克隆边
    edges, err := s.edgeRepo.FindByTopologyID(ctx, sourceID)
    if err != nil {
        return err
    }
    
    for _, edge := range edges {
        newEdge := &model.TopologyEdge{
            TopologyID:        target.ID,
            SourceNodeID:      nodeIDMap[edge.SourceNodeID],
            TargetNodeID:      nodeIDMap[edge.TargetNodeID],
            TrafficPercentage: edge.TrafficPercentage,
            Label:             edge.Label,
            Branch:            edge.Branch,
        }
        if err := s.edgeRepo.Create(ctx, newEdge); err != nil {
            return err
        }
    }
    
    target.ClonedFrom = &sourceID
    return nil
}
```

---

#### 2.2 DeploymentService

**职责**: 发布任务管理、流程编排、状态机控制

```go
// internal/service/deployment_service.go
package service

import (
    "context"
    "time"
    "x-mindflow/internal/model"
    "x-mindflow/internal/repository"
)

type DeploymentService interface {
    CreateDeployment(ctx context.Context, req *CreateDeploymentRequest, userID string) (*model.Deployment, error)
    GetDeployment(ctx context.Context, id string) (*DeploymentDetailResponse, error)
    ListDeployments(ctx context.Context, filter *ListDeploymentsRequest) (*DeploymentListResponse, error)
    
    StartDeployment(ctx context.Context, id string) error
    PauseDeployment(ctx context.Context, id string) error
    ContinueDeployment(ctx context.Context, id string) error
    RollbackDeployment(ctx context.Context, id string, reason string) error
    CancelDeployment(ctx context.Context, id string) error
    
    ExecuteStep(ctx context.Context, deploymentID string, stepNumber int) error
}

type deploymentServiceImpl struct {
    deploymentRepo repository.DeploymentRepository
    stepRepo       repository.DeploymentStepRepository
    instanceSvc    InstanceService
    lbAdapter      LBAdapter
}

func (s *deploymentServiceImpl) StartDeployment(ctx context.Context, id string) error {
    // 1. 查询发布任务
    deployment, err := s.deploymentRepo.FindByID(ctx, id)
    if err != nil {
        return err
    }
    
    // 2. 校验状态
    if deployment.Status != "pending" {
        return errors.New("只有 pending 状态的任务可以启动")
    }
    
    // 3. 更新状态
    deployment.Status = "running"
    deployment.StartedAt = time.Now()
    deployment.CurrentStep = 1
    if err := s.deploymentRepo.Update(ctx, deployment); err != nil {
        return err
    }
    
    // 4. 异步执行第一步
    go s.ExecuteStep(context.Background(), id, 1)
    
    return nil
}

func (s *deploymentServiceImpl) ExecuteStep(ctx context.Context, deploymentID string, stepNumber int) error {
    // 1. 获取步骤详情
    step, err := s.stepRepo.FindByDeploymentAndStep(ctx, deploymentID, stepNumber)
    if err != nil {
        return err
    }
    
    // 2. 更新步骤状态为 running
    step.Status = "running"
    step.StartedAt = time.Now()
    s.stepRepo.Update(ctx, step)
    
    // 3. 计算实例数
    deployment, _ := s.deploymentRepo.FindByID(ctx, deploymentID)
    totalInstances := 10 // 从配置获取
    newInstances := int(float64(totalInstances) * step.InstancesPercent / 100)
    oldInstances := totalInstances - newInstances
    
    // 4. 调整实例
    if err := s.instanceSvc.ScaleInstances(ctx, deployment.ServiceName, deployment.NewVersion, newInstances); err != nil {
        step.Status = "failed"
        step.ErrorMessage = err.Error()
        s.stepRepo.Update(ctx, step)
        return err
    }
    
    if err := s.instanceSvc.ScaleInstances(ctx, deployment.ServiceName, deployment.OldVersion, oldInstances); err != nil {
        step.Status = "failed"
        step.ErrorMessage = err.Error()
        s.stepRepo.Update(ctx, step)
        return err
    }
    
    // 5. 更新 LB 配置
    lbConfig := &LBConfig{
        ServiceName: deployment.ServiceName,
        Rules: []LBRule{
            {Weight: int(step.TrafficPercent), Version: deployment.NewVersion},
            {Weight: 100 - int(step.TrafficPercent), Version: deployment.OldVersion},
        },
    }
    if err := s.lbAdapter.UpdateRules(ctx, deployment.ServiceName, lbConfig.Rules); err != nil {
        step.Status = "failed"
        step.ErrorMessage = err.Error()
        s.stepRepo.Update(ctx, step)
        return err
    }
    
    // 6. 更新步骤状态为 completed
    step.Status = "completed"
    step.CompletedAt = time.Now()
    step.ActualInstancesNew = newInstances
    step.ActualInstancesOld = oldInstances
    s.stepRepo.Update(ctx, step)
    
    // 7. 检查是否需要暂停或继续下一步
    if step.PauseType == "manual" {
        // 等待手动确认
        deployment.Status = "paused"
        s.deploymentRepo.Update(ctx, deployment)
    } else if step.PauseType == "auto" {
        // 自动等待
        time.Sleep(time.Duration(step.PauseDuration) * time.Second)
        // 继续下一步
        s.ContinueDeployment(ctx, deploymentID)
    }
    
    return nil
}
```

---

#### 2.3 InstanceService

**职责**: 服务实例生命周期管理、健康检查

```go
// internal/service/instance_service.go
package service

type InstanceService interface {
    GetInstance(ctx context.Context, id string) (*model.ServiceInstance, error)
    ListInstances(ctx context.Context, filter *InstanceFilter) (*InstanceListResponse, error)
    
    ScaleInstances(ctx context.Context, serviceName, version string, targetCount int) error
    TerminateInstances(ctx context.Context, instanceIDs []string) error
    
    CheckHealth(ctx context.Context, instanceID string) (*HealthStatus, error)
}
```

---

### 3. Repository 层 (数据访问)

#### 3.1 TopologyRepository

**职责**: 拓扑数据的持久化操作

```go
// internal/repository/topology_repository.go
package repository

import (
    "context"
    "gorm.io/gorm"
    "x-mindflow/internal/model"
)

type TopologyRepository interface {
    Create(ctx context.Context, topology *model.Topology) error
    FindByID(ctx context.Context, id string) (*model.Topology, error)
    FindByName(ctx context.Context, name string) (*model.Topology, error)
    FindAll(ctx context.Context, filter *TopologyFilter) ([]*model.Topology, int64, error)
    Update(ctx context.Context, topology *model.Topology) error
    Delete(ctx context.Context, id string) error
}

type topologyRepositoryImpl struct {
    db *gorm.DB
}

func NewTopologyRepository(db *gorm.DB) TopologyRepository {
    return &topologyRepositoryImpl{db: db}
}

func (r *topologyRepositoryImpl) Create(ctx context.Context, topology *model.Topology) error {
    return r.db.WithContext(ctx).Create(topology).Error
}

func (r *topologyRepositoryImpl) FindByID(ctx context.Context, id string) (*model.Topology, error) {
    var topology model.Topology
    err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&topology).Error
    if err != nil {
        return nil, err
    }
    return &topology, nil
}

func (r *topologyRepositoryImpl) FindAll(ctx context.Context, filter *TopologyFilter) ([]*model.Topology, int64, error) {
    var topologies []*model.Topology
    var total int64
    
    query := r.db.WithContext(ctx).Where("deleted_at IS NULL")
    
    if filter.ServiceName != "" {
        query = query.Where("service_name = ?", filter.ServiceName)
    }
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
    
    query.Count(&total)
    
    offset := (filter.Page - 1) * filter.PageSize
    err := query.Offset(offset).Limit(filter.PageSize).Order("created_at DESC").Find(&topologies).Error
    
    return topologies, total, err
}

func (r *topologyRepositoryImpl) Update(ctx context.Context, topology *model.Topology) error {
    return r.db.WithContext(ctx).Save(topology).Error
}

func (r *topologyRepositoryImpl) Delete(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Model(&model.Topology{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}
```

---

### 4. Adapter 层 (外部系统适配)

#### 4.1 K8sAdapter

**职责**: Kubernetes 集群操作适配

```go
// internal/adapter/k8s_adapter.go
package adapter

import (
    "context"
    "k8s.io/client-go/kubernetes"
)

type K8sAdapter interface {
    CreateDeployment(ctx context.Context, namespace, name, image string, replicas int32) error
    ScaleDeployment(ctx context.Context, namespace, name string, replicas int32) error
    DeleteDeployment(ctx context.Context, namespace, name string) error
    ListPods(ctx context.Context, namespace, labelSelector string) ([]PodInfo, error)
}

type k8sAdapterImpl struct {
    clientset *kubernetes.Clientset
}

func NewK8sAdapter(config *K8sConfig) (K8sAdapter, error) {
    // 初始化 K8s 客户端
    clientset, err := kubernetes.NewForConfig(config.RestConfig)
    if err != nil {
        return nil, err
    }
    
    return &k8sAdapterImpl{clientset: clientset}, nil
}
```

---

#### 4.2 LBAdapter

**职责**: 负载均衡器配置管理

```go
// internal/adapter/lb_adapter.go
package adapter

type LBAdapter interface {
    UpdateRules(ctx context.Context, serviceName string, rules []LBRule) error
    GetCurrentRules(ctx context.Context, serviceName string) ([]LBRule, error)
}
```

---

## 🎨 前端模块设计

### 1. API 模块

**职责**: 封装所有后端 API 调用

```typescript
// src/api/topology.ts
import request from '@/utils/request';
import type { TopologyDetail, TopologyList, CreateTopologyRequest } from '@/types/topology';

export const topologyAPI = {
  // 获取拓扑列表
  getTopologies: (params?: {
    service_name?: string;
    status?: string;
    page?: number;
    page_size?: number;
  }) => {
    return request.get<TopologyList>('/topologies', { params });
  },
  
  // 获取拓扑详情
  getTopology: (id: string) => {
    return request.get<TopologyDetail>(`/topologies/${id}`);
  },
  
  // 创建拓扑
  createTopology: (data: CreateTopologyRequest) => {
    return request.post('/topologies', data);
  },
  
  // 更新拓扑
  updateTopology: (id: string, data: Partial<CreateTopologyRequest>) => {
    return request.put(`/topologies/${id}`, data);
  },
  
  // 删除拓扑
  deleteTopology: (id: string) => {
    return request.delete(`/topologies/${id}`);
  },
  
  // 添加节点
  addNode: (topologyId: string, data: any) => {
    return request.post(`/topologies/${topologyId}/nodes`, data);
  },
  
  // 导出 LB 配置
  exportLBConfig: (id: string) => {
    return request.get(`/topologies/${id}/lb-config`);
  },
};
```

---

### 2. Store 模块

**职责**: 全局状态管理

```typescript
// src/stores/topologyStore.ts
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { TopologyData, TopologyMode } from '@/types/topology';
import { topologyAPI } from '@/api/topology';

export const useTopologyStore = defineStore('topology', () => {
  // State
  const currentTopologyId = ref<string | null>(null);
  const topologyData = ref<TopologyData>({ nodes: [], edges: [] });
  const mode = ref<TopologyMode>('view');
  const isLoading = ref(false);
  
  // Getters
  const serviceNodes = computed(() => 
    topologyData.value.nodes.filter(n => n.type === 'service')
  );
  
  const lbNodes = computed(() => 
    topologyData.value.nodes.filter(n => n.type === 'lb')
  );
  
  // Actions
  async function loadTopology(id: string) {
    isLoading.value = true;
    try {
      const response = await topologyAPI.getTopology(id);
      topologyData.value = {
        nodes: response.data.nodes,
        edges: response.data.edges,
      };
      currentTopologyId.value = id;
    } catch (error) {
      console.error('Failed to load topology:', error);
    } finally {
      isLoading.value = false;
    }
  }
  
  function setMode(newMode: TopologyMode) {
    mode.value = newMode;
  }
  
  function addNode(node: any) {
    topologyData.value.nodes.push(node);
  }
  
  function removeNode(nodeId: string) {
    const index = topologyData.value.nodes.findIndex(n => n.id === nodeId);
    if (index !== -1) {
      topologyData.value.nodes.splice(index, 1);
    }
  }
  
  function addEdge(edge: any) {
    topologyData.value.edges.push(edge);
  }
  
  function removeEdge(edgeId: string) {
    const index = topologyData.value.edges.findIndex(e => e.id === edgeId);
    if (index !== -1) {
      topologyData.value.edges.splice(index, 1);
    }
  }
  
  function reset() {
    topologyData.value = { nodes: [], edges: [] };
    currentTopologyId.value = null;
    mode.value = 'view';
  }
  
  return {
    // State
    currentTopologyId,
    topologyData,
    mode,
    isLoading,
    // Getters
    serviceNodes,
    lbNodes,
    // Actions
    loadTopology,
    setMode,
    addNode,
    removeNode,
    addEdge,
    removeEdge,
    reset,
  };
});
```

---

### 3. Component 模块

#### 3.1 TopologyGraph (业务组件)

**职责**: 拓扑图渲染核心组件

```typescript
// src/components/business/TopologyGraph.vue
<script setup lang="ts">
import { ref, onMounted, watch, onBeforeUnmount } from 'vue';
import { Graph } from '@antv/g6';
import { useTopologyStore } from '@/stores/topologyStore';

const topologyStore = useTopologyStore();
const containerRef = ref<HTMLDivElement | null>(null);
let graph: Graph | null = null;

onMounted(() => {
  if (!containerRef.value) return;
  
  graph = new Graph({
    container: containerRef.value,
    width: containerRef.value.offsetWidth,
    height: containerRef.value.offsetHeight,
    layout: {
      type: 'dagre',
      rankdir: 'LR',
      nodesep: 60,
      ranksep: 180,
    },
    // ... 其他配置
  });
  
  renderGraph();
});

function renderGraph() {
  if (!graph) return;
  
  const { nodes, edges } = topologyStore.topologyData;
  
  // 转换数据格式
  const g6Nodes = nodes.map(node => ({
    id: node.id,
    data: node,
    style: {
      // 样式配置
    },
  }));
  
  const g6Edges = edges.map(edge => ({
    id: edge.id,
    source: edge.source,
    target: edge.target,
    data: edge,
  }));
  
  graph.setData({ nodes: g6Nodes, edges: g6Edges });
  graph.render();
}

watch(() => topologyStore.topologyData, renderGraph, { deep: true });

onBeforeUnmount(() => {
  if (graph) {
    graph.destroy();
  }
});
</script>

<template>
  <div class="topology-graph-container">
    <div ref="containerRef" class="graph-canvas" />
  </div>
</template>
```

---

#### 3.2 DeploymentPanel (业务组件)

**职责**: 发布任务控制面板

```typescript
// src/components/business/DeploymentPanel.vue
<script setup lang="ts">
import { ref, computed } from 'vue';
import { deploymentAPI } from '@/api/deployment';
import type { Deployment } from '@/types/deployment';

interface Props {
  deploymentId: string;
}

const props = defineProps<Props>();
const deployment = ref<Deployment | null>(null);
const isLoading = ref(false);

const canStart = computed(() => deployment.value?.status === 'pending');
const canPause = computed(() => deployment.value?.status === 'running');
const canContinue = computed(() => deployment.value?.status === 'paused');
const canRollback = computed(() => ['running', 'paused'].includes(deployment.value?.status || ''));

async function loadDeployment() {
  isLoading.value = true;
  try {
    const response = await deploymentAPI.getDeployment(props.deploymentId);
    deployment.value = response.data;
  } finally {
    isLoading.value = false;
  }
}

async function startDeployment() {
  await deploymentAPI.startDeployment(props.deploymentId);
  loadDeployment();
}

async function pauseDeployment() {
  await deploymentAPI.pauseDeployment(props.deploymentId);
  loadDeployment();
}

async function continueDeployment() {
  await deploymentAPI.continueDeployment(props.deploymentId);
  loadDeployment();
}

async function rollbackDeployment() {
  const reason = prompt('请输入回滚原因:');
  if (reason) {
    await deploymentAPI.rollbackDeployment(props.deploymentId, reason);
    loadDeployment();
  }
}
</script>

<template>
  <div class="deployment-panel">
    <div class="panel-header">
      <h3>发布任务控制</h3>
      <span class="status-badge">{{ deployment?.status }}</span>
    </div>
    
    <div class="panel-body">
      <div class="step-progress">
        当前步骤: {{ deployment?.current_step }} / {{ deployment?.total_steps }}
      </div>
      
      <div class="action-buttons">
        <button v-if="canStart" @click="startDeployment">启动</button>
        <button v-if="canPause" @click="pauseDeployment">暂停</button>
        <button v-if="canContinue" @click="continueDeployment">继续</button>
        <button v-if="canRollback" @click="rollbackDeployment" class="danger">回滚</button>
      </div>
    </div>
  </div>
</template>
```

---

## 🔗 模块间交互流程

### 创建拓扑并添加节点

```
Frontend (Vue Component)
    ↓ 用户填写表单
TopologyStore.createTopology()
    ↓
TopologyAPI.createTopology()
    ↓ Axios POST /api/v1/topologies
Backend Router
    ↓
AuthMiddleware (验证 JWT)
    ↓
TopologyHandler.CreateTopology()
    ↓ 解析请求参数
TopologyService.CreateTopology()
    ↓ 业务逻辑处理
TopologyRepository.Create()
    ↓ GORM 操作
PostgreSQL (插入数据)
    ↓ 返回结果
TopologyHandler 封装响应
    ↓ JSON 序列化
Frontend 接收响应
    ↓
TopologyStore 更新状态
    ↓
Vue Component 重新渲染
```

---

## 📝 依赖注入与模块初始化

### 后端依赖注入

```go
// cmd/server/main.go
package main

import (
    "x-mindflow/internal/api/handler"
    "x-mindflow/internal/api/router"
    "x-mindflow/internal/repository"
    "x-mindflow/internal/service"
    "x-mindflow/pkg/config"
    "x-mindflow/pkg/database"
)

func main() {
    // 1. 加载配置
    cfg := config.LoadConfig()
    
    // 2. 初始化数据库
    db := database.InitDB(cfg.Database)
    
    // 3. 初始化 Repository
    topologyRepo := repository.NewTopologyRepository(db)
    nodeRepo := repository.NewTopologyNodeRepository(db)
    edgeRepo := repository.NewTopologyEdgeRepository(db)
    deploymentRepo := repository.NewDeploymentRepository(db)
    
    // 4. 初始化 Service
    topologyService := service.NewTopologyService(topologyRepo, nodeRepo, edgeRepo)
    deploymentService := service.NewDeploymentService(deploymentRepo, instanceService, lbAdapter)
    
    // 5. 初始化 Handler
    topologyHandler := handler.NewTopologyHandler(topologyService)
    deploymentHandler := handler.NewDeploymentHandler(deploymentService)
    
    // 6. 设置路由
    r := router.SetupRouter(topologyHandler, deploymentHandler, authHandler)
    
    // 7. 启动服务
    r.Run(":8080")
}
```

---

## 📚 相关文档

- **API 设计**: `docs/api/22-topology-api-design.md`
- **数据库设计**: `docs/database/22-topology-schema-design.md`
- **架构设计**: `docs/architecture/22-topology-architecture-design.md`
- **原型文档**: `docs/prototypes/11-topology-visualization-prototype.md`

---

**文档版本**: v1.0  
**最后更新**: 2025-10-24  
**维护人**: xgopilot (Claude Code AI Assistant)
