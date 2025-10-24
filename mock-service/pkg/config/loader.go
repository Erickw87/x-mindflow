package config

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// Loader 配置加载器接口
type Loader interface {
	Load(path string) (*ServiceConfig, error)
	Validate(config *ServiceConfig) error
}

type defaultLoader struct {
	validator *validator.Validate
}

// NewLoader 创建配置加载器
func NewLoader() Loader {
	return &defaultLoader{
		validator: validator.New(),
	}
}

func (l *defaultLoader) Load(path string) (*ServiceConfig, error) {
	if path == "" {
		return nil, fmt.Errorf("config path cannot be empty")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("config file is empty: %s", path)
	}

	var config ServiceConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if err := l.Validate(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (l *defaultLoader) Validate(config *ServiceConfig) error {
	if err := l.validator.Struct(config); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	if err := l.validateBusinessRules(config); err != nil {
		return err
	}

	return nil
}

func (l *defaultLoader) validateBusinessRules(config *ServiceConfig) error {
	for i, ep := range config.Endpoints {
		if ep.Response == nil && ep.Proxy == nil && len(ep.Responses) == 0 {
			return fmt.Errorf("endpoint[%d] (%s %s): must have either response, proxy, or responses", i, ep.Method, ep.Path)
		}

		if len(ep.Responses) > 0 {
			totalWeight := 0
			for _, resp := range ep.Responses {
				totalWeight += resp.Weight
			}
			if totalWeight != 100 {
				return fmt.Errorf("endpoint[%d] (%s %s): response weights must sum to 100, got %d", i, ep.Method, ep.Path, totalWeight)
			}
		}
	}

	if config.Behaviors.CPU.Enabled {
		if config.Behaviors.CPU.Trigger.Type == "endpoint" && config.Behaviors.CPU.Trigger.Endpoint == "" {
			return fmt.Errorf("cpu behavior: endpoint trigger requires endpoint path")
		}
	}

	if config.Behaviors.Memory.Enabled {
		if config.Behaviors.Memory.Trigger.Type == "endpoint" && config.Behaviors.Memory.Trigger.Endpoint == "" {
			return fmt.Errorf("memory behavior: endpoint trigger requires endpoint path")
		}
	}

	if config.Behaviors.Crash.Enabled {
		if config.Behaviors.Crash.Trigger.Type == "endpoint" && config.Behaviors.Crash.Trigger.Endpoint == "" {
			return fmt.Errorf("crash behavior: endpoint trigger requires endpoint path")
		}
	}

	return nil
}
