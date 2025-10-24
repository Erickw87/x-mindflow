# Mock Service 模块设计文档

**Issue**: #13  
**版本**: v1.0  
**日期**: 2025-10-24  
**状态**: 设计阶段

---

## 📌 概述

本文档详细描述 Mock Service 的模块划分、接口定义、数据流和模块间交互关系。

---

## 🎯 模块划分

```
mock-service/
├── cmd/                          # 应用入口
├── internal/                     # 私有业务逻辑
│   ├── server/                   # HTTP 服务器模块
│   ├── handler/                  # 请求处理器模块
│   ├── behavior/                 # 行为模拟器模块
│   └── template/                 # 模板渲染模块
├── pkg/                          # 可复用公共库
│   ├── config/                   # 配置管理模块
│   └── logger/                   # 日志模块
```

---

## 📦 模块详细设计

### 1. Config 模块 (`pkg/config`)

**职责**: 配置文件加载、解析、验证

#### 1.1 接口定义

```go
package config

import (
    "time"
)

// Loader 配置加载器接口
type Loader interface {
    // Load 从文件加载配置
    Load(path string) (*ServiceConfig, error)
    
    // Validate 验证配置有效性
    Validate(config *ServiceConfig) error
}

// Watcher 配置监听器接口（可选，未来扩展）
type Watcher interface {
    // Watch 监听配置文件变更
    Watch(path string, callback func(*ServiceConfig)) error
    
    // Stop 停止监听
    Stop() error
}
```

#### 1.2 数据结构

```go
// ServiceConfig 服务完整配置
type ServiceConfig struct {
    Service       ServiceInfo         `yaml:"service" validate:"required"`
    Behaviors     BehaviorConfig      `yaml:"behaviors"`
    Endpoints     []EndpointConfig    `yaml:"endpoints" validate:"required,dive"`
    VersionBehaviors map[string]VersionBehavior `yaml:"version_behaviors"`
    Observability ObservabilityConfig `yaml:"observability"`
}

// ServiceInfo 服务基本信息
type ServiceInfo struct {
    Name    string `yaml:"name" validate:"required"`
    Version string `yaml:"version" validate:"required,semver"`
    Port    int    `yaml:"port" validate:"required,min=1,max=65535"`
}

// BehaviorConfig 行为模拟配置
type BehaviorConfig struct {
    CPU     CPUBehavior     `yaml:"cpu"`
    Memory  MemoryBehavior  `yaml:"memory"`
    Startup StartupBehavior `yaml:"startup"`
    Crash   CrashBehavior   `yaml:"crash"`
}

// CPUBehavior CPU 行为配置
type CPUBehavior struct {
    Enabled       bool          `yaml:"enabled"`
    TargetPercent int           `yaml:"target_percent" validate:"min=0,max=100"`
    Trigger       TriggerConfig `yaml:"trigger"`
    Duration      time.Duration `yaml:"duration" validate:"required"`
}

// MemoryBehavior 内存行为配置
type MemoryBehavior struct {
    Enabled  bool          `yaml:"enabled"`
    TargetMB int           `yaml:"target_mb" validate:"min=1"`
    Trigger  TriggerConfig `yaml:"trigger"`
    Duration time.Duration `yaml:"duration" validate:"required"`
}

// TriggerConfig 触发器配置
type TriggerConfig struct {
    Type     string `yaml:"type" validate:"required,oneof=startup endpoint periodic"`
    Endpoint string `yaml:"endpoint"`
}

// StartupBehavior 启动行为配置
type StartupBehavior struct {
    Fail     bool          `yaml:"fail"`
    Delay    time.Duration `yaml:"delay"`
    ExitCode int           `yaml:"exit_code" validate:"min=0,max=255"`
}

// CrashBehavior 崩溃行为配置
type CrashBehavior struct {
    Enabled bool          `yaml:"enabled"`
    Trigger TriggerConfig `yaml:"trigger"`
    Delay   time.Duration `yaml:"delay"`
}

// EndpointConfig 端点配置
type EndpointConfig struct {
    Path      string          `yaml:"path" validate:"required,startswith=/"`
    Method    string          `yaml:"method" validate:"required,oneof=GET POST PUT DELETE PATCH"`
    Response  *ResponseConfig `yaml:"response"`
    Proxy     *ProxyConfig    `yaml:"proxy"`
    Responses []ResponseConfig `yaml:"responses"`
    ErrorRate float64         `yaml:"error_rate" validate:"min=0,max=1"`
    Latency   time.Duration   `yaml:"latency"`
}

// ResponseConfig 响应配置
type ResponseConfig struct {
    Status  int                    `yaml:"status" validate:"required,min=100,max=599"`
    Headers map[string]string      `yaml:"headers"`
    Body    interface{}            `yaml:"body"`
    Weight  int                    `yaml:"weight"`
}

// ProxyConfig 代理配置
type ProxyConfig struct {
    Target          string            `yaml:"target" validate:"required,url"`
    Path            string            `yaml:"path"`
    PreserveHeaders bool              `yaml:"preserve_headers"`
    Timeout         time.Duration     `yaml:"timeout"`
    QueryParams     map[string]string `yaml:"query_params"`
}

// ObservabilityConfig 可观测性配置
type ObservabilityConfig struct {
    HealthCheck HealthCheckConfig `yaml:"health_check"`
    Logging     LoggingConfig     `yaml:"logging"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
    Enabled bool   `yaml:"enabled"`
    Port    int    `yaml:"port"`
    Path    string `yaml:"path"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
    Level    string `yaml:"level" validate:"oneof=debug info warn error"`
    Format   string `yaml:"format" validate:"oneof=json console"`
    Output   string `yaml:"output" validate:"oneof=stdout file"`
    FilePath string `yaml:"file_path"`
}
```

