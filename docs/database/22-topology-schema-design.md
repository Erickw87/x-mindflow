# 拓扑可视化与发布管理数据库设计文档

**关联 Issue**: #22  
**文档类型**: 数据库设计  
**创建日期**: 2025-10-24  
**状态**: 设计阶段

---

## 📋 概述

本文档定义了拓扑可视化与发布管理模块的 PostgreSQL 数据库表结构设计,包括拓扑管理、节点管理、发布流程等核心实体的数据模型。

### 设计原则

1. **范式化设计**: 遵循第三范式,减少数据冗余
2. **软删除**: 核心表支持软删除,保留历史记录
3. **审计字段**: 所有表包含创建时间、更新时间、创建人等审计字段
4. **性能优化**: 合理设计索引,优化查询性能
5. **可扩展性**: 预留扩展字段,使用 JSONB 存储灵活配置

---

## 📊 ER 图

```
┌─────────────────┐
│   topologies    │
│  (拓扑树主表)    │
└────────┬────────┘
         │ 1
         │
         │ N
┌────────┴────────┐        ┌──────────────────┐
│  topology_nodes │        │ topology_edges   │
│   (节点表)       │        │    (连接边表)     │
└─────────────────┘        └──────────────────┘
         │
         │ 1
         │
         │ N
┌─────────────────┐
│ service_versions│
│  (服务版本表)    │
└─────────────────┘

┌─────────────────┐        ┌──────────────────┐
│  deployments    │   1:N  │ deployment_steps │
│  (发布任务表)    ├────────┤  (发布步骤表)     │
└─────────────────┘        └──────────────────┘

┌─────────────────┐
│ service_instances│
│  (服务实例表)    │
└─────────────────┘

┌─────────────────┐
│      users      │
│   (用户表)       │
└─────────────────┘
```

---

## 📋 表结构设计

### 1. topologies (拓扑树表)

**表描述**: 存储拓扑树的基本信息

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 拓扑树唯一标识 |
| name | VARCHAR(255) | NOT NULL, UNIQUE | 拓扑树名称 |
| description | TEXT | | 拓扑描述 |
| service_name | VARCHAR(255) | NOT NULL | 关联的服务名称 |
| status | VARCHAR(50) | NOT NULL, DEFAULT 'active' | 状态: active/inactive |
| version | INTEGER | NOT NULL, DEFAULT 1 | 拓扑版本号 |
| cloned_from | UUID | FOREIGN KEY → topologies(id) | 克隆源拓扑 ID |
| config | JSONB | | 拓扑额外配置(JSON 格式) |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 更新时间 |
| deleted_at | TIMESTAMP | | 删除时间(软删除) |
| created_by | UUID | FOREIGN KEY → users(id) | 创建人 |
| updated_by | UUID | FOREIGN KEY → users(id) | 更新人 |

**索引**:
```sql
CREATE INDEX idx_topologies_service_name ON topologies(service_name);
CREATE INDEX idx_topologies_status ON topologies(status);
CREATE INDEX idx_topologies_created_at ON topologies(created_at DESC);
CREATE INDEX idx_topologies_deleted_at ON topologies(deleted_at) WHERE deleted_at IS NULL;
```

**表创建 SQL**:
```sql
CREATE TABLE topologies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    service_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    version INTEGER NOT NULL DEFAULT 1,
    cloned_from UUID REFERENCES topologies(id) ON DELETE SET NULL,
    config JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    
    CONSTRAINT chk_status CHECK (status IN ('active', 'inactive'))
);
```

---

### 2. topology_nodes (拓扑节点表)

**表描述**: 存储拓扑树中的节点(服务节点或 LB 节点)

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 节点唯一标识 |
| topology_id | UUID | NOT NULL, FOREIGN KEY → topologies(id) | 所属拓扑树 |
| node_type | VARCHAR(50) | NOT NULL | 节点类型: service/lb |
| name | VARCHAR(255) | NOT NULL | 节点名称 |
| branch | VARCHAR(100) | DEFAULT 'default' | 分支标识(用于多版本并行) |
| total_traffic_percentage | DECIMAL(5,2) | DEFAULT 0 | 总流量占比(0-100) |
| total_instances | INTEGER | DEFAULT 0 | 总实例数(仅服务节点) |
| lb_rules | JSONB | | LB 规则配置(仅 LB 节点,JSON 格式) |
| metadata | JSONB | | 节点额外元数据 |
| position_x | DECIMAL(10,2) | | 节点 X 坐标(用于拓扑图渲染) |
| position_y | DECIMAL(10,2) | | 节点 Y 坐标 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 更新时间 |
| deleted_at | TIMESTAMP | | 删除时间(软删除) |

