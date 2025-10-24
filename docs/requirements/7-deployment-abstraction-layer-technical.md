# 统一发布抽象层 - 技术实现文档

> 本文档包含统一发布抽象层的详细技术设计和实现方案,面向开发团队。
> 
> 产品需求文档请参考: [7-deployment-abstraction-layer.md](./7-deployment-abstraction-layer.md)

## 需求讨论总结

本文档基于 Issue #7 的需求讨论结果编写,所有核心设计决策已经过确认。

### 讨论参与者

- @Erickw87 (产品负责人)
- @xgopilot (技术负责人)

## 核心设计决策

### 1. 资源对象模型命名

**决策**: 抽象层保持 `Instance` 命名(不使用 `Pod`)

**理由**:
- 维护环境无关的抽象层设计原则
- K8s 适配器内部映射 `Instance` ↔ `Pod`
- 支持未来多环境扩展(裸金属、虚拟机等)

**术语映射关系**:

| 抽象层概念 | K8s 对应概念 | BareMetal 对应概念 |
|-----------|-------------|-------------------|
| Instance  | Pod         | Process/Container |
| Deployment | Deployment | Deployment(通过适配器实现) |
| Service   | Service     | Load Balancer Config |

**BareMetal Deployment 实现方案**:

抽象层的 `Deployment` 概念在 BareMetal 环境中通过适配器映射到具体实现:

**方案 : Supervisor**

```ini
# /etc/supervisor/conf.d/myapp.conf
[program:myapp]
command=/usr/local/bin/myapp
numprocs=3
process_name=%(program_name)s_%(process_num)02d
autostart=true
autorestart=true
```

---

### 2. 发布策略

**决策**: 统一为**渐进式发布(Progressive Deployment)**

**变更内容**:
- 移除 Rolling / Blue-Green / Canary 三种独立策略
- 通过配置参数控制发布比例和节奏

**配置示例**:

```yaml
deployment:
  strategy:
    type: Progressive
    steps:
      - replicas: 10%    # 第一批 10%
        pause: 5m
        health_check_required: true
      - replicas: 50%    # 第二批 50%
        pause: 10m
        health_check_required: true
      - replicas: 100%   # 全量发布
        health_check_required: true
```

**优势**:
- 降低实现复杂度,减少状态机分支
- 提高发布灵活性,用户可自定义阶梯
- 统一状态机逻辑,降低维护成本

**实现要点**:
- 发布控制器维护一个 `ProgressiveDeployment` 策略
- 支持配置 `steps`(阶段列表)
- 每个阶段包含: 目标副本比例 + 暂停时长 + 健康检查条件

---

### 3. 版本历史保留策略

**决策**: 保留最近 **30 个版本**(FIFO 淘汰)

**实现要点**:
- 所有历史版本都保留(最大 30 个)
- 超过 30 个时,自动删除最旧版本
- 正在运行的版本不计入 30 个限制(永久保留)

**存储位置**: PostgreSQL 数据库

**数据模型**:

```go
type DeploymentHistory struct {
    ID        uint            `gorm:"primarykey"`
    App       string          `gorm:"index;not null"` // 服务名称
    Version   string          `gorm:"not null"`
    Template  json.RawMessage `gorm:"type:jsonb"` // 发布模板快照
    Status    string          `gorm:"not null"`   // running/stopped/unstable
    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt  `gorm:"index"`
}
```

**版本淘汰逻辑**:

```go
func SaveDeployment(deployment *Deployment) error {
    // 1. 保存新版本
    history := &DeploymentHistory{
        App:      deployment.App,
        Version:  deployment.Version,
        Template: deployment.Template,
        Status:   "running",
    }
    if err := db.Create(history).Error; err != nil {
        return err
    }
    
    // 2. 检查版本数量(不包含正在运行的版本)
    var count int64
    db.Model(&DeploymentHistory{}).
        Where("app = ? AND status != ?", deployment.App, "running").
        Count(&count)
    
    // 3. 超过 30 个时删除最旧版本
    if count > 30 {
        var oldVersions []DeploymentHistory
        db.Where("app = ? AND status != ?", deployment.App, "running").
            Order("created_at ASC").
            Limit(int(count - 30)).
            Find(&oldVersions)
        
        for _, v := range oldVersions {
            db.Delete(&v)
        }
    }
    
    return nil
}
```

---

### 4. 自动回滚机制

**决策**: **支持自动回滚**(而非 MVP 阶段手动回滚)

**触发条件**:

