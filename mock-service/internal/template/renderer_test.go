package template

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRenderer_Render(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		tmpl        interface{}
		setupCtx    func() *gin.Context
		wantErr     bool
		errContains string
		validate    func(t *testing.T, result interface{})
	}{
		{
			name: "simple static response",
			tmpl: map[string]interface{}{
				"message": "hello",
				"status":  "ok",
			},
			setupCtx: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/test", nil)
				return c
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("result is not a map")
				}
				if m["message"] != "hello" {
					t.Errorf("message = %v, want hello", m["message"])
				}
			},
		},
		{
			name: "template with path param",
			tmpl: map[string]interface{}{
				"user_id": "{{.PathParam.id}}",
				"message": "User details",
			},
			setupCtx: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/users/123", nil)
				c.Params = gin.Params{
					{Key: "id", Value: "123"},
				}
				return c
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("result is not a map")
				}
				if m["user_id"] != "123" {
					t.Errorf("user_id = %v, want 123", m["user_id"])
				}
			},
		},
		{
			name: "template with query param",
			tmpl: map[string]interface{}{
				"filter": "{{.QueryParam.status}}",
			},
			setupCtx: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/users?status=active", nil)
				return c
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("result is not a map")
				}
				if m["filter"] != "active" {
					t.Errorf("filter = %v, want active", m["filter"])
				}
			},
		},
		{
			name: "template with service info",
			tmpl: map[string]interface{}{
				"service": "{{.Service.Name}}",
				"version": "{{.Service.Version}}",
			},
			setupCtx: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/test", nil)
				return c
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("result is not a map")
				}
				if m["service"] != "test-service" {
					t.Errorf("service = %v, want test-service", m["service"])
				}
				if m["version"] != "v1.0.0" {
					t.Errorf("version = %v, want v1.0.0", m["version"])
				}
			},
		},
		{
			name: "template with RandomID",
			tmpl: map[string]interface{}{
				"id": "{{.RandomID}}",
			},
			setupCtx: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/test", nil)
				return c
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("result is not a map")
				}
				id, ok := m["id"].(string)
				if !ok || len(id) == 0 {
					t.Error("RandomID not generated")
				}
			},
		},
		{
			name: "template with Timestamp",
			tmpl: map[string]interface{}{
				"timestamp": "{{.Timestamp}}",
			},
			setupCtx: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/test", nil)
				return c
			},
			wantErr: false,
			validate: func(t *testing.T, result interface{}) {
				m, ok := result.(map[string]interface{})
				if !ok {
					t.Fatal("result is not a map")
				}
				ts, ok := m["timestamp"].(string)
				if !ok || len(ts) == 0 {
					t.Error("Timestamp not generated")
				}
			},
		},
		{
			name: "invalid template syntax",
			tmpl: map[string]interface{}{
				"invalid": "{{.InvalidField",
			},
			setupCtx: func() *gin.Context {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest("GET", "/test", nil)
				return c
			},
			wantErr:     true,
			errContains: "template parse error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := New("test-service", "v1.0.0")
			ctx := tt.setupCtx()

			result, err := renderer.Render(tt.tmpl, ctx)

			if tt.wantErr {
				if err == nil {
					t.Error("Render() expected error but got nil")
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Render() error = %v, want error containing %q", err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("Render() unexpected error = %v", err)
				}
				if tt.validate != nil {
					tt.validate(t, result)
				}
			}
		})
	}
}

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		name   string
		params gin.Params
		want   map[string]string
	}{
		{
			name:   "no params",
			params: gin.Params{},
			want:   map[string]string{},
		},
		{
			name: "single param",
			params: gin.Params{
				{Key: "id", Value: "123"},
			},
			want: map[string]string{"id": "123"},
		},
		{
			name: "multiple params",
			params: gin.Params{
				{Key: "id", Value: "123"},
				{Key: "name", Value: "alice"},
			},
			want: map[string]string{
				"id":   "123",
				"name": "alice",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = tt.params

			got := extractPathParams(c)

			if len(got) != len(tt.want) {
				t.Errorf("extractPathParams() length = %d, want %d", len(got), len(tt.want))
			}

			for key, wantValue := range tt.want {
				if gotValue, ok := got[key]; !ok || gotValue != wantValue {
					t.Errorf("extractPathParams()[%s] = %v, want %v", key, gotValue, wantValue)
				}
			}
		})
	}
}

func TestExtractQueryParams(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want map[string]string
	}{
		{
			name: "no params",
			url:  "/test",
			want: map[string]string{},
		},
		{
			name: "single param",
			url:  "/test?status=active",
			want: map[string]string{"status": "active"},
		},
		{
			name: "multiple params",
			url:  "/test?status=active&page=2",
			want: map[string]string{
				"status": "active",
				"page":   "2",
			},
		},
		{
			name: "param with multiple values uses first",
			url:  "/test?tags=go&tags=rust",
			want: map[string]string{"tags": "go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", tt.url, nil)

			got := extractQueryParams(c)

			if len(got) != len(tt.want) {
				t.Errorf("extractQueryParams() length = %d, want %d", len(got), len(tt.want))
			}

			for key, wantValue := range tt.want {
				if gotValue, ok := got[key]; !ok || gotValue != wantValue {
					t.Errorf("extractQueryParams()[%s] = %v, want %v", key, gotValue, wantValue)
				}
			}
		})
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
