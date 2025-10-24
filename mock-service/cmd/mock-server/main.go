package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LiusCraft/x-mindflow/mock-service/internal/behavior"
	"github.com/LiusCraft/x-mindflow/mock-service/internal/handler"
	"github.com/LiusCraft/x-mindflow/mock-service/internal/server"
	"github.com/LiusCraft/x-mindflow/mock-service/internal/template"
	"github.com/LiusCraft/x-mindflow/mock-service/pkg/config"
	"github.com/LiusCraft/x-mindflow/mock-service/pkg/logger"
)

const version = "v1.0.0"

func main() {
	var (
		configPath = flag.String("config", "", "Path to configuration file")
		port       = flag.Int("port", 0, "Override port from config")
		logLevel   = flag.String("log-level", "", "Override log level (debug, info, warn, error)")
		showVer    = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *showVer {
		fmt.Printf("Mock Service %s\n", version)
		os.Exit(0)
	}

	if *configPath == "" {
		fmt.Fprintf(os.Stderr, "Error: --config flag is required\n")
		flag.Usage()
		os.Exit(1)
	}

	loader := config.NewLoader()
	cfg, err := loader.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if *port > 0 {
		cfg.Service.Port = *port
	}

	level := cfg.Observability.Logging.Level
	if level == "" {
		level = "info"
	}
	if *logLevel != "" {
		level = *logLevel
	}

	format := cfg.Observability.Logging.Format
	if format == "" {
		format = "json"
	}

	log, err := logger.New(level, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	log.Info("Mock Service starting",
		"version", version,
		"service", cfg.Service.Name,
		"service_version", cfg.Service.Version,
		"port", cfg.Service.Port)

	sim := behavior.New(log)
	if err := sim.Init(cfg.Behaviors); err != nil {
		log.Error("Failed to initialize behavior simulator", "error", err)
		os.Exit(1)
	}

	if err := sim.CheckStartup(); err != nil {
		log.Error("Startup check failed", "error", err)
		exitCode := cfg.Behaviors.Startup.ExitCode
		if exitCode == 0 {
			exitCode = 1
		}
		os.Exit(exitCode)
	}

	renderer := template.New(cfg.Service.Name, cfg.Service.Version)
	staticHandler := handler.NewStaticHandler(renderer)
	proxyHandler := handler.NewProxyHandler(log)
	h := handler.New(staticHandler, proxyHandler)

	srv := server.New(cfg, h, sim, log)

	go func() {
		if err := srv.Start(); err != nil {
			log.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("Server exited")
}