| 触发场景 | 检测方式 | 阈值示例 |
|---------|---------|---------|
| 启动失败 | 实例状态检查 | 连续 3 次启动失败 |
| 健康检查失败 | HTTP/TCP 探活 | 失败率 > 50% 持续 2 分钟 |
| CPU 过载 | Prometheus 指标 | CPU > 90% 持续 5 分钟 |
| 内存溢出 | Prometheus 指标 | Memory > 95% 持续 3 分钟 |

**监控配置示例**:

```yaml
monitoring:
  health_check:
    type: http
    path: /health
    interval: 10s
    timeout: 3s
    unhealthy_threshold: 3
    
  metrics:
    - name: cpu_usage
      query: 'rate(container_cpu_usage_seconds_total[5m])'
      threshold: 0.9
      duration: 5m
      
    - name: memory_usage
      query: 'container_memory_usage_bytes / container_spec_memory_limit_bytes'
      threshold: 0.95
      duration: 3m
```

**回滚决策流程**:

```
新版本部署中
    ↓
健康检查 + 指标监控(持续)
    ↓
[ 检测到异常 ]
    ↓
触发回滚决策
    ├─ 计算异常持续时长
    ├─ 验证阈值条件
    └─ 确认上一版本可用
    ↓
执行自动回滚
    ├─ 暂停当前发布
    ├─ 调用回滚指令
    └─ 通知运维人员
```

**自动回滚控制器(伪代码)**:

```go
type AutoRollbackController struct {
    monitoringService *MonitoringService
    deploymentRepo    *DeploymentRepository
    notificationSvc   *NotificationService
}

func (c *AutoRollbackController) WatchDeployment(deploymentID uint) {
    ticker := time.NewTicker(10 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        deployment := c.deploymentRepo.GetByID(deploymentID)
        if deployment.Status != "Deploying" {
            return
        }
        
        issues := []string{}
        
        // 启动失败检查
        if c.monitoringService.GetStartupFailureCount(deploymentID) >= 3 {
            issues = append(issues, "连续启动失败超过 3 次")
        }
        
        // 健康检查
        healthStatus := c.monitoringService.GetHealthCheckStatus(deploymentID)
        if healthStatus.FailureRate > 0.5 && healthStatus.Duration > 2*time.Minute {
            issues = append(issues, "健康检查失败率 > 50% 持续 2 分钟")
        }
        
        // Prometheus 指标
        metrics := c.monitoringService.QueryMetrics(deploymentID)
        if metrics.CPU > 0.9 && metrics.CPUDuration > 5*time.Minute {
            issues = append(issues, "CPU 使用率 > 90% 持续 5 分钟")
        }
        if metrics.Memory > 0.95 && metrics.MemoryDuration > 3*time.Minute {
            issues = append(issues, "内存使用率 > 95% 持续 3 分钟")
        }
        
        if len(issues) > 0 {
            log.Warn("检测到异常,触发自动回滚", zap.Strings("reasons", issues))
            c.TriggerRollback(deployment, issues)
            return
        }
    }
}

func (c *AutoRollbackController) TriggerRollback(
    deployment *Deployment,
    reasons []string,
) error {
    previousVersion := c.deploymentRepo.GetPreviousStableVersion(deployment.App)
    if previousVersion == nil {
        return errors.New("无可用的上一版本")
    }
    
    rollbackDeployment := &Deployment{
        App:      deployment.App,
        Version:  previousVersion.Version,
        Template: previousVersion.Template,
        Reason:   fmt.Sprintf("自动回滚: %s", strings.Join(reasons, "; ")),
    }
    
    // 通知运维
    c.notificationSvc.Send(&Notification{
        Level:   "critical",
        Title:   "自动回滚已触发",
        Content: fmt.Sprintf("应用 %s 从版本 %s 回滚到 %s\n原因: %s",
            deployment.AppName,
            deployment.Version,
            previousVersion.Version,
            strings.Join(reasons, "\n- ")),
    })
    
    return c.deploymentRepo.Create(rollbackDeployment)
}
```

**安全机制**:

为防止"回滚循环"(新版本故障 → 回滚 → 再次部署 → 再次回滚):

1. **冷却期**: 自动回滚后 30 分钟内,禁止自动触发同一应用的回滚
2. **人工确认**: 连续 3 次自动回滚后,要求人工介入
3. **版本黑名单**: 自动回滚的版本标记为"不稳定",禁止再次自动部署

---

### 5. 多集群支持

**决策**: MVP 阶段**仅支持单集群部署**

