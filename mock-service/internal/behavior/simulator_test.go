package behavior

import (
	"testing"
	"time"

	"github.com/LiusCraft/x-mindflow/mock-service/pkg/config"
	"github.com/LiusCraft/x-mindflow/mock-service/pkg/logger"
)

func TestSimulator_Init(t *testing.T) {
	log, _ := logger.New("info", "json")

	tests := []struct {
		name    string
		cfg     config.BehaviorConfig
		wantErr bool
	}{
		{
			name: "no behaviors enabled",
			cfg:  config.BehaviorConfig{},
			wantErr: false,
		},
		{
			name: "cpu behavior on startup",
			cfg: config.BehaviorConfig{
				CPU: config.CPUBehavior{
					Enabled:       true,
					TargetPercent: 50,
					Duration:      100 * time.Millisecond,
					Trigger: config.TriggerConfig{
						Type: "startup",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "memory behavior on startup",
			cfg: config.BehaviorConfig{
				Memory: config.MemoryBehavior{
					Enabled:  true,
					TargetMB: 10,
					Duration: 100 * time.Millisecond,
					Trigger: config.TriggerConfig{
						Type: "startup",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "cpu behavior on endpoint trigger",
			cfg: config.BehaviorConfig{
				CPU: config.CPUBehavior{
					Enabled:       true,
					TargetPercent: 80,
					Duration:      1 * time.Second,
					Trigger: config.TriggerConfig{
						Type:     "endpoint",
						Endpoint: "/api/heavy",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := New(log)
			err := sim.Init(tt.cfg)

			if tt.wantErr && err == nil {
				t.Error("Init() expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Init() unexpected error = %v", err)
			}

			time.Sleep(150 * time.Millisecond)
		})
	}
}

func TestSimulator_CheckStartup(t *testing.T) {
	log, _ := logger.New("info", "json")

	tests := []struct {
		name        string
		cfg         config.BehaviorConfig
		wantErr     bool
		errContains string
	}{
		{
			name:    "no startup failure",
			cfg:     config.BehaviorConfig{},
			wantErr: false,
		},
		{
			name: "startup failure",
			cfg: config.BehaviorConfig{
				Startup: config.StartupBehavior{
					Fail:     true,
					ExitCode: 42,
				},
			},
			wantErr:     true,
			errContains: "simulated startup failure",
		},
		{
			name: "startup delay without failure",
			cfg: config.BehaviorConfig{
				Startup: config.StartupBehavior{
					Fail:  false,
					Delay: 50 * time.Millisecond,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := New(log)
			if err := sim.Init(tt.cfg); err != nil {
				t.Fatalf("Init() failed: %v", err)
			}

			start := time.Now()
			err := sim.CheckStartup()
			elapsed := time.Since(start)

			if tt.wantErr {
				if err == nil {
					t.Error("CheckStartup() expected error but got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("CheckStartup() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("CheckStartup() unexpected error = %v", err)
				}
			}

			if tt.cfg.Startup.Delay > 0 && elapsed < tt.cfg.Startup.Delay {
				t.Errorf("CheckStartup() elapsed = %v, want >= %v", elapsed, tt.cfg.Startup.Delay)
			}
		})
	}
}

func TestSimulator_TriggerCPU(t *testing.T) {
	log, _ := logger.New("info", "json")

	tests := []struct {
		name     string
		percent  int
		duration time.Duration
		wantErr  bool
	}{
		{
			name:     "low cpu load",
			percent:  10,
			duration: 100 * time.Millisecond,
			wantErr:  false,
		},
		{
			name:     "medium cpu load",
			percent:  50,
			duration: 100 * time.Millisecond,
			wantErr:  false,
		},
		{
			name:     "high cpu load",
			percent:  80,
			duration: 100 * time.Millisecond,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := New(log)
			err := sim.TriggerCPU(tt.percent, tt.duration)

			if tt.wantErr && err == nil {
				t.Error("TriggerCPU() expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("TriggerCPU() unexpected error = %v", err)
			}

			time.Sleep(tt.duration + 50*time.Millisecond)
		})
	}
}

func TestSimulator_TriggerMemory(t *testing.T) {
	log, _ := logger.New("info", "json")

	tests := []struct {
		name     string
		mb       int
		duration time.Duration
		wantErr  bool
	}{
		{
			name:     "small memory allocation",
			mb:       1,
			duration: 100 * time.Millisecond,
			wantErr:  false,
		},
		{
			name:     "medium memory allocation",
			mb:       10,
			duration: 100 * time.Millisecond,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := New(log)
			err := sim.TriggerMemory(tt.mb, tt.duration)

			if tt.wantErr && err == nil {
				t.Error("TriggerMemory() expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("TriggerMemory() unexpected error = %v", err)
			}

			time.Sleep(tt.duration + 50*time.Millisecond)
		})
	}
}

func TestSimulator_TriggerCrash(t *testing.T) {
	log, _ := logger.New("info", "json")

	sim := New(log)
	
	done := make(chan bool)
	go func() {
		sim.TriggerCrash(50 * time.Millisecond)
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
