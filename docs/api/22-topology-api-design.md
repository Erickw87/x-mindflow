# 拓扑可视化与发布管理 API 设计文档

**关联 Issue**: #22  
**文档类型**: API 设计  
**创建日期**: 2025-10-24  
**状态**: 设计阶段

---

## 📋 概述

本文档定义了拓扑可视化与发布管理模块的 RESTful API 接口,包括拓扑图管理、服务节点管理、LB 配置管理和发布流程管理的所有端点。

### 设计原则

1. **RESTful 规范**: 遵循 REST 最佳实践,使用标准 HTTP 方法
2. **统一响应格式**: 所有接口返回统一的 JSON 结构
3. **版本控制**: API 路径包含版本号 `/api/v1`
4. **错误处理**: 统一的错误码和错误信息格式
5. **认证鉴权**: 使用 JWT Token 进行身份验证

---

## 🌐 API 基础信息

### Base URL

```
http://localhost:8080/api/v1
```

### 通用响应格式

#### 成功响应

```json
{
  "code": 200,
  "message": "success",
  "data": {
    // 具体数据
  }
}
```

#### 错误响应

```json
{
  "code": 400,
  "message": "错误描述",
  "details": "详细错误信息"
}
```

### 状态码规范

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 400 | 请求参数错误 |
| 401 | 未授权(Token 无效或过期) |
| 403 | 禁止访问(权限不足) |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 500 | 服务器内部错误 |

### 认证方式

所有需要认证的接口在请求头中携带 JWT Token:

```
Authorization: Bearer <token>
```

---

## 📊 拓扑管理 API

### 1. 获取拓扑列表

**端点**: `GET /topologies`

**描述**: 获取所有拓扑树列表

**请求参数**:

Query Parameters:
```typescript
{
  service_name?: string;    // 过滤: 服务名称
  status?: string;          // 过滤: 状态 (active/inactive)
  page?: number;            // 分页: 页码,默认 1
  page_size?: number;       // 分页: 每页大小,默认 20
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 5,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": "topo-001",
        "name": "user-service-v1",
        "description": "用户服务主拓扑树",
        "service_name": "user-service",
        "status": "active",
        "version": 1,
        "cloned_from": null,
        "created_at": "2025-10-24T10:00:00Z",
        "updated_at": "2025-10-24T11:30:00Z",
        "created_by": "admin"
      }
    ]
  }
}
```

---

### 2. 获取拓扑详情

**端点**: `GET /topologies/:id`

**描述**: 获取指定拓扑树的详细信息,包括节点和边

