package handler

import (
	"log/slog"

	"github.com/blueship581/gbcheckup/internal/dto"
	"github.com/blueship581/gbcheckup/internal/service"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
)

// ExamResultHandler 检查结果接口。
type ExamResultHandler struct {
	svc *service.ExamResultService
	log *slog.Logger
}

// NewExamResultHandler 构造检查结果接口。
func NewExamResultHandler(svc *service.ExamResultService, log *slog.Logger) *ExamResultHandler {
	return &ExamResultHandler{svc: svc, log: log}
}

// Enter 录入结果。
func (h *ExamResultHandler) Enter(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req dto.EnterResultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("检查结果（ExamResult）参数不合法", err))
		return
	}
	res, err := h.svc.Enter(c.Request.Context(), id, userID(c), service.EnterInput{ResultValue: req.ResultValue, ResultText: req.ResultText, ImageURL: req.ImageURL})
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, res)
}

// Review 审核结果。
func (h *ExamResultHandler) Review(c *gin.Context) {
	id := parseUint(c.Param("id"))
	if err := h.svc.Review(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"id": id, "reviewed": true})
}

// ListByRegistration 登记下结果列表。
func (h *ExamResultHandler) ListByRegistration(c *gin.Context) {
	regID := parseUint(c.Query("registration_id"))
	items, err := h.svc.ListByRegistration(c.Request.Context(), regID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, items)
}

// ListPending 待录入工作台。
func (h *ExamResultHandler) ListPending(c *gin.Context) {
	page := parseQueryInt(c.Query("page"), 1)
	pageSize := parseQueryInt(c.Query("page_size"), 20)
	items, total, err := h.svc.ListPending(c.Request.Context(), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: items, Total: total, Page: page, Size: pageSize})
}
