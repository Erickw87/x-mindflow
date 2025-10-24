# LB 流量控制与拓扑管理 - 技术需求文档

**Issue**: #9  
**版本**: v1.0  
**日期**: 2025-10-24  
**状态**: 需求确认阶段

---

## 📌 概述

本文档详细描述 x-mindflow 智能发布系统中 **LB 流量控制与拓扑管理**模块的技术需求和实现细节。该模块是系统的核心，通过负载均衡器配置实现流量分流、多版本并行发布和拓扑可视化管理。

---

## 🎯 功能目标

通过 LB 配置实现流量分流和拓扑可视化，实现"拓扑即配置"的设计理念。

---

## 🏗️ 核心设计思想

### 1. LB 配置即拓扑

**关键原则**：
- 程序配置中移除下游 host，改为引用 LB 配置名称
- 拓扑结构完全由 LB 配置定义
- 配置生图、图生配置双向稳定

**LB 配置包含**：
- 下游服务列表及权重分配
- 流量路由规则（使用 `X-LB-Route-ID` Header 标识）
- LB 规则与实例映射关系

---

### 2. 流量路由机制

#### 2.1 LB 路由规则

**基础规则**：
- 每个服务的 LB 至少有一个**默认路由规则**（Default Route）
- 默认规则采用轮询（Round Robin）方式负载均衡，流量平均分配到所有实例
- 可定义多个自定义路由规则，通过 `X-LB-Route-ID` Header 匹配

**多规则路由示例**：

```yaml
# B 服务的 LB 配置
service: service-b
routes:
  - id: default                    # 默认路由
    weight: 100
    instances:
      - service-b-v1-instance-1
      - service-b-v1-instance-2
  
  - id: canary-route                # 灰度路由
    match:
      header: X-LB-Route-ID
      value: canary-route
    instances:
      - service-b-v2-instance-1
```

**路由匹配逻辑**：
1. 检查请求中是否存在 `X-LB-Route-ID` Header
2. 如果存在且匹配某个路由规则 ID，则使用该规则
3. 如果不存在或无匹配，则使用默认路由
4. 如果指定的路由规则不存在，降级到默认路由（应记录告警日志）

---

#### 2.2 流量标签注入机制

**标签注入流程**（以 A 服务调用 B 服务为例）：

**场景 1：入口流量无标签**

```
1. 请求到达 A 服务 LB（无 X-LB-Route-ID）
2. A 服务 LB 根据配置对部分流量注入标签
   - 参考 nginx split_clients 机制实现按比例分流
   - 例如：30% 流量注入 X-LB-Route-ID: canary-route
3. 流量路由到 A 服务实例
4. A 服务透传 X-LB-Route-ID 到下游 B 服务
5. B 服务 LB 根据 X-LB-Route-ID 路由到对应实例组
```

**场景 2：入口流量已有标签**

```
1. 请求到达 A 服务 LB（携带 X-LB-Route-ID: canary-route）
2. A 服务 LB 根据标签路由到对应实例组
3. A 服务实例透传标签到 B 服务
4. B 服务 LB 解析标签，路由到对应实例组
```

**LB 配置示例（A 服务）**：

```yaml
service: service-a
routes:
  - id: default
    weight: 70                      # 70% 流量走默认路由
    instances:
      - service-a-v1-instance-1
      - service-a-v1-instance-2
  
  - id: canary-route
    weight: 30                      # 30% 流量注入 canary-route 标签
    inject_header:
      name: X-LB-Route-ID
      value: canary-route
    instances:
      - service-a-v2-instance-1
```

**权重分流实现**：
- 参考 nginx `split_clients` 指令
- 基于客户端 IP 或其他哈希值实现一致性分流
- 确保同一用户/会话的流量保持在同一路由

---

#### 2.3 Header 透传机制

**实现方式**：
- 应用程序接入透传框架/中间件
- 自动透传 `X-LB-Route-ID` 等指定 Header
- 当前 mock-service 已实现透传（见 `/mock-service/internal/handler/proxy.go:94-118`）

**必须透传的 Headers**：
- `X-Request-ID`: 请求追踪
- `X-LB-Route-ID`: 路由规则标识（新增）
- `X-Traffic-Tag`: 流量标签（已有）
- `X-Forwarded-For`: 客户端 IP
- `Authorization`: 认证信息

**透传框架要求**：
- 提供多语言 SDK（Go、Java、Python 等）
- 支持常见 Web 框架（Gin、Spring Boot、Flask 等）
- 零配置或最小化配置
- 性能开销低于 1ms

**验证状态**：
- ✅ mock-service 已实现（`/mock-service/internal/handler/proxy.go`）
- ⏳ 生产环境 SDK 待开发

---

### 3. 流量出入口定义

**明确定义**：

