package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoader_Load(t *testing.T) {
	tests := []struct {
		name        string
		configYAML  string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			configYAML: `
service:
  name: test-service
  version: v1.0.0
  port: 8080

endpoints:
  - path: /api/test
    method: GET
    response:
      status: 200
      body:
        message: "test"
`,
			wantErr: false,
		},
		{
			name: "missing service name",
			configYAML: `
service:
  version: v1.0.0
  port: 8080

endpoints:
  - path: /api/test
    method: GET
    response:
      status: 200
      body: {}
`,
			wantErr:     true,
			errContains: "validation failed",
		},
		{
			name: "invalid port",
			configYAML: `
service:
  name: test-service
  version: v1.0.0
  port: 99999

endpoints:
  - path: /api/test
    method: GET
    response:
      status: 200
      body: {}
`,
			wantErr:     true,
			errContains: "validation failed",
		},
		{
			name: "endpoint without response or proxy",
			configYAML: `
service:
  name: test-service
  version: v1.0.0
  port: 8080

endpoints:
  - path: /api/test
    method: GET
`,
			wantErr:     true,
			errContains: "must have either response, proxy, or responses",
		},
		{
			name: "responses with incorrect weight sum",
			configYAML: `
service:
  name: test-service
  version: v1.0.0
  port: 8080

endpoints:
  - path: /api/test
    method: GET
    responses:
      - status: 200
        body: {}
        weight: 50
      - status: 500
        body: {}
        weight: 30
`,
			wantErr:     true,
			errContains: "must sum to 100",
		},
		{
			name: "valid responses with correct weights",
			configYAML: `
service:
  name: test-service
  version: v1.0.0
  port: 8080

endpoints:
  - path: /api/test
    method: GET
    responses:
      - status: 200
        body: {}
        weight: 70
      - status: 500
        body: {}
        weight: 30
`,
			wantErr: false,
		},
		{
			name: "cpu behavior with endpoint trigger but no endpoint",
			configYAML: `
service:
  name: test-service
  version: v1.0.0
  port: 8080

behaviors:
  cpu:
    enabled: true
    target_percent: 80
    trigger:
      type: endpoint

endpoints:
  - path: /api/test
    method: GET
    response:
      status: 200
      body: {}
`,
			wantErr:     true,
			errContains: "endpoint trigger requires endpoint path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			
			if err := os.WriteFile(configPath, []byte(tt.configYAML), 0644); err != nil {
				t.Fatalf("failed to write test config: %v", err)
			}

			loader := NewLoader()
			cfg, err := loader.Load(configPath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error but got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Load() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Load() unexpected error = %v", err)
				}
				if cfg == nil {
					t.Error("Load() returned nil config")
				}
			}
		})
	}
}

func TestLoader_LoadFileNotFound(t *testing.T) {
	loader := NewLoader()
	_, err := loader.Load("/nonexistent/path/config.yaml")
	
	if err == nil {
		t.Error("Load() expected error for nonexistent file")
	}
	if !contains(err.Error(), "config file not found") {
		t.Errorf("Load() error = %v, want error containing 'config file not found'", err)
	}
}

func TestLoader_LoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	invalidYAML := `
service:
  name: test
  invalid yaml: [[[
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	loader := NewLoader()
	_, err := loader.Load(configPath)
	
	if err == nil {
		t.Error("Load() expected error for invalid YAML")
	}
	if !contains(err.Error(), "failed to parse YAML") {
		t.Errorf("Load() error = %v, want error containing 'failed to parse YAML'", err)
	}
}

func TestLoader_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *ServiceConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			config: &ServiceConfig{
				Service: ServiceInfo{
					Name:    "test-service",
					Version: "v1.0.0",
					Port:    8080,
				},
				Endpoints: []EndpointConfig{
					{
						Path:   "/test",
						Method: "GET",
						Response: &ResponseConfig{
							Status: 200,
							Body:   map[string]interface{}{"message": "ok"},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing port",
			config: &ServiceConfig{
				Service: ServiceInfo{
					Name:    "test-service",
					Version: "v1.0.0",
				},
				Endpoints: []EndpointConfig{
					{
						Path:   "/test",
						Method: "GET",
						Response: &ResponseConfig{
							Status: 200,
							Body:   map[string]interface{}{},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "validation failed",
		},
		{
			name: "invalid endpoint path",
			config: &ServiceConfig{
				Service: ServiceInfo{
					Name:    "test-service",
					Version: "v1.0.0",
					Port:    8080,
				},
				Endpoints: []EndpointConfig{
					{
						Path:   "invalid",
						Method: "GET",
						Response: &ResponseConfig{
							Status: 200,
							Body:   map[string]interface{}{},
						},
					},
				},
			},
			wantErr:     true,
			errContains: "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader()
			err := loader.Validate(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error but got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestLoader_validateBusinessRules(t *testing.T) {
	loader := &defaultLoader{
		validator: NewLoader().(*defaultLoader).validator,
	}

	tests := []struct {
		name        string
		config      *ServiceConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "endpoint with proxy",
			config: &ServiceConfig{
				Service: ServiceInfo{Name: "test", Version: "v1", Port: 8080},
				Endpoints: []EndpointConfig{
					{
						Path:   "/test",
						Method: "GET",
						Proxy: &ProxyConfig{
							Target: "http://example.com",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "memory behavior with endpoint trigger missing endpoint",
			config: &ServiceConfig{
				Service: ServiceInfo{Name: "test", Version: "v1", Port: 8080},
				Endpoints: []EndpointConfig{
					{
						Path:   "/test",
						Method: "GET",
						Response: &ResponseConfig{
							Status: 200,
							Body:   map[string]interface{}{},
						},
					},
				},
				Behaviors: BehaviorConfig{
					Memory: MemoryBehavior{
						Enabled:  true,
						TargetMB: 100,
						Trigger: TriggerConfig{
							Type: "endpoint",
						},
					},
				},
			},
			wantErr:     true,
			errContains: "endpoint trigger requires endpoint path",
		},
		{
			name: "crash behavior with endpoint trigger missing endpoint",
			config: &ServiceConfig{
				Service: ServiceInfo{Name: "test", Version: "v1", Port: 8080},
				Endpoints: []EndpointConfig{
					{
						Path:   "/test",
						Method: "GET",
						Response: &ResponseConfig{
							Status: 200,
							Body:   map[string]interface{}{},
						},
					},
				},
				Behaviors: BehaviorConfig{
					Crash: CrashBehavior{
						Enabled: true,
						Trigger: TriggerConfig{
							Type: "endpoint",
						},
						Delay: 1 * time.Second,
					},
				},
			},
			wantErr:     true,
			errContains: "endpoint trigger requires endpoint path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loader.validateBusinessRules(tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("validateBusinessRules() expected error but got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("validateBusinessRules() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("validateBusinessRules() unexpected error = %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