**后续扩展**: 
- 多集群部署能力作为后续迭代功能
- 当前架构设计预留扩展性(通过 Environment/Cluster 抽象)

---

### 6. 配置存储方案

**决策**: **PostgreSQL 数据库**

**存储内容**:
- 应用配置(Application)
- 部署模板(Deployment Template)
- 版本历史(Deployment History)
- 发布策略配置(Strategy Config)
- 健康检查配置(HealthCheck Config)

**优势**:
- 查询性能更好(相比 Git)
- 支持复杂关联查询
- 事务一致性保证

---

### 7. Prometheus 集成方式

**决策**: **抽象层直接查询 Prometheus API**

**实现要点**:

```go
type MonitoringService struct {
    prometheusClient *prometheus.Client
}

func (s *MonitoringService) QueryMetrics(deploymentID uint) (*Metrics, error) {
    query := fmt.Sprintf(
        `rate(container_cpu_usage_seconds_total{deployment_id="%d"}[5m])`,
        deploymentID,
    )
    result, err := s.prometheusClient.Query(query)
    if err != nil {
        return nil, err
    }
    
    return &Metrics{
        CPU:         result.Value,
        CPUDuration: result.Duration,
    }, nil
}
```

**优势**:
- 统一的监控数据源
- 减少适配器复杂度
- 支持复杂的 PromQL 查询

**注意事项**:
- 需要确保 Prometheus 能够采集各环境的指标(K8s 和 BareMetal)
- BareMetal 环境需要部署 Node Exporter / cAdvisor

---

### 8. 健康检查配置

**决策**: 
- **存储位置**: PostgreSQL 数据库
- **自定义脚本**: 参考 K8s 的 `livenessProbe` / `readinessProbe` 机制

**数据模型设计**:

```go
type HealthCheckConfig struct {
    ID           uint   `gorm:"primarykey"`
    DeploymentID uint   `gorm:"index"`
    
    // 健康检查类型: http / tcp / exec
    Type         string `gorm:"not null"`
    
    // HTTP 检查配置
    HTTPPath     string `gorm:"column:http_path"`
    HTTPPort     int    `gorm:"column:http_port"`
    HTTPHeaders  string `gorm:"type:jsonb"` // JSON 格式
    
    // TCP 检查配置
    TCPPort      int    `gorm:"column:tcp_port"`
    
    // 自定义脚本检查
    ExecCommand  string `gorm:"column:exec_command"` // 命令行
    
    // 通用配置
    InitialDelaySeconds int `gorm:"default:10"`
    TimeoutSeconds      int `gorm:"default:3"`
    PeriodSeconds       int `gorm:"default:10"`
    SuccessThreshold    int `gorm:"default:1"`
    FailureThreshold    int `gorm:"default:3"`
}
```

**配置示例**:

```yaml
# HTTP 健康检查
health_check:
  type: http
  http_path: /health
  http_port: 8080
  http_headers:
    User-Agent: HealthChecker/1.0
  timeout_seconds: 3
  period_seconds: 10
  failure_threshold: 3

# TCP 健康检查
health_check:
  type: tcp
  tcp_port: 3306
  timeout_seconds: 5

# 自定义脚本检查(类似 K8s exec probe)
health_check:
  type: exec
  exec_command: "/bin/sh -c 'pg_isready -U postgres'"
  timeout_seconds: 5
  failure_threshold: 3
```

**K8s 对照**:

| 抽象层配置 | K8s livenessProbe | 说明 |
|-----------|-------------------|------|
| `type: http` | `httpGet` | HTTP GET 请求 |
| `type: tcp` | `tcpSocket` | TCP 端口检查 |
| `type: exec` | `exec.command` | 执行自定义脚本 |
| `failure_threshold` | `failureThreshold` | 失败次数阈值 |
| `timeout_seconds` | `timeoutSeconds` | 超时时间 |

**安全限制**:
- `exec` 类型的脚本需要白名单机制(防止恶意命令注入)
- 脚本执行超时强制终止
- 限制脚本可访问的资源(沙箱环境)

---

### 9. 通知机制

**决策**: **自动回滚时通过企业微信通知运维人员**

**实现架构**:

```go
type NotificationService interface {
    Send(notification *Notification) error
}

type WeChatWorkNotifier struct {
    webhookURL string
    httpClient *http.Client
}

func (w *WeChatWorkNotifier) Send(notification *Notification) error {
    message := map[string]interface{}{
        "msgtype": "markdown",
        "markdown": map[string]string{
            "content": fmt.Sprintf(`
### 🚨 %s

