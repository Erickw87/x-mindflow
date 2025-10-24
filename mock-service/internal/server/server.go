package server

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/LiusCraft/x-mindflow/mock-service/internal/behavior"
	"github.com/LiusCraft/x-mindflow/mock-service/internal/handler"
	"github.com/LiusCraft/x-mindflow/mock-service/internal/server/middleware"
	"github.com/LiusCraft/x-mindflow/mock-service/pkg/config"
	"github.com/LiusCraft/x-mindflow/mock-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Server HTTP 服务器接口
type Server interface {
	Start() error
	Shutdown(ctx context.Context) error
}

type mockServer struct {
	engine    *gin.Engine
	config    *config.ServiceConfig
	handler   handler.Handler
	simulator behavior.Simulator
	server    *http.Server
	logger    logger.Logger
	startTime time.Time
}

// New 创建服务器实例
func New(cfg *config.ServiceConfig, h handler.Handler, sim behavior.Simulator, log logger.Logger) Server {
	return &mockServer{
		config:    cfg,
		handler:   h,
		simulator: sim,
		logger:    log,
		startTime: time.Now(),
	}
}

func (s *mockServer) Start() error {
	if s.config == nil {
		return fmt.Errorf("server config is nil")
	}

	if s.config.Service.Port <= 0 || s.config.Service.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", s.config.Service.Port)
	}

	gin.SetMode(gin.ReleaseMode)
	s.engine = gin.New()

	s.engine.Use(
		middleware.RequestID(),
		middleware.Logger(s.logger),
		middleware.Recovery(s.logger),
	)

	if s.config.Observability.HealthCheck.Enabled {
		s.engine.GET(s.config.Observability.HealthCheck.Path, s.healthCheckHandler)
	}

	if err := s.registerEndpoints(s.config.Endpoints); err != nil {
		return fmt.Errorf("failed to register endpoints: %w", err)
	}

	addr := fmt.Sprintf(":%d", s.config.Service.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.Info("Server starting",
		"address", addr,
		"service", s.config.Service.Name,
		"version", s.config.Service.Version)

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func (s *mockServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Server shutting down")
	return s.server.Shutdown(ctx)
}

func (s *mockServer) registerEndpoints(endpoints []config.EndpointConfig) error {
	healthPath := s.config.Observability.HealthCheck.Path
	for _, ep := range endpoints {
		if s.config.Observability.HealthCheck.Enabled && ep.Path == healthPath {
			s.logger.Warn("Skipping endpoint that conflicts with health check",
				"path", ep.Path,
				"method", ep.Method)
			continue
		}

		handlerFunc := s.createHandler(ep)
		s.engine.Handle(ep.Method, ep.Path, handlerFunc)
		s.logger.Debug("Registered endpoint",
			"method", ep.Method,
			"path", ep.Path)
	}
	return nil
}

func (s *mockServer) createHandler(ep config.EndpointConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if ep.Latency > 0 {
			time.Sleep(ep.Latency)
		}

		if ep.ErrorRate > 0 && rand.Float64() < ep.ErrorRate {
			s.logger.Warn("Injected error",
				"request_id", c.GetString("request_id"),
				"path", c.Request.URL.Path)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "injected error"})
			return
		}

		s.checkBehaviorTriggers(c.Request.URL.Path)

		if ep.Proxy != nil {
			s.handler.HandleProxy(c, ep.Proxy)
		} else if len(ep.Responses) > 0 {
			s.handler.HandleRandomResponse(c, ep.Responses)
		} else if ep.Response != nil {
			s.handler.HandleStaticResponse(c, ep.Response)
		}
	}
}

func (s *mockServer) checkBehaviorTriggers(path string) {
	if s.config.Behaviors.CPU.Enabled &&
		s.config.Behaviors.CPU.Trigger.Type == "endpoint" &&
		s.config.Behaviors.CPU.Trigger.Endpoint == path {
		s.simulator.TriggerCPU(s.config.Behaviors.CPU.TargetPercent, s.config.Behaviors.CPU.Duration)
	}

	if s.config.Behaviors.Memory.Enabled &&
		s.config.Behaviors.Memory.Trigger.Type == "endpoint" &&
		s.config.Behaviors.Memory.Trigger.Endpoint == path {
		s.simulator.TriggerMemory(s.config.Behaviors.Memory.TargetMB, s.config.Behaviors.Memory.Duration)
	}

	if s.config.Behaviors.Crash.Enabled &&
		s.config.Behaviors.Crash.Trigger.Type == "endpoint" &&
		s.config.Behaviors.Crash.Trigger.Endpoint == path {
		s.simulator.TriggerCrash(s.config.Behaviors.Crash.Delay)
	}
}

func (s *mockServer) healthCheckHandler(c *gin.Context) {
	uptime := time.Since(s.startTime)

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"version":   s.config.Service.Version,
		"uptime":    uptime.String(),
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