#### 1.3 实现要点

```go
package config

import (
    "fmt"
    "os"
    "gopkg.in/yaml.v3"
    "github.com/go-playground/validator/v10"
)

type defaultLoader struct {
    validator *validator.Validate
}

func NewLoader() Loader {
    return &defaultLoader{
        validator: validator.New(),
    }
}

func (l *defaultLoader) Load(path string) (*ServiceConfig, error) {
    // 读取文件
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    // 解析 YAML
    var config ServiceConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse YAML: %w", err)
    }
    
    // 验证配置
    if err := l.Validate(&config); err != nil {
        return nil, err
    }
    
    return &config, nil
}

func (l *defaultLoader) Validate(config *ServiceConfig) error {
    // 结构验证
    if err := l.validator.Struct(config); err != nil {
        return fmt.Errorf("config validation failed: %w", err)
    }
    
    // 业务逻辑验证
    if err := l.validateBusinessRules(config); err != nil {
        return err
    }
    
    return nil
}

func (l *defaultLoader) validateBusinessRules(config *ServiceConfig) error {
    // 验证端点配置: response 和 proxy 二选一
    for i, ep := range config.Endpoints {
        if ep.Response == nil && ep.Proxy == nil && len(ep.Responses) == 0 {
            return fmt.Errorf("endpoint[%d]: must have either response, proxy, or responses", i)
        }
        
        // 验证随机响应权重总和
        if len(ep.Responses) > 0 {
            totalWeight := 0
            for _, resp := range ep.Responses {
                totalWeight += resp.Weight
            }
            if totalWeight != 100 {
                return fmt.Errorf("endpoint[%d]: response weights must sum to 100, got %d", i, totalWeight)
            }
        }
    }
    
    // 验证 CPU 行为触发器配置
    if config.Behaviors.CPU.Enabled {
        if config.Behaviors.CPU.Trigger.Type == "endpoint" && config.Behaviors.CPU.Trigger.Endpoint == "" {
            return fmt.Errorf("cpu behavior: endpoint trigger requires endpoint path")
        }
    }
    
    // 验证 Memory 行为触发器配置
    if config.Behaviors.Memory.Enabled {
        if config.Behaviors.Memory.Trigger.Type == "endpoint" && config.Behaviors.Memory.Trigger.Endpoint == "" {
            return fmt.Errorf("memory behavior: endpoint trigger requires endpoint path")
        }
    }
    
    // 验证 Crash 行为触发器配置
    if config.Behaviors.Crash.Enabled {
        if config.Behaviors.Crash.Trigger.Type == "endpoint" && config.Behaviors.Crash.Trigger.Endpoint == "" {
            return fmt.Errorf("crash behavior: endpoint trigger requires endpoint path")
        }
    }
    
    return nil
}
```