**东西向流量**（服务间调用）：
- A 服务 → B 服务的出流量**必须经过 B 服务的 LB**
- A 服务的出流量**不再经过 A 服务的 LB**
- 目的：实现服务间的流量控制和路由

**南北向流量**（客户端-服务）：
- 客户端 → 服务的入流量通过统一网关（Gateway）
- 网关负责安全认证、限流、协议转换
- 服务 → 客户端的响应流量原路返回

**流量路径示例**：

```
┌─────────┐
│ Client  │
└────┬────┘
     │
     ↓
┌─────────────┐
│   Gateway   │ (南北向入口)
└──────┬──────┘
       │
       ↓
┌──────────────┐
│ Service A LB │ (东西向入口)
└──────┬───────┘
       │
       ↓
┌──────────────────┐
│ Service A        │
│ Instance         │ (应用实例)
└──────┬───────────┘
       │ (调用 Service B)
       ↓
┌──────────────┐
│ Service B LB │ (东西向入口)
└──────┬───────┘
       │
       ↓
┌──────────────────┐
│ Service B        │
│ Instance         │
└──────────────────┘
```

---

### 4. 拓扑树生成与编辑

**生成流程**：
1. 从 PostgreSQL 数据库查询所有 LB 配置
2. 解析 LB 配置构建服务依赖关系图
3. 计算流量权重和路由规则
4. 使用前端图可视化库（AntV G6）渲染拓扑图

**编辑能力**：
- 图形化界面修改 LB 配置（避免手动改 YAML）
- 支持添加/删除服务节点
- 调整路由权重和规则
- 实时预览配置变更效果

**配置管理**：
- Git 版本控制（配置文件存储在 Git 仓库）
- 配置 diff 对比（变更前后对比）
- 一键回滚能力（回滚到历史版本）
- 变更审批流程（可选）

---

## 📋 最小功能集（MVP）

### 1. LB 配置模板（YAML/JSON）

**Schema 定义**：

```yaml
service: string                      # 服务名称
version: string                      # 配置版本
routes:                              # 路由规则列表
  - id: string                       # 路由规则 ID（唯一）
    weight: integer                  # 权重（0-100）
    match:                           # 匹配条件（可选）
      header: string                 # Header 名称
      value: string                  # Header 值
    inject_header:                   # 注入 Header（可选）
      name: string                   # Header 名称
      value: string                  # Header 值
    instances:                       # 实例列表
      - string                       # 实例标识符
```

**验证规则**：
- `service` 必填，格式为 `^[a-z][a-z0-9-]*$`
- `routes` 必填，至少包含一个默认路由
- 所有路由的 `weight` 总和必须为 100
- `id` 在同一服务内唯一
- `instances` 不能为空

**示例模板**：

```yaml
service: user-service
version: v1.0.0
routes:
  - id: default
    weight: 70
    instances:
      - user-service-v1-1
      - user-service-v1-2
  
  - id: canary-route
    weight: 30
    inject_header:
      name: X-LB-Route-ID
      value: canary-route
    instances:
      - user-service-v2-1
```

---

### 2. 拓扑图渲染引擎

**功能要求**：
- 解析 LB 配置生成服务依赖图
- 计算流量百分比（基于路由规则权重）
- 支持多层级服务调用链展示
- 实时更新拓扑图

**技术实现**：
- 后端：Go 解析 YAML 配置，构建图结构
- 前端：AntV G6 渲染拓扑图
- 数据格式：标准 JSON 图数据

**图结构示例**：

```json
{
  "nodes": [
    {
      "id": "service-a",
      "label": "Service A",
      "version": "v1.0"
    },
    {
      "id": "service-b",
      "label": "Service B",
      "version": "v1.0"
    }
  ],
  "edges": [
    {
      "source": "service-a",
      "target": "service-b",
      "label": "70%",
      "route": "default"
    },
    {
      "source": "service-a",
      "target": "service-b-v2",
      "label": "30%",
      "route": "canary-route"
    }
  ]
}
```

---

### 3. 图形化编辑器

**编辑能力**：
- 添加/删除节点（服务）
- 调整路由权重（拖动滑块）
- 修改路由规则（表单编辑）
- 添加/删除实例

**交互设计**：
- 点击节点显示详细信息
- 点击边显示路由规则
- 拖拽调整节点位置（布局）
- 实时验证配置合法性

**错误提示**：
- 权重总和不为 100 时高亮警告
- 实例列表为空时显示错误
- 循环依赖检测

---

### 4. 配置管理

**版本控制**：
- LB 配置文件存储在 Git 仓库
- 每次变更创建新的 commit
- Commit 消息记录变更内容

**配置对比**：
- 显示配置文件的 diff
- 高亮权重变更、实例增减
- 支持版本回滚预览

**回滚能力**：
- 选择历史版本一键回滚
- 回滚前显示影响范围
- 回滚操作记录审计日志

---

