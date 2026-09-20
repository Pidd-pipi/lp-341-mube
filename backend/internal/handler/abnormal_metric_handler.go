package handler

import (
	"log/slog"

	"github.com/blueship581/gbcheckup/internal/dto"
	"github.com/blueship581/gbcheckup/internal/service"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
)

// AbnormalMetricHandler 异常指标接口。
type AbnormalMetricHandler struct {
	svc *service.AbnormalMetricService
	log *slog.Logger
}

// NewAbnormalMetricHandler 构造异常指标接口。
func NewAbnormalMetricHandler(svc *service.AbnormalMetricService, log *slog.Logger) *AbnormalMetricHandler {
	return &AbnormalMetricHandler{svc: svc, log: log}
}

// List 异常指标列表。
func (h *AbnormalMetricHandler) List(c *gin.Context) {
	page := parseQueryInt(c.Query("page"), 1)
	pageSize := parseQueryInt(c.Query("page_size"), 20)
	items, total, err := h.svc.List(c.Request.Context(), parseUint(c.Query("examinee_id")), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: items, Total: total, Page: page, Size: pageSize})
}

// UpdateFollowUp 更新复查跟踪。
func (h *AbnormalMetricHandler) UpdateFollowUp(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req dto.FollowUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("复查跟踪（AbnormalMetric）参数不合法", err))
		return
	}
	item, err := h.svc.UpdateFollowUp(c.Request.Context(), id, req.Status, req.Advice)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, item)
}