**索引**:
```sql
CREATE INDEX idx_topology_nodes_topology_id ON topology_nodes(topology_id);
CREATE INDEX idx_topology_nodes_type ON topology_nodes(node_type);
CREATE INDEX idx_topology_nodes_branch ON topology_nodes(branch);
CREATE INDEX idx_topology_nodes_deleted_at ON topology_nodes(deleted_at) WHERE deleted_at IS NULL;
```

**表创建 SQL**:
```sql
CREATE TABLE topology_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topology_id UUID NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    node_type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    branch VARCHAR(100) DEFAULT 'default',
    total_traffic_percentage DECIMAL(5,2) DEFAULT 0,
    total_instances INTEGER DEFAULT 0,
    lb_rules JSONB,
    metadata JSONB,
    position_x DECIMAL(10,2),
    position_y DECIMAL(10,2),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_node_type CHECK (node_type IN ('service', 'lb')),
    CONSTRAINT chk_traffic_percentage CHECK (total_traffic_percentage >= 0 AND total_traffic_percentage <= 100),
    CONSTRAINT chk_instances CHECK (total_instances >= 0)
);
```

---

### 3. service_versions (服务版本表)

**表描述**: 存储服务节点的版本信息

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 版本记录唯一标识 |
| node_id | UUID | NOT NULL, FOREIGN KEY → topology_nodes(id) | 所属节点 |
| version | VARCHAR(100) | NOT NULL | 版本号(如 v1.0, v1.1) |
| instance_count | INTEGER | NOT NULL, DEFAULT 0 | 实例数量 |
| color | VARCHAR(20) | DEFAULT '#3b82f6' | 展示颜色(十六进制) |
| is_target | BOOLEAN | DEFAULT false | 是否为目标版本 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 更新时间 |

**索引**:
```sql
CREATE INDEX idx_service_versions_node_id ON service_versions(node_id);
CREATE INDEX idx_service_versions_version ON service_versions(version);
CREATE UNIQUE INDEX idx_service_versions_node_version ON service_versions(node_id, version);
```

**表创建 SQL**:
```sql
CREATE TABLE service_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id UUID NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    version VARCHAR(100) NOT NULL,
    instance_count INTEGER NOT NULL DEFAULT 0,
    color VARCHAR(20) DEFAULT '#3b82f6',
    is_target BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_instance_count CHECK (instance_count >= 0)
);
```

---

### 4. topology_edges (拓扑连接边表)

**表描述**: 存储拓扑树中节点间的连接关系

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 边唯一标识 |
| topology_id | UUID | NOT NULL, FOREIGN KEY → topologies(id) | 所属拓扑树 |
| source_node_id | UUID | NOT NULL, FOREIGN KEY → topology_nodes(id) | 源节点 |
| target_node_id | UUID | NOT NULL, FOREIGN KEY → topology_nodes(id) | 目标节点 |
| traffic_percentage | DECIMAL(5,2) | NOT NULL, DEFAULT 0 | 流量占比(0-100) |
| label | VARCHAR(255) | | 边标签(如 "70% 流量") |
| branch | VARCHAR(100) | DEFAULT 'default' | 分支标识 |
| metadata | JSONB | | 边额外元数据 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 更新时间 |
| deleted_at | TIMESTAMP | | 删除时间(软删除) |

**索引**:
```sql
CREATE INDEX idx_topology_edges_topology_id ON topology_edges(topology_id);
CREATE INDEX idx_topology_edges_source ON topology_edges(source_node_id);
CREATE INDEX idx_topology_edges_target ON topology_edges(target_node_id);
CREATE INDEX idx_topology_edges_branch ON topology_edges(branch);
CREATE INDEX idx_topology_edges_deleted_at ON topology_edges(deleted_at) WHERE deleted_at IS NULL;
```

