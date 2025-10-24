package behavior

import (
	"fmt"
	"os"
	"time"

	"github.com/LiusCraft/x-mindflow/mock-service/pkg/config"
	"github.com/LiusCraft/x-mindflow/mock-service/pkg/logger"
)

// Simulator 行为模拟器接口
type Simulator interface {
	Init(cfg config.BehaviorConfig) error
	TriggerCPU(percent int, duration time.Duration) error
	TriggerMemory(mb int, duration time.Duration) error
	TriggerCrash(delay time.Duration)
	CheckStartup() error
}

type defaultSimulator struct {
	config    config.BehaviorConfig
	cpuWorker *cpuWorker
	memWorker *memoryWorker
	logger    logger.Logger
}

// New 创建行为模拟器
func New(log logger.Logger) Simulator {
	return &defaultSimulator{
		cpuWorker: newCPUWorker(log),
		memWorker: newMemoryWorker(log),
		logger:    log,
	}
}

func (s *defaultSimulator) Init(cfg config.BehaviorConfig) error {
	s.config = cfg

	if cfg.CPU.Enabled && cfg.CPU.Trigger.Type == "startup" {
		s.logger.Info("Triggering CPU load on startup",
			"target_percent", cfg.CPU.TargetPercent,
			"duration", cfg.CPU.Duration)
		go s.cpuWorker.Simulate(cfg.CPU.TargetPercent, cfg.CPU.Duration)
	}

	if cfg.Memory.Enabled && cfg.Memory.Trigger.Type == "startup" {
		s.logger.Info("Triggering memory load on startup",
			"target_mb", cfg.Memory.TargetMB,
			"duration", cfg.Memory.Duration)
		go s.memWorker.Simulate(cfg.Memory.TargetMB, cfg.Memory.Duration)
	}

	return nil
}

func (s *defaultSimulator) TriggerCPU(percent int, duration time.Duration) error {
	s.logger.Info("Triggering CPU load", "percent", percent, "duration", duration)
	go s.cpuWorker.Simulate(percent, duration)
	return nil
}

func (s *defaultSimulator) TriggerMemory(mb int, duration time.Duration) error {
	s.logger.Info("Triggering memory load", "mb", mb, "duration", duration)
	go s.memWorker.Simulate(mb, duration)
	return nil
}

func (s *defaultSimulator) TriggerCrash(delay time.Duration) {
	s.logger.Warn("Crash triggered", "delay", delay)
	go func() {
		time.Sleep(delay)
		s.logger.Error("Simulated crash - exiting")
		os.Exit(1)
	}()
}

func (s *defaultSimulator) CheckStartup() error {
	if s.config.Startup.Fail {
		if s.config.Startup.Delay > 0 {
			s.logger.Info("Delaying startup", "delay", s.config.Startup.Delay)
			time.Sleep(s.config.Startup.Delay)
		}
		exitCode := s.config.Startup.ExitCode
		if exitCode == 0 {
			exitCode = 1
		}
		s.logger.Error("Simulated startup failure", "exit_code", exitCode)
		return fmt.Errorf("simulated startup failure (exit code: %d)", exitCode)
	}

	if s.config.Startup.Delay > 0 {
		s.logger.Info("Startup delay configured", "delay", s.config.Startup.Delay)
		time.Sleep(s.config.Startup.Delay)
	}

	return nil
}
