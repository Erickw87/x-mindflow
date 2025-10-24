package handler

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/LiusCraft/x-mindflow/mock-service/pkg/config"
	"github.com/LiusCraft/x-mindflow/mock-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

type proxyHandler struct {
	logger logger.Logger
}

// NewProxyHandler 创建代理处理器
func NewProxyHandler(log logger.Logger) *proxyHandler {
	return &proxyHandler{
		logger: log,
	}
}

func (h *proxyHandler) Handle(c *gin.Context, cfg *config.ProxyConfig) {
	targetURL := cfg.Target
	if cfg.Path != "" {
		targetURL = cfg.Target + cfg.Path
	} else {
		targetURL = cfg.Target + c.Request.URL.Path
	}

	query := c.Request.URL.Query()
	for k, v := range cfg.QueryParams {
		query.Set(k, v)
	}
	if len(query) > 0 {
		targetURL += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if c.Request.Body != nil {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(c.Request.Method, targetURL, bodyReader)
	if err != nil {
		h.logger.Error("Failed to create proxy request", "error", err, "target", targetURL)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create proxy request"})
		return
	}

	h.copyHeaders(c.Request.Header, req.Header, cfg.PreserveHeaders)

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	h.logger.Debug("Proxying request",
		"method", c.Request.Method,
		"target", targetURL,
		"request_id", c.GetHeader("X-Request-ID"))

	resp, err := client.Do(req)
	if err != nil {
		h.logger.Warn("Proxy request failed", "error", err, "target", targetURL)
		c.JSON(http.StatusBadGateway, gin.H{"error": "proxy request failed", "details": err.Error()})
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	c.Status(resp.StatusCode)
	io.Copy(c.Writer, resp.Body)

	h.logger.Debug("Proxy request completed",
		"target", targetURL,
		"status", resp.StatusCode)
}

func (h *proxyHandler) copyHeaders(src, dst http.Header, preserveAll bool) {
	mustCopy := []string{
		"X-Request-ID",
		"X-Traffic-Tag",
		"X-Forwarded-For",
		"Content-Type",
		"Authorization",
		"User-Agent",
	}

	for _, key := range mustCopy {
		if value := src.Get(key); value != "" {
			dst.Set(key, value)
		}
	}

	if preserveAll {
		for key, values := range src {
			if strings.HasPrefix(key, "X-") {
				for _, value := range values {
					dst.Add(key, value)
				}
			}
		}
	}
}