**表创建 SQL**:
```sql
CREATE TABLE topology_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topology_id UUID NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    source_node_id UUID NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    target_node_id UUID NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    traffic_percentage DECIMAL(5,2) NOT NULL DEFAULT 0,
    label VARCHAR(255),
    branch VARCHAR(100) DEFAULT 'default',
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_traffic_percentage CHECK (traffic_percentage >= 0 AND traffic_percentage <= 100),
    CONSTRAINT chk_different_nodes CHECK (source_node_id != target_node_id)
);
```

---

### 5. deployments (发布任务表)

**表描述**: 存储发布任务的主信息

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 发布任务唯一标识 |
| topology_id | UUID | NOT NULL, FOREIGN KEY → topologies(id) | 关联拓扑树 |
| service_name | VARCHAR(255) | NOT NULL | 服务名称 |
| old_version | VARCHAR(100) | | 旧版本号 |
| new_version | VARCHAR(100) | NOT NULL | 新版本号 |
| status | VARCHAR(50) | NOT NULL, DEFAULT 'pending' | 状态: pending/running/paused/completed/failed/rolled_back/cancelled |
| current_step | INTEGER | DEFAULT 0 | 当前执行步骤(0 表示未开始) |
| total_steps | INTEGER | NOT NULL | 总步骤数 |
| rollback_reason | TEXT | | 回滚原因 |
| config | JSONB | | 发布配置(JSON 格式) |
| started_at | TIMESTAMP | | 开始时间 |
| completed_at | TIMESTAMP | | 完成时间 |
| paused_at | TIMESTAMP | | 暂停时间 |
| rolled_back_at | TIMESTAMP | | 回滚时间 |
| cancelled_at | TIMESTAMP | | 取消时间 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 更新时间 |
| deleted_at | TIMESTAMP | | 删除时间(软删除) |
| created_by | UUID | FOREIGN KEY → users(id) | 创建人 |

**索引**:
```sql
CREATE INDEX idx_deployments_topology_id ON deployments(topology_id);
CREATE INDEX idx_deployments_service_name ON deployments(service_name);
CREATE INDEX idx_deployments_status ON deployments(status);
CREATE INDEX idx_deployments_created_at ON deployments(created_at DESC);
CREATE INDEX idx_deployments_deleted_at ON deployments(deleted_at) WHERE deleted_at IS NULL;
```

**表创建 SQL**:
```sql
CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topology_id UUID NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    service_name VARCHAR(255) NOT NULL,
    old_version VARCHAR(100),
    new_version VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    current_step INTEGER DEFAULT 0,
    total_steps INTEGER NOT NULL,
    rollback_reason TEXT,
    config JSONB,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    paused_at TIMESTAMP,
    rolled_back_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT chk_deployment_status CHECK (status IN ('pending', 'running', 'paused', 'completed', 'failed', 'rolled_back', 'cancelled')),
    CONSTRAINT chk_current_step CHECK (current_step >= 0),
    CONSTRAINT chk_total_steps CHECK (total_steps > 0)
);
```

---

### 6. deployment_steps (发布步骤表)

**表描述**: 存储发布任务的详细步骤信息

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 步骤唯一标识 |
| deployment_id | UUID | NOT NULL, FOREIGN KEY → deployments(id) | 所属发布任务 |
| step_number | INTEGER | NOT NULL | 步骤序号(从 1 开始) |
| status | VARCHAR(50) | NOT NULL, DEFAULT 'pending' | 状态: pending/running/completed/failed |
| instances_percent | DECIMAL(5,2) | NOT NULL | 实例百分比(0-100) |
| traffic_percent | DECIMAL(5,2) | NOT NULL | 流量百分比(0-100) |
| pause_type | VARCHAR(50) | NOT NULL | 暂停类型: manual/auto |
| pause_duration | INTEGER | | 自动暂停时长(秒) |
| timeout | INTEGER | | 超时时间(秒) |
| actual_instances_old | INTEGER | | 实际旧版本实例数 |
| actual_instances_new | INTEGER | | 实际新版本实例数 |
| actual_traffic_old | DECIMAL(5,2) | | 实际旧版本流量占比 |
| actual_traffic_new | DECIMAL(5,2) | | 实际新版本流量占比 |
| error_message | TEXT | | 错误信息 |
| started_at | TIMESTAMP | | 开始时间 |
| completed_at | TIMESTAMP | | 完成时间 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 更新时间 |