## ✅ 已明确的设计决策

### Q5：程序如何读取 LB 配置？

**答案**：从发布系统的 PostgreSQL 数据库读取

**实现方式**：
- 拓扑图生成器查询数据库中的所有 LB 配置
- LB 配置存储在 `lb_configs` 表（具体表结构待设计）
- 通过 RESTful API 查询当前运行中的所有 LB 配置
- 前端根据查询结果渲染拓扑图

**数据库表结构（初步设计）**：

```sql
CREATE TABLE lb_configs (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    config_yaml TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'active',
    UNIQUE(service_name, version)
);
```

---

### Q6：流量标签如何从配置变成可传递的 Header？

**答案**：LB 自动注入 `X-LB-Route-ID` Header

**实现方式**：
- 在 LB 配置中定义注入规则（`inject_header` 字段）
- LB（如 Nginx）根据权重分配规则自动注入 Header
- 应用程序无需感知，仅需透传
- 参考 nginx `split_clients` 和 `add_header` 指令实现

**Nginx 配置示例**：

```nginx
# 基于 split_clients 实现按比例分流
split_clients "${remote_addr}${request_id}" $route_id {
    30%     "canary-route";
    *       "default";
}

# 注入 Header
location / {
    add_header X-LB-Route-ID $route_id;
    
    # 根据 route_id 路由到不同 upstream
    if ($route_id = "canary-route") {
        proxy_pass http://service-a-canary;
    }
    proxy_pass http://service-a-default;
}
```

---

### Q7：如何避免 LB 标记丢失？

**答案**：应用接入透传框架/中间件

**实现方式**：
- 提供多语言 SDK/中间件（Go、Java、Python 等）
- 自动透传 `X-LB-Route-ID`、`X-Request-ID` 等 Header
- mock-service 已实现透传机制（无需额外开发）
- 真实业务应用需接入透传框架

**Go SDK 示例**：

```go
// Gin 中间件示例
func HeaderPropagation() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 需要透传的 Headers
        propagateHeaders := []string{
            "X-Request-ID",
            "X-LB-Route-ID",
            "X-Traffic-Tag",
            "X-Forwarded-For",
        }
        
        // 保存到 context
        for _, header := range propagateHeaders {
            if value := c.GetHeader(header); value != "" {
                c.Set(header, value)
            }
        }
        
        c.Next()
    }
}

// HTTP 客户端包装
func NewHTTPClient(ctx *gin.Context) *http.Client {
    return &http.Client{
        Transport: &HeaderPropagationTransport{
            Context: ctx,
        },
    }
}
```

**验证状态**：
- ✅ mock-service 已支持（`/mock-service/internal/handler/proxy.go`）
- ⏳ 生产环境 SDK 待开发

---

### Q8：出流量是否经过 LB？

**答案**：
- ❌ A 服务的出流量**不经过** A 服务的 LB
- ✅ A 服务调用 B 服务的流量**必须经过** B 服务的 LB

**流量路径**：

```
Service A Instance → Service B LB → Service B Instance
```

**实现方式**：
- 应用配置中不再包含下游服务的具体 host
- 应用通过服务发现或配置中心获取下游服务的 LB endpoint
- 所有服务间调用必须通过 LB，不允许直连实例

---

## 🎯 设计原则

- ✅ LB 是运维系统的核心
- ✅ 拓扑由 LB 配置定义
- ✅ 流量路由标识（`X-LB-Route-ID`）只有 LB 关心，程序仅透传
- ✅ 物理拓扑由 LB 路由规则实现拆分
- ✅ 使用专业 Header 命名：`X-LB-Route-ID`（而非 `x-lbid`）
- ✅ 配置即代码，所有配置纳入版本控制

---

## ❓ 待讨论问题

### 1. LB 配置的详细 Schema 设计

**待确认**：
- 是否需要支持更复杂的路由匹配条件（如基于 URL、Query 参数）？
- 是否需要支持流量镜像（Shadow Traffic）？
- 是否需要支持熔断、限流等高级功能？

### 2. 如何处理共用中间节点的场景

**场景描述**：
```
Service A (v1) ──┐
                 ├──→ Service B (共用) ──→ Service C (v1)
Service A (v2) ──┘                    └──→ Service C (v2)
```

**问题**：
- Service B 是否需要感知上游版本？
- 如何确保 Service A v2 的流量最终到达 Service C v2？

**候选方案**：
- 方案 A：Service B 透传路由标签，Service C LB 根据标签路由
- 方案 B：Service B 也拆分为两个版本（增加复杂度）

### 3. LB 配置如何与 GitOps 集成

**待确认**：
- LB 配置是否存储在独立的 Git 仓库？
- 配置变更是否触发自动部署？
- 如何处理配置冲突和合并？

### 4. 配置变更的审批流程