**路径参数**:
- `id`: 拓扑ID

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "topo-001",
    "name": "user-service-v1",
    "description": "用户服务主拓扑树",
    "service_name": "user-service",
    "status": "active",
    "version": 1,
    "nodes": [
      {
        "id": "node-entry",
        "type": "service",
        "name": "入口网关",
        "versions": [
          {
            "version": "v2.0",
            "instance_count": 10,
            "color": "#3b82f6"
          }
        ],
        "total_traffic_percentage": 100,
        "total_instances": 10,
        "branch": "default"
      },
      {
        "id": "node-lb1",
        "type": "lb",
        "rules": [
          {
            "name": "规则1",
            "target_versions": ["v1.0"],
            "traffic_percentage": 70,
            "rule_details": "Header: x-version=stable"
          }
        ],
        "branch": "default"
      }
    ],
    "edges": [
      {
        "id": "edge-1",
        "source": "node-entry",
        "target": "node-lb1",
        "traffic_percentage": 100,
        "label": "100% 流量",
        "branch": "default"
      }
    ],
    "created_at": "2025-10-24T10:00:00Z",
    "updated_at": "2025-10-24T11:30:00Z"
  }
}
```

---

### 3. 创建拓扑

**端点**: `POST /topologies`

**描述**: 创建新的拓扑树

**请求体**:

```json
{
  "name": "user-service-v2",
  "description": "用户服务 v2 版本拓扑树",
  "service_name": "user-service",
  "cloned_from": "topo-001"
}
```

**请求字段说明**:

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 拓扑名称 |
| description | string | 否 | 描述信息 |
| service_name | string | 是 | 关联的服务名称 |
| cloned_from | string | 否 | 克隆源拓扑 ID(为空则创建空拓扑) |

**响应示例**:

```json
{
  "code": 201,
  "message": "拓扑创建成功",
  "data": {
    "id": "topo-002",
    "name": "user-service-v2",
    "description": "用户服务 v2 版本拓扑树",
    "service_name": "user-service",
    "status": "active",
    "version": 1,
    "cloned_from": "topo-001",
    "created_at": "2025-10-24T12:00:00Z",
    "updated_at": "2025-10-24T12:00:00Z",
    "created_by": "admin"
  }
}
```

---

### 4. 更新拓扑

**端点**: `PUT /topologies/:id`

**描述**: 更新拓扑的基本信息或配置

**路径参数**:
- `id`: 拓扑ID

**请求体**:

```json
{
  "name": "user-service-v2-updated",
  "description": "更新后的描述",
  "status": "inactive"
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "拓扑更新成功",
  "data": {
    "id": "topo-002",
    "name": "user-service-v2-updated",
    "description": "更新后的描述",
    "status": "inactive",
    "updated_at": "2025-10-24T12:30:00Z"
  }
}
```

---

### 5. 删除拓扑

**端点**: `DELETE /topologies/:id`

**描述**: 删除指定拓扑树(软删除)

**路径参数**:
- `id`: 拓扑ID

**响应示例**:

```json
{
  "code": 200,
  "message": "拓扑删除成功",
  "data": {
    "id": "topo-002",
    "deleted_at": "2025-10-24T13:00:00Z"
  }
}
```

---

### 6. 获取拓扑配置(导出为 LB 配置)

**端点**: `GET /topologies/:id/lb-config`

**描述**: 将拓扑树转换为 LB 配置格式

**路径参数**:
- `id`: 拓扑ID

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "topology_id": "topo-001",
    "service_name": "user-service",
    "lb_configs": [
      {
        "lb_id": "lb1",
        "rules": [
          {
            "name": "规则1",
            "conditions": {
              "header": "x-version=stable"
            },
            "targets": [
              {
                "version": "v1.0",
                "weight": 70
              }
            ]
          }
        ]
      }
    ],
    "generated_at": "2025-10-24T14:00:00Z"
  }
}
```

---

## 🔧 节点管理 API

### 7. 添加节点到拓扑

**端点**: `POST /topologies/:topology_id/nodes`

**描述**: 向拓扑树中添加服务节点或 LB 节点

**路径参数**:
- `topology_id`: 拓扑ID

**请求体(服务节点)**:

```json
{
  "type": "service",
  "name": "服务 A",
  "versions": [
    {
      "version": "v1.1",
      "instance_count": 6,
      "color": "#f59e0b",
      "is_target": true
    }
  ],
  "total_traffic_percentage": 30,
  "branch": "v1.1"
}
```

**请求体(LB 节点)**:

```json
{
  "type": "lb",
  "rules": [
    {
      "name": "规则1",
      "target_versions": ["v1.1"],
      "traffic_percentage": 30,
      "rule_details": "Header: x-version=canary"
    }
  ],
  "branch": "v1.1"
}
```

**响应示例**:

```json
{
  "code": 201,
  "message": "节点添加成功",
  "data": {
    "id": "node-abc123",
    "type": "service",
    "name": "服务 A",
    "created_at": "2025-10-24T15:00:00Z"
  }
}
```

---

### 8. 更新节点

**端点**: `PUT /topologies/:topology_id/nodes/:node_id`

**描述**: 更新节点信息(如实例数、流量权重等)

**路径参数**:
- `topology_id`: 拓扑ID
- `node_id`: 节点ID

**请求体**:

```json
{
  "versions": [
    {
      "version": "v1.1",
      "instance_count": 10,
      "color": "#f59e0b"
    }
  ],
  "total_traffic_percentage": 50
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "节点更新成功",
  "data": {
    "id": "node-abc123",
    "updated_at": "2025-10-24T15:30:00Z"
  }
}
```

---

### 9. 删除节点

**端点**: `DELETE /topologies/:topology_id/nodes/:node_id`

**描述**: 从拓扑树中删除节点

**路径参数**:
- `topology_id`: 拓扑ID
- `node_id`: 节点ID

**响应示例**:

```json
{
  "code": 200,
  "message": "节点删除成功",
  "data": {
    "id": "node-abc123",
    "deleted_at": "2025-10-24T16:00:00Z"
  }
}
```