#### 1.4 错误处理

```go
var (
    ErrConfigNotFound    = errors.New("config file not found")
    ErrInvalidYAML       = errors.New("invalid YAML syntax")
    ErrValidationFailed  = errors.New("validation failed")
)
```

---

### 2. Server 模块 (`internal/server`)

**职责**: HTTP 服务器生命周期管理、路由注册、中间件

#### 2.1 接口定义

```go
package server

import (
    "context"
    "github.com/gin-gonic/gin"
    "mock-service/pkg/config"
)

// Server HTTP 服务器接口
type Server interface {
    // Start 启动服务器（阻塞）
    Start() error
    
    // Shutdown 优雅关闭
    Shutdown(ctx context.Context) error
    
    // RegisterEndpoints 注册端点
    RegisterEndpoints(endpoints []config.EndpointConfig) error
}

// Middleware 中间件接口
type Middleware interface {
    Handle() gin.HandlerFunc
}
```

#### 2.2 实现

```go
package server

import (
    "context"
    "fmt"
    "net/http"
    "github.com/gin-gonic/gin"
    "mock-service/internal/handler"
    "mock-service/internal/server/middleware"
    "mock-service/pkg/config"
    "mock-service/pkg/logger"
)

type mockServer struct {
    engine   *gin.Engine
    config   *config.ServiceConfig
    handler  handler.Handler
    server   *http.Server
    logger   *logger.Logger
}

func New(cfg *config.ServiceConfig, h handler.Handler, log *logger.Logger) Server {
    return &mockServer{
        config:  cfg,
        handler: h,
        logger:  log,
    }
}

func (s *mockServer) Start() error {
    // 初始化 Gin
    gin.SetMode(gin.ReleaseMode)
    s.engine = gin.New()
    
    // 注册中间件
    s.engine.Use(
        middleware.RequestID(),
        middleware.Logger(s.logger),
        middleware.Recovery(s.logger),
    )
    
    // 注册健康检查
    s.engine.GET("/health", s.healthCheckHandler)
    
    // 注册业务端点
    if err := s.RegisterEndpoints(s.config.Endpoints); err != nil {
        return fmt.Errorf("failed to register endpoints: %w", err)
    }
    
    // 启动 HTTP 服务器
    addr := fmt.Sprintf(":%d", s.config.Service.Port)
    s.server = &http.Server{
        Addr:    addr,
        Handler: s.engine,
    }
    
    s.logger.Info("Server starting", "address", addr)
    
    if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        return fmt.Errorf("server failed: %w", err)
    }
    
    return nil
}

func (s *mockServer) Shutdown(ctx context.Context) error {
    s.logger.Info("Server shutting down")
    return s.server.Shutdown(ctx)
}

func (s *mockServer) RegisterEndpoints(endpoints []config.EndpointConfig) error {
    for _, ep := range endpoints {
        handlerFunc := s.createHandler(ep)
        s.engine.Handle(ep.Method, ep.Path, handlerFunc)
        s.logger.Debug("Registered endpoint", "method", ep.Method, "path", ep.Path)
    }
    return nil
}

func (s *mockServer) createHandler(ep config.EndpointConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 应用延迟
        if ep.Latency > 0 {
            time.Sleep(ep.Latency)
        }
        
        // 错误注入
        if ep.ErrorRate > 0 && rand.Float64() < ep.ErrorRate {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "injected error"})
            return
        }
        
        // 路由到具体处理器
        if ep.Proxy != nil {
            s.handler.HandleProxy(c, ep.Proxy)
        } else if len(ep.Responses) > 0 {
            s.handler.HandleRandomResponse(c, ep.Responses)
        } else {
            s.handler.HandleStaticResponse(c, ep.Response)
        }
    }
}

func (s *mockServer) healthCheckHandler(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":    "healthy",
        "version":   s.config.Service.Version,
        "timestamp": time.Now().Format(time.RFC3339),
    })
}
```