**索引**:
```sql
CREATE INDEX idx_deployment_steps_deployment_id ON deployment_steps(deployment_id);
CREATE INDEX idx_deployment_steps_step_number ON deployment_steps(deployment_id, step_number);
CREATE UNIQUE INDEX idx_deployment_steps_unique ON deployment_steps(deployment_id, step_number);
```

**表创建 SQL**:
```sql
CREATE TABLE deployment_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    step_number INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    instances_percent DECIMAL(5,2) NOT NULL,
    traffic_percent DECIMAL(5,2) NOT NULL,
    pause_type VARCHAR(50) NOT NULL,
    pause_duration INTEGER,
    timeout INTEGER,
    actual_instances_old INTEGER,
    actual_instances_new INTEGER,
    actual_traffic_old DECIMAL(5,2),
    actual_traffic_new DECIMAL(5,2),
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_step_status CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    CONSTRAINT chk_instances_percent CHECK (instances_percent >= 0 AND instances_percent <= 100),
    CONSTRAINT chk_traffic_percent CHECK (traffic_percent >= 0 AND traffic_percent <= 100),
    CONSTRAINT chk_pause_type CHECK (pause_type IN ('manual', 'auto')),
    CONSTRAINT chk_step_number CHECK (step_number > 0)
);
```

---

### 7. service_instances (服务实例表)

**表描述**: 存储服务实例的运行时信息

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 实例唯一标识 |
| service_name | VARCHAR(255) | NOT NULL | 服务名称 |
| version | VARCHAR(100) | NOT NULL | 版本号 |
| instance_name | VARCHAR(255) | NOT NULL, UNIQUE | 实例名称 |
| status | VARCHAR(50) | NOT NULL | 状态: running/stopped/unhealthy |
| health | VARCHAR(50) | NOT NULL, DEFAULT 'unknown' | 健康状态: healthy/unhealthy/unknown |
| ip_address | VARCHAR(45) | | IP 地址 |
| port | INTEGER | | 端口号 |
| cpu_usage | DECIMAL(5,2) | | CPU 使用率(%) |
| memory_usage | DECIMAL(5,2) | | 内存使用率(%) |
| metadata | JSONB | | 实例元数据 |
| started_at | TIMESTAMP | | 启动时间 |
| last_health_check | TIMESTAMP | | 最后健康检查时间 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 更新时间 |
| deleted_at | TIMESTAMP | | 删除时间(软删除) |

**索引**:
```sql
CREATE INDEX idx_service_instances_service_name ON service_instances(service_name);
CREATE INDEX idx_service_instances_version ON service_instances(version);
CREATE INDEX idx_service_instances_status ON service_instances(status);
CREATE INDEX idx_service_instances_health ON service_instances(health);
CREATE INDEX idx_service_instances_deleted_at ON service_instances(deleted_at) WHERE deleted_at IS NULL;
```

**表创建 SQL**:
```sql
CREATE TABLE service_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(255) NOT NULL,
    version VARCHAR(100) NOT NULL,
    instance_name VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL,
    health VARCHAR(50) NOT NULL DEFAULT 'unknown',
    ip_address VARCHAR(45),
    port INTEGER,
    cpu_usage DECIMAL(5,2),
    memory_usage DECIMAL(5,2),
    metadata JSONB,
    started_at TIMESTAMP,
    last_health_check TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_instance_status CHECK (status IN ('running', 'stopped', 'unhealthy')),
    CONSTRAINT chk_instance_health CHECK (health IN ('healthy', 'unhealthy', 'unknown')),
    CONSTRAINT chk_cpu_usage CHECK (cpu_usage >= 0 AND cpu_usage <= 100),
    CONSTRAINT chk_memory_usage CHECK (memory_usage >= 0 AND memory_usage <= 100)
);
```

---

### 8. users (用户表)

**表描述**: 存储系统用户信息