**待确认**：
- 是否需要多级审批？
- 审批流程如何与 Git PR 结合？
- 紧急变更如何快速生效？

### 5. 路由规则权重算法的具体实现

**待确认**：
- 使用 nginx `split_clients` 还是其他方案？
- 如何确保流量分配的精确性？
- 如何处理权重动态调整时的平滑过渡？

### 6. LB 实现选择

**候选方案**：

| LB 实现 | 优点 | 缺点 | 建议 |
|---------|------|------|------|
| **Nginx** | 成熟稳定、性能高、运维熟悉 | 动态配置能力弱 | ✅ 推荐（MVP） |
| **Envoy** | 动态配置、可观测性强 | 学习曲线陡峭 | ⏳ 后续考虑 |
| **HAProxy** | 性能极高、健康检查完善 | 社区不如 Nginx | ❌ 不推荐 |

**建议**：
- MVP 阶段使用 Nginx（团队熟悉，运维成熟）
- 通过配置文件重载实现动态更新（reload）
- 后续可考虑迁移到 Envoy（更好的动态配置能力）

---

## 📦 数据库设计（初步）

### `lb_configs` 表

```sql
CREATE TABLE lb_configs (
    id SERIAL PRIMARY KEY,
    service_name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    config_yaml TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'active',  -- active, inactive, deprecated
    created_by VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(service_name, version)
);

CREATE INDEX idx_lb_configs_service ON lb_configs(service_name);
CREATE INDEX idx_lb_configs_status ON lb_configs(status);
```

### `lb_config_history` 表（变更历史）

```sql
CREATE TABLE lb_config_history (
    id SERIAL PRIMARY KEY,
    config_id INTEGER REFERENCES lb_configs(id),
    config_yaml TEXT NOT NULL,
    change_type VARCHAR(50),  -- create, update, delete, rollback
    changed_by VARCHAR(255),
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    git_commit_sha VARCHAR(40)
);

CREATE INDEX idx_lb_config_history_config_id ON lb_config_history(config_id);
```

---

## 🔧 技术实现要点

### 后端（Go + Gin）

**核心模块**：
1. **LB Config Manager**: 配置 CRUD 操作
2. **Topology Generator**: 拓扑图生成
3. **Config Validator**: 配置验证
4. **Git Integration**: Git 版本控制集成

**API 端点**：
- `GET /api/v1/lb-configs`: 查询所有 LB 配置
- `GET /api/v1/lb-configs/:service`: 查询指定服务的配置
- `POST /api/v1/lb-configs`: 创建新配置
- `PUT /api/v1/lb-configs/:service`: 更新配置
- `DELETE /api/v1/lb-configs/:service`: 删除配置
- `GET /api/v1/topology`: 生成拓扑图数据
- `POST /api/v1/lb-configs/:service/rollback`: 回滚配置

---

### 前端（Vue 3 + AntV G6）

**核心组件**：
1. **TopologyViewer**: 拓扑图查看器
2. **TopologyEditor**: 拓扑图编辑器
3. **ConfigEditor**: YAML 配置编辑器
4. **VersionHistory**: 版本历史查看

**交互流程**：
```
用户编辑拓扑图
    ↓
实时验证配置
    ↓
生成 YAML 配置
    ↓
提交到后端
    ↓
后端验证 + Git commit
    ↓
返回结果
```

---

## 🚀 实施路线图

### Phase 1: MVP 核心功能（4 周）

- [ ] LB 配置 Schema 定义和验证
- [ ] 数据库表结构设计和创建
- [ ] 配置 CRUD API 实现
- [ ] 基础拓扑图渲染（只读）
- [ ] Nginx 配置生成器

### Phase 2: 编辑能力（3 周）

- [ ] 图形化拓扑编辑器
- [ ] 权重调整交互
- [ ] 实时配置验证
- [ ] 配置预览和对比

### Phase 3: 版本管理（2 周）

- [ ] Git 集成
- [ ] 版本历史查看
- [ ] 配置回滚
- [ ] 变更审批流程（可选）

### Phase 4: 高级功能（4 周）

- [ ] Header 透传 SDK（多语言）
- [ ] 流量监控和可视化
- [ ] A/B 测试支持
- [ ] 灰度发布自动化

---

## 📚 相关 Issue

- 主 Issue: [#3 智能发布系统](https://github.com/LiusCraft/x-mindflow/issues/3)
- 本需求 Issue: [#9 LB 流量控制与拓扑管理](https://github.com/LiusCraft/x-mindflow/issues/9)
- Mock Service: [#13 Mock Service 实现](https://github.com/LiusCraft/x-mindflow/issues/13)
- 统一发布抽象层: [#7](https://github.com/LiusCraft/x-mindflow/issues/7)
- 环境适配层: [#8](https://github.com/LiusCraft/x-mindflow/issues/8)

---

**最后更新**: 2025-10-24
