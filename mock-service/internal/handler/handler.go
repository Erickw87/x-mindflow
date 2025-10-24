package handler

import (
	"github.com/LiusCraft/x-mindflow/mock-service/pkg/config"
	"github.com/gin-gonic/gin"
)

// Handler 请求处理器接口
type Handler interface {
	HandleStaticResponse(c *gin.Context, cfg *config.ResponseConfig)
	HandleProxy(c *gin.Context, cfg *config.ProxyConfig)
	HandleRandomResponse(c *gin.Context, responses []config.ResponseConfig)
}

type compositeHandler struct {
	static *staticHandler
	proxy  *proxyHandler
}

// New 创建请求处理器
func New(static *staticHandler, proxy *proxyHandler) Handler {
	return &compositeHandler{
		static: static,
		proxy:  proxy,
	}
}

func (h *compositeHandler) HandleStaticResponse(c *gin.Context, cfg *config.ResponseConfig) {
	h.static.Handle(c, cfg)
}

func (h *compositeHandler) HandleProxy(c *gin.Context, cfg *config.ProxyConfig) {
	h.proxy.Handle(c, cfg)
}

func (h *compositeHandler) HandleRandomResponse(c *gin.Context, responses []config.ResponseConfig) {
	h.static.HandleRandom(c, responses)
}