#### 2.3 中间件设计

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "mock-service/pkg/logger"
    "time"
)

// RequestID 生成请求 ID
func RequestID() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        c.Next()
    }
}

// Logger 结构化日志中间件
func Logger(log *logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        
        c.Next()
        
        latency := time.Since(start)
        status := c.Writer.Status()
        
        log.Info("request completed",
            "request_id", c.GetString("request_id"),
            "method", c.Request.Method,
            "path", path,
            "status", status,
            "latency", latency.String(),
        )
    }
}

// Recovery 恢复 panic
func Recovery(log *logger.Logger) gin.HandlerFunc {
    return gin.CustomRecovery(func(c *gin.Context, err interface{}) {
        log.Error("panic recovered",
            "request_id", c.GetString("request_id"),
            "error", err,
        )
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Internal Server Error",
        })
    })
}
```

---

### 3. Handler 模块 (`internal/handler`)

**职责**: 请求处理、响应生成、代理转发

#### 3.1 接口定义

```go
package handler

import (
    "github.com/gin-gonic/gin"
    "mock-service/pkg/config"
)

// Handler 请求处理器接口
type Handler interface {
    // HandleStaticResponse 处理静态响应
    HandleStaticResponse(c *gin.Context, cfg *config.ResponseConfig)
    
    // HandleProxy 处理代理请求
    HandleProxy(c *gin.Context, cfg *config.ProxyConfig)
    
    // HandleRandomResponse 处理随机响应
    HandleRandomResponse(c *gin.Context, responses []config.ResponseConfig)
}
```

#### 3.2 静态响应处理器

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "mock-service/internal/template"
    "mock-service/pkg/config"
)

type staticHandler struct {
    renderer template.Renderer
}

func NewStaticHandler(renderer template.Renderer) *staticHandler {
    return &staticHandler{
        renderer: renderer,
    }
}

func (h *staticHandler) Handle(c *gin.Context, cfg *config.ResponseConfig) {
    // 设置 Headers
    for key, value := range cfg.Headers {
        c.Header(key, value)
    }
    
    // 渲染响应体
    body, err := h.renderer.Render(cfg.Body, c)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to render response"})
        return
    }
    
    c.JSON(cfg.Status, body)
}
```

#### 3.3 代理处理器

