package template

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Renderer 模板渲染器接口
type Renderer interface {
	Render(tmpl interface{}, ctx *gin.Context) (interface{}, error)
}

type defaultRenderer struct {
	serviceName    string
	serviceVersion string
}

// New 创建模板渲染器
func New(serviceName, serviceVersion string) Renderer {
	return &defaultRenderer{
		serviceName:    serviceName,
		serviceVersion: serviceVersion,
	}
}

func (r *defaultRenderer) Render(tmpl interface{}, ctx *gin.Context) (interface{}, error) {
	data := r.prepareTemplateData(ctx)

	jsonBytes, err := json.Marshal(tmpl)
	if err != nil {
		return nil, err
	}

	t, err := template.New("response").Parse(string(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("template parse error: %w (check your template syntax in response body)", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("template execution error: %w (available variables: PathParam, QueryParam, Service, RandomID, Timestamp)", err)
	}

	var result interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *defaultRenderer) prepareTemplateData(ctx *gin.Context) map[string]interface{} {
	return map[string]interface{}{
		"PathParam":  extractPathParams(ctx),
		"QueryParam": extractQueryParams(ctx),
		"Service": map[string]string{
			"Name":    r.serviceName,
			"Version": r.serviceVersion,
		},
		"RandomID":  uuid.New().String(),
		"Timestamp": time.Now().Format(time.RFC3339),
	}
}

func extractPathParams(ctx *gin.Context) map[string]string {
	params := make(map[string]string)
	for _, param := range ctx.Params {
		params[param.Key] = param.Value
	}
	return params
}

func extractQueryParams(ctx *gin.Context) map[string]string {
	params := make(map[string]string)
	for key, values := range ctx.Request.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params
}
