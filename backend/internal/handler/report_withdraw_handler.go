package handler

import (
	"fmt"
	"log/slog"

	"github.com/blueship581/gbcheckup/internal/dto"
	"github.com/blueship581/gbcheckup/internal/service"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
)

// ReportWithdrawHandler 报告撤回重签接口。
type ReportWithdrawHandler struct {
	svc *service.ReportWithdrawService
	log *slog.Logger
}

// NewReportWithdrawHandler 构造撤回重签接口处理器。
func NewReportWithdrawHandler(svc *service.ReportWithdrawService, log *slog.Logger) *ReportWithdrawHandler {
	return &ReportWithdrawHandler{svc: svc, log: log}
}

// Apply 申请撤回（报告详情页触发）。
func (h *ReportWithdrawHandler) Apply(c *gin.Context) {
	reportID := parseUint(c.Param("id"))
	var req dto.ReportWithdrawApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest(
			fmt.Sprintf("Report[id=%d] withdraw apply 参数（ReportWithdraw.reason）不合法", reportID), err))
		return
	}
	w, err := h.svc.Apply(c.Request.Context(), reportID, userID(c), req.Reason)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, w)
}

// Approve 管理员批准撤回。
func (h *ReportWithdrawHandler) Approve(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req dto.ReportWithdrawReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest(
			fmt.Sprintf("ReportWithdraw[id=%d] approve 参数不合法", id), err))
		return
	}
	res, err := h.svc.Approve(c.Request.Context(), id, userID(c), req.Comment)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, res)
}

// Reject 管理员驳回撤回。
func (h *ReportWithdrawHandler) Reject(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req dto.ReportWithdrawReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest(
			fmt.Sprintf("ReportWithdraw[id=%d] reject 参数不合法", id), err))
		return
	}
	w, err := h.svc.Reject(c.Request.Context(), id, userID(c), req.Comment)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, w)
}

// Detail 撤回申请详情。
func (h *ReportWithdrawHandler) Detail(c *gin.Context) {
	id := parseUint(c.Param("id"))
	w, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, w)
}

// List 撤回申请列表（管理员审批台，可按 status 过滤）。
func (h *ReportWithdrawHandler) List(c *gin.Context) {
	page := parseQueryInt(c.Query("page"), 1)
	pageSize := parseQueryInt(c.Query("page_size"), 20)
	items, total, err := h.svc.List(c.Request.Context(), c.Query("status"), parseUint(c.Query("report_id")), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: items, Total: total, Page: page, Size: pageSize})
}

// Chain 报告版本链（含撤回记录）。
func (h *ReportWithdrawHandler) Chain(c *gin.Context) {
	id := parseUint(c.Param("id"))
	chain, err := h.svc.Chain(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, chain)
}