```go
package handler

import (
    "bytes"
    "io"
    "net/http"
    "strings"
    "time"
    "github.com/gin-gonic/gin"
    "mock-service/pkg/config"
)

type proxyHandler struct{}

func NewProxyHandler() *proxyHandler {
    return &proxyHandler{}
}

func (h *proxyHandler) Handle(c *gin.Context, cfg *config.ProxyConfig) {
    // 构建目标 URL
    targetURL := cfg.Target
    if cfg.Path != "" {
        targetURL = cfg.Target + cfg.Path
    } else {
        targetURL = cfg.Target + c.Request.URL.Path
    }
    
    // 添加查询参数
    if len(cfg.QueryParams) > 0 {
        query := c.Request.URL.Query()
        for k, v := range cfg.QueryParams {
            query.Set(k, v)
        }
        targetURL += "?" + query.Encode()
    }
    
    // 创建新请求
    var bodyReader io.Reader
    if c.Request.Body != nil {
        bodyBytes, _ := io.ReadAll(c.Request.Body)
        bodyReader = bytes.NewReader(bodyBytes)
    }
    
    req, err := http.NewRequest(c.Request.Method, targetURL, bodyReader)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create proxy request"})
        return
    }
    
    // 复制 Headers
    h.copyHeaders(c.Request.Header, req.Header, cfg.PreserveHeaders)
    
    // 为每个请求创建独立的 client 以避免并发修改超时配置
    timeout := cfg.Timeout
    if timeout == 0 {
        timeout = 30 * time.Second
    }
    
    client := &http.Client{
        Timeout: timeout,
    }
    
    // 发送请求
    resp, err := client.Do(req)
    if err != nil {
        c.JSON(http.StatusBadGateway, gin.H{"error": "proxy request failed"})
        return
    }
    defer resp.Body.Close()
    
    // 复制响应 Headers
    for key, values := range resp.Header {
        for _, value := range values {
            c.Header(key, value)
        }
    }
    
    // 复制响应体
    c.Status(resp.StatusCode)
    io.Copy(c.Writer, resp.Body)
}

func (h *proxyHandler) copyHeaders(src, dst http.Header, preserveAll bool) {
    // 始终透传的 Headers
    mustCopy := []string{
        "X-Request-ID", 
        "X-Traffic-Tag", 
        "X-Forwarded-For",
        "Content-Type",
        "Authorization",
        "User-Agent",
    }
    for _, key := range mustCopy {
        if value := src.Get(key); value != "" {
            dst.Set(key, value)
        }
    }
    
    // 可选: 透传所有 X- 开头的 Headers
    if preserveAll {
        for key, values := range src {
            if strings.HasPrefix(key, "X-") {
                for _, value := range values {
                    dst.Add(key, value)
                }
            }
        }
    }
}
```

---

### 4. Behavior 模块 (`internal/behavior`)

**职责**: 模拟各种运行时行为

#### 4.1 接口定义

```go
package behavior

import (
    "time"
    "mock-service/pkg/config"
)

// Simulator 行为模拟器接口
type Simulator interface {
    // Init 初始化模拟器
    Init(cfg config.BehaviorConfig) error
    
    // TriggerCPU 触发 CPU 负载
    TriggerCPU(percent int, duration time.Duration) error
    
    // TriggerMemory 触发内存负载
    TriggerMemory(mb int, duration time.Duration) error
    
    // TriggerCrash 触发崩溃
    TriggerCrash(delay time.Duration)
    
    // CheckStartup 检查启动配置
    CheckStartup() error
}
```

#### 4.2 CPU 模拟实现

```go
package behavior

import (
    "runtime"
    "time"
)

type cpuWorker struct{}

func NewCPUWorker() *cpuWorker {
    return &cpuWorker{}
}

func (w *cpuWorker) Simulate(targetPercent int, duration time.Duration) {
    numCPU := runtime.NumCPU()
    
    for i := 0; i < numCPU; i++ {
        go func() {
            start := time.Now()
            cycleTime := 100 * time.Millisecond  // 每个周期 100ms
            workTime := time.Duration(targetPercent) * time.Millisecond
            sleepTime := cycleTime - workTime
            
            for time.Since(start) < duration {
                // 工作阶段: 忙循环
                workEnd := time.Now().Add(workTime)
                for time.Now().Before(workEnd) {
                    // Busy loop
                    for j := 0; j < 100000; j++ {
                        _ = j * j
                    }
                }
                // 休息阶段
                time.Sleep(sleepTime)
            }
        }()
    }
}
```

#### 4.3 内存模拟实现