---

## 🔗 连接边管理 API

### 10. 添加连接边

**端点**: `POST /topologies/:topology_id/edges`

**描述**: 在拓扑树中添加节点间的连接边

**路径参数**:
- `topology_id`: 拓扑ID

**请求体**:

```json
{
  "source": "node-entry",
  "target": "node-lb1",
  "traffic_percentage": 100,
  "label": "100% 流量",
  "branch": "default"
}
```

**响应示例**:

```json
{
  "code": 201,
  "message": "连接边添加成功",
  "data": {
    "id": "edge-xyz789",
    "source": "node-entry",
    "target": "node-lb1",
    "created_at": "2025-10-24T16:30:00Z"
  }
}
```

---

### 11. 更新连接边

**端点**: `PUT /topologies/:topology_id/edges/:edge_id`

**描述**: 更新连接边的流量权重

**路径参数**:
- `topology_id`: 拓扑ID
- `edge_id`: 边ID

**请求体**:

```json
{
  "traffic_percentage": 50,
  "label": "50% 流量"
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "连接边更新成功",
  "data": {
    "id": "edge-xyz789",
    "updated_at": "2025-10-24T17:00:00Z"
  }
}
```

---

### 12. 删除连接边

**端点**: `DELETE /topologies/:topology_id/edges/:edge_id`

**描述**: 删除拓扑树中的连接边

**路径参数**:
- `topology_id`: 拓扑ID
- `edge_id`: 边ID

**响应示例**:

```json
{
  "code": 200,
  "message": "连接边删除成功",
  "data": {
    "id": "edge-xyz789",
    "deleted_at": "2025-10-24T17:30:00Z"
  }
}
```

---

## 🚀 发布流程管理 API

### 13. 创建发布任务

**端点**: `POST /deployments`

**描述**: 创建新的发布任务

**请求体**:

```json
{
  "topology_id": "topo-001",
  "service_name": "user-service",
  "new_version": "v1.2.0",
  "rollout_plan": {
    "steps": [
      {
        "step": 1,
        "instances_percent": 10,
        "traffic_percent": 10,
        "pause_type": "manual",
        "timeout": "30m"
      },
      {
        "step": 2,
        "instances_percent": 30,
        "traffic_percent": 30,
        "pause_type": "auto",
        "pause_duration": "5m",
        "timeout": "30m"
      },
      {
        "step": 3,
        "instances_percent": 100,
        "traffic_percent": 100,
        "pause_type": "auto",
        "pause_duration": "10m",
        "timeout": "30m"
      }
    ]
  }
}
```

**响应示例**:

```json
{
  "code": 201,
  "message": "发布任务创建成功",
  "data": {
    "id": "deploy-001",
    "topology_id": "topo-001",
    "service_name": "user-service",
    "new_version": "v1.2.0",
    "status": "pending",
    "current_step": 0,
    "created_at": "2025-10-24T18:00:00Z",
    "created_by": "admin"
  }
}
```

---

### 14. 获取发布任务列表

**端点**: `GET /deployments`

**描述**: 获取发布任务列表

**请求参数**:

Query Parameters:
```typescript
{
  service_name?: string;
  status?: string;         // pending/running/paused/completed/failed/rolled_back
  page?: number;
  page_size?: number;
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 10,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "id": "deploy-001",
        "topology_id": "topo-001",
        "service_name": "user-service",
        "new_version": "v1.2.0",
        "status": "running",
        "current_step": 2,
        "total_steps": 3,
        "created_at": "2025-10-24T18:00:00Z",
        "updated_at": "2025-10-24T18:30:00Z"
      }
    ]
  }
}
```

---

### 15. 获取发布任务详情

**端点**: `GET /deployments/:id`

**描述**: 获取发布任务的详细信息和执行状态

