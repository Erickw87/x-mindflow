package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LiusCraft/x-mindflow/mock-service/internal/behavior"
	"github.com/LiusCraft/x-mindflow/mock-service/internal/handler"
	"github.com/LiusCraft/x-mindflow/mock-service/internal/server"
	"github.com/LiusCraft/x-mindflow/mock-service/internal/template"
	"github.com/LiusCraft/x-mindflow/mock-service/pkg/config"
	"github.com/LiusCraft/x-mindflow/mock-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

func TestIntegration_BasicEndpoint(t *testing.T) {
	cfg := createTestConfig(t, `
service:
  name: integration-test
  version: v1.0.0
  port: 18080

endpoints:
  - path: /api/users/:id
    method: GET
    response:
      status: 200
      headers:
        Content-Type: application/json
      body:
        user_id: "{{.PathParam.id}}"
        service: "{{.Service.Name}}"
        version: "{{.Service.Version}}"
`)

	srv := setupTestServer(t, cfg)
	defer srv.Shutdown(nil)

	resp, body := makeRequest(t, "GET", "http://localhost:18080/api/users/123", nil)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["user_id"] != "123" {
		t.Errorf("user_id = %v, want 123", result["user_id"])
	}

	if result["service"] != "integration-test" {
		t.Errorf("service = %v, want integration-test", result["service"])
	}
}

func TestIntegration_HealthCheck(t *testing.T) {
	cfg := createTestConfig(t, `
service:
  name: health-test
  version: v1.0.0
  port: 18081

observability:
  health_check:
    enabled: true
    path: /health

endpoints:
  - path: /api/test
    method: GET
    response:
      status: 200
      body:
        status: ok
`)

	srv := setupTestServer(t, cfg)
	defer srv.Shutdown(nil)

	resp, body := makeRequest(t, "GET", "http://localhost:18081/health", nil)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["status"] != "healthy" {
		t.Errorf("status = %v, want healthy", result["status"])
	}
}

func TestIntegration_MultipleResponses(t *testing.T) {
	cfg := createTestConfig(t, `
service:
  name: multi-response-test
  version: v1.0.0
  port: 18082

endpoints:
  - path: /api/random
    method: GET
    responses:
      - status: 200
        body:
          result: success
        weight: 60
      - status: 500
        body:
          result: error
        weight: 40
`)

	srv := setupTestServer(t, cfg)
	defer srv.Shutdown(nil)

	statusCounts := make(map[int]int)
	iterations := 50

	for i := 0; i < iterations; i++ {
		resp, _ := makeRequest(t, "GET", "http://localhost:18082/api/random", nil)
		resp.Body.Close()
		statusCounts[resp.StatusCode]++
	}

	if statusCounts[200] == 0 {
		t.Error("never received 200 response")
	}

	if statusCounts[500] == 0 {
		t.Error("never received 500 response")
	}
}

func TestIntegration_Proxy(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Traffic-Tag") != "test-tag" {
			t.Errorf("X-Traffic-Tag = %s, want test-tag", r.Header.Get("X-Traffic-Tag"))
		}

		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]string{
			"backend": "response",
		})
	}))
	defer backend.Close()

	configYAML := fmt.Sprintf(`
service:
  name: proxy-test
  version: v1.0.0
  port: 18083

endpoints:
  - path: /api/proxy
    method: GET
    proxy:
      target: %s
      preserve_headers: true
`, backend.URL)

	cfg := createTestConfig(t, configYAML)
	srv := setupTestServer(t, cfg)
	defer srv.Shutdown(nil)

	req, _ := http.NewRequest("GET", "http://localhost:18083/api/proxy", nil)
	req.Header.Set("X-Traffic-Tag", "test-tag")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	var result map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["backend"] != "response" {
		t.Errorf("backend = %v, want response", result["backend"])
	}
}

func TestIntegration_LatencySimulation(t *testing.T) {
	cfg := createTestConfig(t, `
service:
  name: latency-test
  version: v1.0.0
  port: 18084

endpoints:
  - path: /api/slow
    method: GET
    latency: 200ms
    response:
      status: 200
      body:
        message: slow response
`)

	srv := setupTestServer(t, cfg)
	defer srv.Shutdown(nil)

	start := time.Now()
	resp, _ := makeRequest(t, "GET", "http://localhost:18084/api/slow", nil)
	elapsed := time.Since(start)
	resp.Body.Close()

	if elapsed < 200*time.Millisecond {
		t.Errorf("response too fast: %v, want >= 200ms", elapsed)
	}

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestIntegration_QueryParameters(t *testing.T) {
	cfg := createTestConfig(t, `
service:
  name: query-test
  version: v1.0.0
  port: 18085

endpoints:
  - path: /api/search
    method: GET
    response:
      status: 200
      body:
        query: "{{.QueryParam.q}}"
        page: "{{.QueryParam.page}}"
`)

	srv := setupTestServer(t, cfg)
	defer srv.Shutdown(nil)

	resp, body := makeRequest(t, "GET", "http://localhost:18085/api/search?q=test&page=2", nil)
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result["query"] != "test" {
		t.Errorf("query = %v, want test", result["query"])
	}

	if result["page"] != "2" {
		t.Errorf("page = %v, want 2", result["page"])
	}
}

func createTestConfig(t *testing.T, configYAML string) *config.ServiceConfig {
	t.Helper()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	loader := config.NewLoader()
	cfg, err := loader.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	return cfg
}

func setupTestServer(t *testing.T, cfg *config.ServiceConfig) server.Server {
	t.Helper()

	gin.SetMode(gin.TestMode)

	log, err := logger.New("error", "json")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	sim := behavior.New(log)
	if err := sim.Init(cfg.Behaviors); err != nil {
		t.Fatalf("failed to init simulator: %v", err)
	}

	renderer := template.New(cfg.Service.Name, cfg.Service.Version)
	staticHandler := handler.NewStaticHandler(renderer)
	proxyHandler := handler.NewProxyHandler(log)
	h := handler.New(staticHandler, proxyHandler)

	srv := server.New(cfg, h, sim, log)

	go func() {
		if err := srv.Start(); err != nil {
			t.Logf("server error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	return srv
}

func makeRequest(t *testing.T, method, url string, body io.Reader) (*http.Response, []byte) {
	t.Helper()

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	return resp, respBody
}