```go
package behavior

import (
    "runtime"
    "time"
)

type memoryWorker struct {
    allocations [][]byte
}

func NewMemoryWorker() *memoryWorker {
    return &memoryWorker{
        allocations: make([][]byte, 0),
    }
}

func (w *memoryWorker) Simulate(targetMB int, duration time.Duration) {
    // 分配内存
    data := make([]byte, targetMB*1024*1024)
    for i := range data {
        data[i] = byte(i % 256)
    }
    w.allocations = append(w.allocations, data)
    
    // 持有一段时间
    time.Sleep(duration)
    
    // 释放内存
    w.allocations = w.allocations[:0]
    runtime.GC()
}
```

---

### 5. Template 模块 (`internal/template`)

**职责**: 模板渲染、变量替换

#### 5.1 接口定义

```go
package template

import (
    "github.com/gin-gonic/gin"
)

// Renderer 模板渲染器接口
type Renderer interface {
    // Render 渲染模板
    Render(template interface{}, ctx *gin.Context) (interface{}, error)
}
```

#### 5.2 实现

```go
package template

import (
    "bytes"
    "encoding/json"
    "text/template"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type defaultRenderer struct{}

func New() Renderer {
    return &defaultRenderer{}
}

func (r *defaultRenderer) Render(tmpl interface{}, ctx *gin.Context) (interface{}, error) {
    // 准备模板数据
    data := r.prepareTemplateData(ctx)
    
    // 转换为 JSON 字符串
    jsonBytes, err := json.Marshal(tmpl)
    if err != nil {
        return nil, err
    }
    
    // 执行模板渲染
    t, err := template.New("response").Parse(string(jsonBytes))
    if err != nil {
        return nil, fmt.Errorf("template parse error: %w (check your template syntax in response body)", err)
    }
    
    var buf bytes.Buffer
    if err := t.Execute(&buf, data); err != nil {
        return nil, fmt.Errorf("template execution error: %w (available variables: PathParam, QueryParam, Service, RandomID, Timestamp)", err)
    }
    
    // 解析回 JSON
    var result interface{}
    if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
        return nil, err
    }
    
    return result, nil
}

func (r *defaultRenderer) prepareTemplateData(ctx *gin.Context) map[string]interface{} {
    return map[string]interface{}{
        "PathParam":  extractPathParams(ctx),
        "QueryParam": extractQueryParams(ctx),
        "Service": map[string]string{
            "Name":    ctx.GetString("service_name"),
            "Version": ctx.GetString("service_version"),
        },
        "RandomID":  uuid.New().String(),
        "Timestamp": time.Now().Format(time.RFC3339),
    }
}

func extractPathParams(ctx *gin.Context) map[string]string {
    params := make(map[string]string)
    for _, param := range ctx.Params {
        params[param.Key] = param.Value
    }
    return params
}

func extractQueryParams(ctx *gin.Context) map[string]string {
    params := make(map[string]string)
    for key, values := range ctx.Request.URL.Query() {
        if len(values) > 0 {
            params[key] = values[0]
        }
    }
    return params
}
```

---

### 6. Logger 模块 (`pkg/logger`)

**职责**: 结构化日志

#### 6.1 接口定义

```go
package logger

// Logger 日志接口
type Logger interface {
    Debug(msg string, keysAndValues ...interface{})
    Info(msg string, keysAndValues ...interface{})
    Warn(msg string, keysAndValues ...interface{})
    Error(msg string, keysAndValues ...interface{})
}
```

#### 6.2 实现（基于 zap）

