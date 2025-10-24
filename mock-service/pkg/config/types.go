package config

import "time"

// ServiceConfig 服务完整配置
type ServiceConfig struct {
	Service          ServiceInfo                `yaml:"service" validate:"required"`
	Behaviors        BehaviorConfig             `yaml:"behaviors"`
	Endpoints        []EndpointConfig           `yaml:"endpoints" validate:"required,dive"`
	VersionBehaviors map[string]VersionBehavior `yaml:"version_behaviors"`
	Observability    ObservabilityConfig        `yaml:"observability"`
}

// VersionBehavior 版本特定行为
type VersionBehavior struct {
	Endpoints []EndpointConfig `yaml:"endpoints"`
	Behaviors BehaviorConfig   `yaml:"behaviors"`
}

// ServiceInfo 服务基本信息
type ServiceInfo struct {
	Name    string `yaml:"name" validate:"required"`
	Version string `yaml:"version" validate:"required"`
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
	Duration      time.Duration `yaml:"duration"`
}

// MemoryBehavior 内存行为配置
type MemoryBehavior struct {
	Enabled  bool          `yaml:"enabled"`
	TargetMB int           `yaml:"target_mb" validate:"omitempty,min=1"`
	Trigger  TriggerConfig `yaml:"trigger"`
	Duration time.Duration `yaml:"duration"`
}

// TriggerConfig 触发器配置
type TriggerConfig struct {
	Type     string `yaml:"type" validate:"omitempty,oneof=startup endpoint periodic"`
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
	Path      string           `yaml:"path" validate:"required,startswith=/"`
	Method    string           `yaml:"method" validate:"required,oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
	Response  *ResponseConfig  `yaml:"response"`
	Proxy     *ProxyConfig     `yaml:"proxy"`
	Responses []ResponseConfig `yaml:"responses"`
	ErrorRate float64          `yaml:"error_rate" validate:"min=0,max=1"`
	Latency   time.Duration    `yaml:"latency"`
}

// ResponseConfig 响应配置
type ResponseConfig struct {
	Status  int               `yaml:"status" validate:"required,min=100,max=599"`
	Headers map[string]string `yaml:"headers"`
	Body    interface{}       `yaml:"body"`
	Weight  int               `yaml:"weight" validate:"min=0,max=100"`
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
	Level    string `yaml:"level" validate:"omitempty,oneof=debug info warn error"`
	Format   string `yaml:"format" validate:"omitempty,oneof=json console"`
	Output   string `yaml:"output" validate:"omitempty,oneof=stdout file"`
	FilePath string `yaml:"file_path"`
}