**路径参数**:
- `id`: 发布任务ID

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": "deploy-001",
    "topology_id": "topo-001",
    "service_name": "user-service",
    "new_version": "v1.2.0",
    "old_version": "v1.1.0",
    "status": "running",
    "current_step": 2,
    "total_steps": 3,
    "rollout_plan": {
      "steps": [
        {
          "step": 1,
          "status": "completed",
          "instances_percent": 10,
          "traffic_percent": 10,
          "started_at": "2025-10-24T18:00:00Z",
          "completed_at": "2025-10-24T18:10:00Z"
        },
        {
          "step": 2,
          "status": "running",
          "instances_percent": 30,
          "traffic_percent": 30,
          "started_at": "2025-10-24T18:15:00Z",
          "completed_at": null
        }
      ]
    },
    "created_at": "2025-10-24T18:00:00Z",
    "updated_at": "2025-10-24T18:20:00Z",
    "created_by": "admin"
  }
}
```

---

### 16. 启动发布任务

**端点**: `POST /deployments/:id/start`

**描述**: 启动待执行的发布任务

**路径参数**:
- `id`: 发布任务ID

**响应示例**:

```json
{
  "code": 200,
  "message": "发布任务已启动",
  "data": {
    "id": "deploy-001",
    "status": "running",
    "current_step": 1,
    "started_at": "2025-10-24T18:00:00Z"
  }
}
```

---

### 17. 继续发布(手动确认)

**端点**: `POST /deployments/:id/continue`

**描述**: 当发布任务在手动暂停点时,继续执行下一步

**路径参数**:
- `id`: 发布任务ID

**响应示例**:

```json
{
  "code": 200,
  "message": "发布任务已继续",
  "data": {
    "id": "deploy-001",
    "status": "running",
    "current_step": 2,
    "resumed_at": "2025-10-24T18:15:00Z"
  }
}
```

---

### 18. 暂停发布

**端点**: `POST /deployments/:id/pause`

**描述**: 暂停正在执行的发布任务

**路径参数**:
- `id`: 发布任务ID

**响应示例**:

```json
{
  "code": 200,
  "message": "发布任务已暂停",
  "data": {
    "id": "deploy-001",
    "status": "paused",
    "paused_at": "2025-10-24T18:20:00Z"
  }
}
```

---

### 19. 回滚发布

**端点**: `POST /deployments/:id/rollback`

**描述**: 回滚发布任务到原始状态

**路径参数**:
- `id`: 发布任务ID

**请求体**:

```json
{
  "reason": "发现严重 Bug,立即回滚"
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "发布任务已回滚",
  "data": {
    "id": "deploy-001",
    "status": "rolled_back",
    "rollback_reason": "发现严重 Bug,立即回滚",
    "rolled_back_at": "2025-10-24T18:25:00Z"
  }
}
```

---

### 20. 取消发布

**端点**: `DELETE /deployments/:id`

**描述**: 取消待执行或暂停中的发布任务

**路径参数**:
- `id`: 发布任务ID

**响应示例**:

```json
{
  "code": 200,
  "message": "发布任务已取消",
  "data": {
    "id": "deploy-001",
    "status": "cancelled",
    "cancelled_at": "2025-10-24T18:30:00Z"
  }
}
```

---

## 📈 监控与状态查询 API

### 21. 获取服务实例状态

**端点**: `GET /services/:service_name/instances`

**描述**: 查询指定服务的所有实例状态

**路径参数**:
- `service_name`: 服务名称

**请求参数**:

Query Parameters:
```typescript
{
  version?: string;        // 过滤: 版本号
  status?: string;         // 过滤: 状态 (running/stopped/unhealthy)
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "service_name": "user-service",
    "instances": [
      {
        "id": "instance-001",
        "version": "v1.1.0",
        "status": "running",
        "health": "healthy",
        "cpu_usage": 45.3,
        "memory_usage": 60.2,
        "started_at": "2025-10-24T10:00:00Z",
        "last_health_check": "2025-10-24T18:40:00Z"
      }
    ]
  }
}
```

---

### 22. 获取流量统计

**端点**: `GET /topologies/:topology_id/traffic-stats`

**描述**: 获取拓扑树的流量统计信息

**路径参数**:
- `topology_id`: 拓扑ID

**请求参数**:

Query Parameters:
```typescript
{
  time_range?: string;     // 时间范围: 1h/6h/24h/7d,默认 1h
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "topology_id": "topo-001",
    "time_range": "1h",
    "total_requests": 1000000,
    "nodes_traffic": [
      {
        "node_id": "node-service-a-v1",
        "request_count": 700000,
        "traffic_percentage": 70.0,
        "avg_latency_ms": 50,
        "error_rate": 0.01
      },
      {
        "node_id": "node-service-a-v2",
        "request_count": 300000,
        "traffic_percentage": 30.0,
        "avg_latency_ms": 55,
        "error_rate": 0.02
      }
    ],
    "generated_at": "2025-10-24T19:00:00Z"
  }
}
```

---

## 🔐 认证与授权 API

### 23. 用户登录

**端点**: `POST /auth/login`

**描述**: 用户登录,获取 JWT Token

**请求体**:

```json
{
  "username": "admin",
  "password": "password123"
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600,
    "user": {
      "id": "user-001",
      "username": "admin",
      "role": "admin"
    }
  }
}
```

---

### 24. 刷新 Token

**端点**: `POST /auth/refresh`

**描述**: 使用 Refresh Token 刷新 Access Token

**请求体**:

```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**响应示例**:

