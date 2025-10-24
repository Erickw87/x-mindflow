package handler

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/LiusCraft/x-mindflow/mock-service/internal/template"
	"github.com/LiusCraft/x-mindflow/mock-service/pkg/config"
	"github.com/gin-gonic/gin"
)

func TestStaticHandler_Handle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		cfg            *config.ResponseConfig
		wantStatus     int
		wantBody       map[string]interface{}
		wantHeader     string
		wantHeaderVal  string
	}{
		{
			name: "simple response",
			cfg: &config.ResponseConfig{
				Status: 200,
				Body: map[string]interface{}{
					"message": "success",
				},
			},
			wantStatus: 200,
			wantBody: map[string]interface{}{
				"message": "success",
			},
		},
		{
			name: "response with headers",
			cfg: &config.ResponseConfig{
				Status: 201,
				Headers: map[string]string{
					"X-Custom-Header": "custom-value",
				},
				Body: map[string]interface{}{
					"created": true,
				},
			},
			wantStatus:    201,
			wantHeader:    "X-Custom-Header",
			wantHeaderVal: "custom-value",
			wantBody: map[string]interface{}{
				"created": true,
			},
		},
		{
			name: "error response",
			cfg: &config.ResponseConfig{
				Status: 500,
				Body: map[string]interface{}{
					"error": "internal server error",
				},
			},
			wantStatus: 500,
			wantBody: map[string]interface{}{
				"error": "internal server error",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/test", nil)

			renderer := template.New("test-service", "v1.0.0")
			handler := NewStaticHandler(renderer)
			handler.Handle(c, tt.cfg)

			if w.Code != tt.wantStatus {
				t.Errorf("Handle() status = %d, want %d", w.Code, tt.wantStatus)
			}

			if tt.wantHeader != "" {
				if got := w.Header().Get(tt.wantHeader); got != tt.wantHeaderVal {
					t.Errorf("Handle() header %s = %s, want %s", tt.wantHeader, got, tt.wantHeaderVal)
				}
			}

			var body map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("Failed to unmarshal response body: %v", err)
			}

			for key, wantValue := range tt.wantBody {
				if gotValue, ok := body[key]; !ok || gotValue != wantValue {
					t.Errorf("Handle() body[%s] = %v, want %v", key, gotValue, wantValue)
				}
			}
		})
	}
}

func TestStaticHandler_HandleRandom(t *testing.T) {
	gin.SetMode(gin.TestMode)

	responses := []config.ResponseConfig{
		{
			Status: 200,
			Body:   map[string]interface{}{"status": "ok"},
			Weight: 70,
		},
		{
			Status: 500,
			Body:   map[string]interface{}{"status": "error"},
			Weight: 30,
		},
	}

	renderer := template.New("test-service", "v1.0.0")
	handler := NewStaticHandler(renderer)

	statusCounts := make(map[int]int)
	iterations := 100

	for i := 0; i < iterations; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		handler.HandleRandom(c, responses)
		statusCounts[w.Code]++
	}

	if statusCounts[200] == 0 {
		t.Error("HandleRandom() never returned status 200")
	}

	if statusCounts[500] == 0 {
		t.Error("HandleRandom() never returned status 500")
	}

	if statusCounts[200]+statusCounts[500] != iterations {
		t.Errorf("HandleRandom() total responses = %d, want %d", statusCounts[200]+statusCounts[500], iterations)
	}
}

func TestStaticHandler_selectRandomResponse(t *testing.T) {
	tests := []struct {
		name      string
		responses []config.ResponseConfig
		validate  func(t *testing.T, selected config.ResponseConfig, responses []config.ResponseConfig)
	}{
		{
			name: "single response",
			responses: []config.ResponseConfig{
				{Status: 200, Weight: 100},
			},
			validate: func(t *testing.T, selected config.ResponseConfig, responses []config.ResponseConfig) {
				if selected.Status != 200 {
					t.Errorf("selectRandomResponse() status = %d, want 200", selected.Status)
				}
			},
		},
		{
			name: "multiple responses with weights",
			responses: []config.ResponseConfig{
				{Status: 200, Weight: 50},
				{Status: 404, Weight: 30},
				{Status: 500, Weight: 20},
			},
			validate: func(t *testing.T, selected config.ResponseConfig, responses []config.ResponseConfig) {
				validStatus := false
				for _, resp := range responses {
					if selected.Status == resp.Status {
						validStatus = true
						break
					}
				}
				if !validStatus {
					t.Errorf("selectRandomResponse() returned unexpected status %d", selected.Status)
				}
			},
		},
		{
			name: "all weight on first response",
			responses: []config.ResponseConfig{
				{Status: 200, Weight: 100},
				{Status: 500, Weight: 0},
			},
			validate: func(t *testing.T, selected config.ResponseConfig, responses []config.ResponseConfig) {
				count200 := 0
				iterations := 20
				handler := &staticHandler{}
				
				for i := 0; i < iterations; i++ {
					sel := handler.selectRandomResponse(responses)
					if sel.Status == 200 {
						count200++
					}
				}
				
				if count200 != iterations {
					t.Errorf("selectRandomResponse() with 100%% weight on 200 returned it %d/%d times", count200, iterations)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &staticHandler{}
			selected := handler.selectRandomResponse(tt.responses)
			tt.validate(t, selected, tt.responses)
		})
	}
}