| 字段名 | 类型 | 约束 | 说明 |
|--------|------|------|------|
| id | UUID | PRIMARY KEY | 用户唯一标识 |
| username | VARCHAR(100) | NOT NULL, UNIQUE | 用户名 |
| email | VARCHAR(255) | NOT NULL, UNIQUE | 邮箱 |
| password_hash | VARCHAR(255) | NOT NULL | 密码哈希(bcrypt) |
| role | VARCHAR(50) | NOT NULL, DEFAULT 'user' | 角色: admin/user |
| status | VARCHAR(50) | NOT NULL, DEFAULT 'active' | 状态: active/inactive |
| last_login_at | TIMESTAMP | | 最后登录时间 |
| created_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 创建时间 |
| updated_at | TIMESTAMP | NOT NULL, DEFAULT NOW() | 更新时间 |
| deleted_at | TIMESTAMP | | 删除时间(软删除) |

**索引**:
```sql
CREATE UNIQUE INDEX idx_users_username ON users(username);
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;
```

**表创建 SQL**:
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_user_role CHECK (role IN ('admin', 'user')),
    CONSTRAINT chk_user_status CHECK (status IN ('active', 'inactive'))
);
```

---

## 🔄 数据库迁移脚本示例

### Migration: 创建拓扑管理表

**文件**: `migrations/001_create_topology_tables.up.sql`

```sql
-- 创建用户表
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_user_role CHECK (role IN ('admin', 'user')),
    CONSTRAINT chk_user_status CHECK (status IN ('active', 'inactive'))
);

CREATE UNIQUE INDEX idx_users_username ON users(username);
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;

-- 创建拓扑表
CREATE TABLE topologies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    service_name VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    version INTEGER NOT NULL DEFAULT 1,
    cloned_from UUID REFERENCES topologies(id) ON DELETE SET NULL,
    config JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    
    CONSTRAINT chk_status CHECK (status IN ('active', 'inactive'))
);

CREATE INDEX idx_topologies_service_name ON topologies(service_name);
CREATE INDEX idx_topologies_status ON topologies(status);
CREATE INDEX idx_topologies_created_at ON topologies(created_at DESC);
CREATE INDEX idx_topologies_deleted_at ON topologies(deleted_at) WHERE deleted_at IS NULL;

-- 创建拓扑节点表
CREATE TABLE topology_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topology_id UUID NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    node_type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    branch VARCHAR(100) DEFAULT 'default',
    total_traffic_percentage DECIMAL(5,2) DEFAULT 0,
    total_instances INTEGER DEFAULT 0,
    lb_rules JSONB,
    metadata JSONB,
    position_x DECIMAL(10,2),
    position_y DECIMAL(10,2),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_node_type CHECK (node_type IN ('service', 'lb')),
    CONSTRAINT chk_traffic_percentage CHECK (total_traffic_percentage >= 0 AND total_traffic_percentage <= 100),
    CONSTRAINT chk_instances CHECK (total_instances >= 0)
);

CREATE INDEX idx_topology_nodes_topology_id ON topology_nodes(topology_id);
CREATE INDEX idx_topology_nodes_type ON topology_nodes(node_type);
CREATE INDEX idx_topology_nodes_branch ON topology_nodes(branch);
CREATE INDEX idx_topology_nodes_deleted_at ON topology_nodes(deleted_at) WHERE deleted_at IS NULL;

-- 创建服务版本表
CREATE TABLE service_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    node_id UUID NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    version VARCHAR(100) NOT NULL,
    instance_count INTEGER NOT NULL DEFAULT 0,
    color VARCHAR(20) DEFAULT '#3b82f6',
    is_target BOOLEAN DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_instance_count CHECK (instance_count >= 0)
);

CREATE INDEX idx_service_versions_node_id ON service_versions(node_id);
CREATE INDEX idx_service_versions_version ON service_versions(version);
CREATE UNIQUE INDEX idx_service_versions_node_version ON service_versions(node_id, version);

-- 创建拓扑连接边表
CREATE TABLE topology_edges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topology_id UUID NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    source_node_id UUID NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    target_node_id UUID NOT NULL REFERENCES topology_nodes(id) ON DELETE CASCADE,
    traffic_percentage DECIMAL(5,2) NOT NULL DEFAULT 0,
    label VARCHAR(255),
    branch VARCHAR(100) DEFAULT 'default',
    metadata JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_traffic_percentage CHECK (traffic_percentage >= 0 AND traffic_percentage <= 100),
    CONSTRAINT chk_different_nodes CHECK (source_node_id != target_node_id)
);