**应用**: %s
**版本**: %s → %s
**时间**: %s
**原因**:
%s

[查看详情](%s)
            `,
                notification.Title,
                notification.AppName,
                notification.FromVersion,
                notification.ToVersion,
                time.Now().Format("2006-01-02 15:04:05"),
                notification.Reason,
                notification.DetailURL,
            ),
        },
    }
    
    body, _ := json.Marshal(message)
    resp, err := w.httpClient.Post(w.webhookURL, "application/json", bytes.NewBuffer(body))
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    return nil
}
```

**企业微信 Webhook 配置**:

```yaml
# config/notification.yaml
wechat_work:
  webhook_url: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=YOUR_KEY"
  mention_users:
    - "@all"  # 通知所有人
    - "zhangsan"
    - "lisi"
```

**通知场景**:
1. 自动回滚触发
2. 部署失败
3. 健康检查持续失败
4. 版本发布完成

**消息示例**:

```markdown
### 🚨 自动回滚已触发

**应用**: x-mindflow-backend
**版本**: v1.2.3 → v1.2.2
**时间**: 2025-10-24 15:30:45
**原因**:
- CPU 使用率 > 90% 持续 5 分钟
- 内存使用率 > 95% 持续 3 分钟

[查看详情](https://dashboard.example.com/deployments/123)

@all @zhangsan
```

---

## 统一资源对象模型

### 环境无关的抽象设计

| 抽象层概念 | 语义 | K8s 实现 | BareMetal 实现 |
|-----------|------|---------|---------------|
| **Application** | 逻辑应用(多个 Deployment 的集合) | Namespace + Labels | 配置组(数据库记录) |
| **Deployment** | 部署单元(版本 + 副本数 + 策略) | Deployment 资源 | Systemd/Supervisor/Docker |
| **Service** | 逻辑服务名(流量入口) | Service 资源 | Nginx/HAProxy 配置 |
| **Instance** | 运行实例(单个副本) | Pod | Process/Container |
| **Environment** | 环境抽象 | K8s Cluster | BareMetal Cluster |

---

## 待进一步讨论的问题

### 发布状态机的状态流转

**当前状态**: 尚未确定最终方案

**下一步**: 需要进一步展开讨论以下内容:
- 状态节点定义(Creating / Progressing / Paused / Completed / Failed / RollingBack 等)
- 状态转换条件
- 异常状态处理逻辑
- 状态持久化与恢复机制

**建议**: 在模块设计阶段通过状态图或状态转换表明确定义

---

## 架构设计核心原则

- ✅ 抽象层只负责状态模型、控制逻辑与任务派发
- ✅ 不关心 Pod、Node、容器、IP、进程等执行层细节
- ✅ 通过统一接口与环境适配层通信
- ✅ 配置存储于 PostgreSQL
- ✅ MVP 阶段专注单集群场景

---

## 下一步行动建议

基于以上决策,建议按以下顺序推进:

### 1. 创建模块设计 Issue

产出以下文档:
- API 设计文档(`/docs/api/`)
  - 发布 API 端点定义
  - 版本管理 API
  - 健康检查配置 API
- 数据模型设计(`/docs/database/`)
  - DeploymentHistory 表结构
  - HealthCheckConfig 表结构
  - NotificationConfig 表结构
- 发布状态机详细设计(状态图 + 转换表)
- 自动回滚逻辑设计(Prometheus 查询方案)
- BareMetal 适配器 Deployment 实现方案

### 2. 技术预研

- 验证 Prometheus API 查询性能(大量 Deployment 场景)
- 调研 BareMetal 环境监控方案(Node Exporter + cAdvisor)
- 测试企业微信 Webhook 集成
- 验证 Systemd/Supervisor 作为 BareMetal Deployment 实现的可行性

### 3. 原型验证

- 小规模测试渐进式发布逻辑
- 验证自动回滚触发机制的可靠性
- 测试版本历史 FIFO 淘汰逻辑
- 验证健康检查 exec 类型的安全性

### 4. 开发实现

- 基于模块设计创建开发任务 Sub-Issues
- 按模块并行开发:
  - 资源模型
  - 发布控制器
  - 状态聚合器
  - 监控服务
  - 通知服务

---

## 相关 Issue

- 主 Issue: [#3 智能发布系统](https://github.com/LiusCraft/x-mindflow/issues/3)
- 需求讨论 Issue: [#7 需求讨论:统一发布抽象层设计](https://github.com/LiusCraft/x-mindflow/issues/7)

---

**最后更新**: 2025-10-24
**文档版本**: v1.0