```json
{
  "code": 200,
  "message": "Token 刷新成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 3600
  }
}
```

---

## 🚨 错误码规范

### 通用错误码

| 错误码 | 说明 | HTTP 状态码 |
|--------|------|-------------|
| 1000 | 未知错误 | 500 |
| 1001 | 请求参数错误 | 400 |
| 1002 | 资源不存在 | 404 |
| 1003 | 资源已存在(冲突) | 409 |
| 1004 | 未授权 | 401 |
| 1005 | 权限不足 | 403 |

### 业务错误码

| 错误码 | 说明 | HTTP 状态码 |
|--------|------|-------------|
| 2001 | 拓扑不存在 | 404 |
| 2002 | 拓扑名称已存在 | 409 |
| 2003 | 节点不存在 | 404 |
| 2004 | 边不存在 | 404 |
| 2005 | 拓扑包含节点,无法删除 | 409 |
| 3001 | 发布任务不存在 | 404 |
| 3002 | 发布任务状态不允许此操作 | 400 |
| 3003 | 发布步骤配置错误 | 400 |
| 3004 | 实例创建失败 | 500 |
| 3005 | LB 配置更新失败 | 500 |

---

## 📝 请求/响应示例完整流程

### 场景: 创建灰度发布任务并执行

#### 1. 获取拓扑列表

**请求**:
```http
GET /api/v1/topologies?service_name=user-service
Authorization: Bearer <token>
```

**响应**: (略,参考 API 1)

#### 2. 创建发布任务

**请求**:
```http
POST /api/v1/deployments
Authorization: Bearer <token>
Content-Type: application/json

{
  "topology_id": "topo-001",
  "service_name": "user-service",
  "new_version": "v1.2.0",
  "rollout_plan": {
    "steps": [
      {
        "step": 1,
        "instances_percent": 10,
        "traffic_percent": 10,
        "pause_type": "manual",
        "timeout": "30m"
      }
    ]
  }
}
```

**响应**: (参考 API 13)

#### 3. 启动发布任务

**请求**:
```http
POST /api/v1/deployments/deploy-001/start
Authorization: Bearer <token>
```

**响应**: (参考 API 16)

#### 4. 监控发布进度

**请求**:
```http
GET /api/v1/deployments/deploy-001
Authorization: Bearer <token>
```

**响应**: (参考 API 15)

#### 5. 手动确认,继续下一步

**请求**:
```http
POST /api/v1/deployments/deploy-001/continue
Authorization: Bearer <token>
```

**响应**: (参考 API 17)

---

## 🔄 API 版本演进策略

### 版本策略

- 当前版本: `v1`
- 向后兼容: 小版本更新不破坏现有接口
- 废弃通知: 废弃的接口在响应头中标注 `X-API-Deprecated: true`
- 版本切换: 新版本通过路径区分,如 `/api/v2/...`

### 废弃流程

1. 在新版本中标记为 `@deprecated`
2. 响应头添加 `X-API-Deprecated: true`
3. 至少保留 3 个月过渡期
4. 发布公告后正式下线

---

## 📚 相关文档

- **需求文档**: `docs/requirements/010-requirement-multi-version-canary.md`
- **原型文档**: `docs/prototypes/11-topology-visualization-prototype.md`
- **数据库设计**: `docs/database/22-topology-schema-design.md`
- **架构设计**: `docs/architecture/22-topology-architecture-design.md`
- **模块设计**: `docs/modules/22-topology-module-design.md`

---

**文档版本**: v1.0  
**最后更新**: 2025-10-24  
**维护人**: xgopilot (Claude Code AI Assistant)