CREATE INDEX idx_topology_edges_topology_id ON topology_edges(topology_id);
CREATE INDEX idx_topology_edges_source ON topology_edges(source_node_id);
CREATE INDEX idx_topology_edges_target ON topology_edges(target_node_id);
CREATE INDEX idx_topology_edges_branch ON topology_edges(branch);
CREATE INDEX idx_topology_edges_deleted_at ON topology_edges(deleted_at) WHERE deleted_at IS NULL;

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为所有表添加更新时间触发器
CREATE TRIGGER update_topologies_updated_at BEFORE UPDATE ON topologies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_topology_nodes_updated_at BEFORE UPDATE ON topology_nodes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_service_versions_updated_at BEFORE UPDATE ON service_versions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_topology_edges_updated_at BEFORE UPDATE ON topology_edges
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**文件**: `migrations/001_create_topology_tables.down.sql`

```sql
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_topology_edges_updated_at ON topology_edges;
DROP TRIGGER IF EXISTS update_service_versions_updated_at ON service_versions;
DROP TRIGGER IF EXISTS update_topology_nodes_updated_at ON topology_nodes;
DROP TRIGGER IF EXISTS update_topologies_updated_at ON topologies;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS topology_edges;
DROP TABLE IF EXISTS service_versions;
DROP TABLE IF EXISTS topology_nodes;
DROP TABLE IF EXISTS topologies;
DROP TABLE IF EXISTS users;
```

---

### Migration: 创建发布管理表

**文件**: `migrations/002_create_deployment_tables.up.sql`

```sql
-- 创建发布任务表
CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topology_id UUID NOT NULL REFERENCES topologies(id) ON DELETE CASCADE,
    service_name VARCHAR(255) NOT NULL,
    old_version VARCHAR(100),
    new_version VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    current_step INTEGER DEFAULT 0,
    total_steps INTEGER NOT NULL,
    rollback_reason TEXT,
    config JSONB,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    paused_at TIMESTAMP,
    rolled_back_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT chk_deployment_status CHECK (status IN ('pending', 'running', 'paused', 'completed', 'failed', 'rolled_back', 'cancelled')),
    CONSTRAINT chk_current_step CHECK (current_step >= 0),
    CONSTRAINT chk_total_steps CHECK (total_steps > 0)
);

CREATE INDEX idx_deployments_topology_id ON deployments(topology_id);
CREATE INDEX idx_deployments_service_name ON deployments(service_name);
CREATE INDEX idx_deployments_status ON deployments(status);
CREATE INDEX idx_deployments_created_at ON deployments(created_at DESC);
CREATE INDEX idx_deployments_deleted_at ON deployments(deleted_at) WHERE deleted_at IS NULL;

-- 创建发布步骤表
CREATE TABLE deployment_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deployment_id UUID NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    step_number INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    instances_percent DECIMAL(5,2) NOT NULL,
    traffic_percent DECIMAL(5,2) NOT NULL,
    pause_type VARCHAR(50) NOT NULL,
    pause_duration INTEGER,
    timeout INTEGER,
    actual_instances_old INTEGER,
    actual_instances_new INTEGER,
    actual_traffic_old DECIMAL(5,2),
    actual_traffic_new DECIMAL(5,2),
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_step_status CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    CONSTRAINT chk_instances_percent CHECK (instances_percent >= 0 AND instances_percent <= 100),
    CONSTRAINT chk_traffic_percent CHECK (traffic_percent >= 0 AND traffic_percent <= 100),
    CONSTRAINT chk_pause_type CHECK (pause_type IN ('manual', 'auto')),
    CONSTRAINT chk_step_number CHECK (step_number > 0)
);

CREATE INDEX idx_deployment_steps_deployment_id ON deployment_steps(deployment_id);
CREATE INDEX idx_deployment_steps_step_number ON deployment_steps(deployment_id, step_number);
CREATE UNIQUE INDEX idx_deployment_steps_unique ON deployment_steps(deployment_id, step_number);

-- 创建服务实例表
CREATE TABLE service_instances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(255) NOT NULL,
    version VARCHAR(100) NOT NULL,
    instance_name VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL,
    health VARCHAR(50) NOT NULL DEFAULT 'unknown',
    ip_address VARCHAR(45),
    port INTEGER,
    cpu_usage DECIMAL(5,2),
    memory_usage DECIMAL(5,2),
    metadata JSONB,
    started_at TIMESTAMP,
    last_health_check TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    CONSTRAINT chk_instance_status CHECK (status IN ('running', 'stopped', 'unhealthy')),
    CONSTRAINT chk_instance_health CHECK (health IN ('healthy', 'unhealthy', 'unknown')),
    CONSTRAINT chk_cpu_usage CHECK (cpu_usage >= 0 AND cpu_usage <= 100),
    CONSTRAINT chk_memory_usage CHECK (memory_usage >= 0 AND memory_usage <= 100)
);

CREATE INDEX idx_service_instances_service_name ON service_instances(service_name);
CREATE INDEX idx_service_instances_version ON service_instances(version);
CREATE INDEX idx_service_instances_status ON service_instances(status);
CREATE INDEX idx_service_instances_health ON service_instances(health);
CREATE INDEX idx_service_instances_deleted_at ON service_instances(deleted_at) WHERE deleted_at IS NULL;

-- 添加触发器
CREATE TRIGGER update_deployments_updated_at BEFORE UPDATE ON deployments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_deployment_steps_updated_at BEFORE UPDATE ON deployment_steps
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_service_instances_updated_at BEFORE UPDATE ON service_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**文件**: `migrations/002_create_deployment_tables.down.sql`

```sql
DROP TRIGGER IF EXISTS update_service_instances_updated_at ON service_instances;
DROP TRIGGER IF EXISTS update_deployment_steps_updated_at ON deployment_steps;
DROP TRIGGER IF EXISTS update_deployments_updated_at ON deployments;

