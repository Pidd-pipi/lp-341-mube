package handler

import (
	"log/slog"

	"github.com/blueship581/gbcheckup/internal/service"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
)

// StatsHandler 统计接口。
type StatsHandler struct {
	svc *service.StatsService
	log *slog.Logger
}

// NewStatsHandler 构造统计接口。
func NewStatsHandler(svc *service.StatsService, log *slog.Logger) *StatsHandler {
	return &StatsHandler{svc: svc, log: log}
}

// Dashboard 运营统计。
func (h *StatsHandler) Dashboard(c *gin.Context) {
	stats, err := h.svc.Dashboard(c.Request.Context())
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, stats)
}
