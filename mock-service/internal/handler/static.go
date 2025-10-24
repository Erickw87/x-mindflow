package handler

import (
	"math/rand"
	"net/http"

	"github.com/LiusCraft/x-mindflow/mock-service/internal/template"
	"github.com/LiusCraft/x-mindflow/mock-service/pkg/config"
	"github.com/gin-gonic/gin"
)

type staticHandler struct {
	renderer template.Renderer
}

// NewStaticHandler 创建静态响应处理器
func NewStaticHandler(renderer template.Renderer) *staticHandler {
	return &staticHandler{
		renderer: renderer,
	}
}

func (h *staticHandler) Handle(c *gin.Context, cfg *config.ResponseConfig) {
	for key, value := range cfg.Headers {
		c.Header(key, value)
	}

	body, err := h.renderer.Render(cfg.Body, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to render response", "details": err.Error()})
		return
	}

	c.JSON(cfg.Status, body)
}

func (h *staticHandler) HandleRandom(c *gin.Context, responses []config.ResponseConfig) {
	selected := h.selectRandomResponse(responses)
	h.Handle(c, &selected)
}

func (h *staticHandler) selectRandomResponse(responses []config.ResponseConfig) config.ResponseConfig {
	random := rand.Intn(100)
	cumulative := 0

	for _, resp := range responses {
		cumulative += resp.Weight
		if random < cumulative {
			return resp
		}
	}

	return responses[len(responses)-1]
}
