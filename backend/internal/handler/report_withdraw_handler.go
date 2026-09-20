package handler

import (
	"log/slog"

	"github.com/blueship581/gbcheckup/internal/dto"
	"github.com/blueship581/gbcheckup/internal/service"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
)

// ReportWithdrawHandler 报告撤回重签接口：申请、审批、版本链。
type ReportWithdrawHandler struct {
	svc *service.ReportWithdrawService
	log *slog.Logger
}

// NewReportWithdrawHandler 构造撤回重签接口。
func NewReportWithdrawHandler(svc *service.ReportWithdrawService, log *slog.Logger) *ReportWithdrawHandler {
	return &ReportWithdrawHandler{svc: svc, log: log}
}

// Request 申请撤回重签（admin/doctor）。
func (h *ReportWithdrawHandler) Request(c *gin.Context) {
	reportID := parseUint(c.Param("id"))
	var req dto.ReportWithdrawRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("撤回申请（ReportWithdrawRequest）参数不合法", err))
		return
	}
	created, err := h.svc.Request(c.Request.Context(), reportID, userID(c), req.Reason)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, created)
}

// Approve 管理员批准撤回。
func (h *ReportWithdrawHandler) Approve(c *gin.Context) {
	requestID := parseUint(c.Param("requestId"))
	var req dto.ReportWithdrawReviewRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("撤回审批（ReportWithdrawRequest）参数不合法", err))
		return
	}
	result, err := h.svc.Approve(c.Request.Context(), requestID, userID(c), req.Comment)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}

// Reject 管理员驳回撤回。
func (h *ReportWithdrawHandler) Reject(c *gin.Context) {
	requestID := parseUint(c.Param("requestId"))
	var req dto.ReportWithdrawReviewRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("撤回审批（ReportWithdrawRequest）参数不合法", err))
		return
	}
	result, err := h.svc.Reject(c.Request.Context(), requestID, userID(c), req.Comment)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, result)
}

// VersionChain 查看报告版本链（admin/doctor）。
func (h *ReportWithdrawHandler) VersionChain(c *gin.Context) {
	reportID := parseUint(c.Param("id"))
	chain, err := h.svc.VersionChain(c.Request.Context(), reportID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, chain)
}