DROP TABLE IF EXISTS service_instances;
DROP TABLE IF EXISTS deployment_steps;
DROP TABLE IF EXISTS deployments;
```

---

## 📈 性能优化建议

### 1. 索引优化

- **复合索引**: 根据常见查询模式创建复合索引
  ```sql
  CREATE INDEX idx_topologies_service_status ON topologies(service_name, status) 
  WHERE deleted_at IS NULL;
  ```

- **部分索引**: 仅索引活跃数据(未软删除的记录)
  ```sql
  CREATE INDEX idx_nodes_active ON topology_nodes(topology_id) 
  WHERE deleted_at IS NULL;
  ```

### 2. 分区策略

对于大表(如 `service_instances`, `deployment_steps`),可考虑按时间分区:

```sql
CREATE TABLE service_instances_2025_01 PARTITION OF service_instances
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

### 3. JSONB 字段优化

为 JSONB 字段的常用查询路径创建 GIN 索引:

```sql
CREATE INDEX idx_topology_nodes_lb_rules ON topology_nodes 
USING GIN (lb_rules);

CREATE INDEX idx_deployments_config ON deployments 
USING GIN (config);
```

### 4. 查询优化

- 使用 `EXPLAIN ANALYZE` 分析慢查询
- 避免 `SELECT *`,明确指定需要的列
- 合理使用 JOIN,避免 N+1 查询

---

## 🔒 数据完整性约束

### 外键约束

所有外键关系都已在表定义中声明,确保引用完整性。

### 级联删除策略

- **CASCADE**: 拓扑删除时级联删除节点和边
- **SET NULL**: 拓扑克隆源删除时设置为 NULL

### Check 约束

- 百分比字段限制在 0-100 之间
- 状态字段限制为预定义的枚举值
- 实例数、步骤数等必须为非负数

---

## 📚 相关文档

- **API 设计**: `docs/api/22-topology-api-design.md`
- **架构设计**: `docs/architecture/22-topology-architecture-design.md`
- **模块设计**: `docs/modules/22-topology-module-design.md`
- **原型文档**: `docs/prototypes/11-topology-visualization-prototype.md`

---

**文档版本**: v1.0  
**最后更新**: 2025-10-24  
**维护人**: xgopilot (Claude Code AI Assistant)
