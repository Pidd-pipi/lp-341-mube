package handler

import (
	"log/slog"

	"github.com/blueship581/gbcheckup/internal/dto"
	"github.com/blueship581/gbcheckup/internal/service"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
)

// ReportHandler 报告接口。
type ReportHandler struct {
	svc *service.ReportService
	log *slog.Logger
}

// NewReportHandler 构造报告接口。
func NewReportHandler(svc *service.ReportService, log *slog.Logger) *ReportHandler {
	return &ReportHandler{svc: svc, log: log}
}

// Draft 创建/获取草稿。
func (h *ReportHandler) Draft(c *gin.Context) {
	regID := parseUint(c.Query("registration_id"))
	report, err := h.svc.DraftOrGet(c.Request.Context(), regID)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, report)
}

// Generate 生成报告。
func (h *ReportHandler) Generate(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req dto.ReportContentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("报告内容（Report）参数不合法", err))
		return
	}
	report, err := h.svc.Generate(c.Request.Context(), id, userID(c), req.Conclusion, req.HealthAdvice, req.FollowUpReminder)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, report)
}

// Review 审核报告。
func (h *ReportHandler) Review(c *gin.Context) {
	id := parseUint(c.Param("id"))
	report, err := h.svc.Review(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, report)
}

// Publish 发布报告。
func (h *ReportHandler) Publish(c *gin.Context) {
	id := parseUint(c.Param("id"))
	report, err := h.svc.Publish(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, report)
}

// List 报告列表。
func (h *ReportHandler) List(c *gin.Context) {
	page := parseQueryInt(c.Query("page"), 1)
	pageSize := parseQueryInt(c.Query("page_size"), 20)
	items, total, err := h.svc.List(c.Request.Context(), c.Query("status"), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: items, Total: total, Page: page, Size: pageSize})
}

// Detail 报告详情。
func (h *ReportHandler) Detail(c *gin.Context) {
	id := parseUint(c.Param("id"))
	report, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, report)
}

// DownloadPDF 下载报告 PDF。
func (h *ReportHandler) DownloadPDF(c *gin.Context) {
	id := parseUint(c.Param("id"))
	content, err := h.svc.PDFBytes(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", `attachment; filename="report.pdf"`)
	c.Data(200, "application/pdf", content)
}