```go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "mock-service/pkg/config"
)

type zapLogger struct {
    logger *zap.SugaredLogger
}

func New(cfg config.LoggingConfig) (Logger, error) {
    // 配置 zap
    zapConfig := zap.NewProductionConfig()
    
    // 设置日志级别
    level, err := zapcore.ParseLevel(cfg.Level)
    if err != nil {
        level = zapcore.InfoLevel
    }
    zapConfig.Level = zap.NewAtomicLevelAt(level)
    
    // 设置输出格式
    if cfg.Format == "console" {
        zapConfig.Encoding = "console"
    }
    
    // 构建 logger
    l, err := zapConfig.Build()
    if err != nil {
        return nil, err
    }
    
    return &zapLogger{
        logger: l.Sugar(),
    }, nil
}

func (l *zapLogger) Debug(msg string, keysAndValues ...interface{}) {
    l.logger.Debugw(msg, keysAndValues...)
}

func (l *zapLogger) Info(msg string, keysAndValues ...interface{}) {
    l.logger.Infow(msg, keysAndValues...)
}

func (l *zapLogger) Warn(msg string, keysAndValues ...interface{}) {
    l.logger.Warnw(msg, keysAndValues...)
}

func (l *zapLogger) Error(msg string, keysAndValues ...interface{}) {
    l.logger.Errorw(msg, keysAndValues...)
}
```

---

## 🔄 模块交互流程

### 场景 1: 静态响应流程

```
1. Client → HTTP Request
2. Server.engine → Middleware Chain
   - RequestID: 生成/提取 Request ID
   - Logger: 记录请求开始
3. Server.createHandler → 应用延迟和错误注入
4. Handler.HandleStaticResponse
   - Template.Render: 渲染模板
   - 返回 JSON 响应
5. Middleware.Logger → 记录请求完成
6. Server.engine → HTTP Response
```

### 场景 2: 代理请求流程

```
1. Client → HTTP Request (带 X-Traffic-Tag header)
2. Server.engine → Middleware Chain
3. Handler.HandleProxy
   - 构建目标 URL
   - 复制 Headers (包括 X-Traffic-Tag)
   - 发送请求到下游服务
   - 接收下游响应
   - 原样返回响应
4. Server.engine → HTTP Response
```

### 场景 3: 行为模拟流程

```
1. Main → Config.Load
2. Main → Behavior.Init(config.Behaviors)
3. Behavior.CheckStartup
   - 如果 startup.fail=true，返回错误，程序退出
4. Behavior.TriggerCPU (如果 trigger.type=startup)
   - cpuWorker.Simulate: 启动多个 Goroutine 消耗 CPU
5. Server.Start
6. 外部 node_exporter 采集 CPU 指标
```

---

## 📊 依赖关系图

```
main.go
  ├── config.Loader → 加载配置
  ├── logger.New → 创建日志
  ├── behavior.Simulator → 初始化行为模拟器
  ├── template.Renderer → 创建模板渲染器
  ├── handler.Handler → 创建请求处理器
  │     ├── handler.StaticHandler
  │     ├── handler.ProxyHandler
  │     └── template.Renderer
  └── server.Server → 创建并启动服务器
        ├── handler.Handler
        ├── logger.Logger
        └── middleware.*
```

---

## 🧪 测试策略

### 单元测试

每个模块独立测试:

```go
// config_test.go
func TestConfigLoader_Load(t *testing.T) {
    tests := []struct {
        name    string
        config  string
        wantErr bool
    }{
        {"valid config", "testdata/valid.yaml", false},
        {"invalid port", "testdata/invalid_port.yaml", true},
    }
    // ...
}

// handler_test.go
func TestStaticHandler_Handle(t *testing.T) {
    // 使用 httptest 测试处理器
}

// behavior_test.go
func TestCPUWorker_Simulate(t *testing.T) {
    // 验证 CPU 占用是否达到目标值
}
```

### 集成测试

```go
// integration_test.go
func TestMockServer_E2E(t *testing.T) {
    // 1. 加载配置
    // 2. 启动服务器
    // 3. 发送 HTTP 请求
    // 4. 验证响应
    // 5. 关闭服务器
}
```

---

## 📦 构建与部署

### Makefile

```makefile
.PHONY: build test lint

build:
	go build -o bin/mock-server cmd/mock-server/main.go

test:
	go test -v -cover ./...

lint:
	golangci-lint run ./...

docker-build:
	docker build -t mock-service:latest .

run:
	./bin/mock-server --config configs/example.yaml
```

---

**最后更新**: 2025-10-24
