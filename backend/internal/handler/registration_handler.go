package handler

import (
	"log/slog"

	"github.com/blueship581/gbcheckup/internal/dto"
	"github.com/blueship581/gbcheckup/internal/service"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
)

// RegistrationHandler 登记接口。
type RegistrationHandler struct {
	svc *service.RegistrationService
	log *slog.Logger
}

// NewRegistrationHandler 构造登记接口。
func NewRegistrationHandler(svc *service.RegistrationService, log *slog.Logger) *RegistrationHandler {
	return &RegistrationHandler{svc: svc, log: log}
}

// Register 体检登记。
func (h *RegistrationHandler) Register(c *gin.Context) {
	var req dto.RegisterRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("登记参数（Registration）不合法", err))
		return
	}
	reg, err := h.svc.Register(c.Request.Context(), req.ExamineeID, req.PackageID, userID(c))
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, reg)
}

// List 登记列表。
func (h *RegistrationHandler) List(c *gin.Context) {
	page := parseQueryInt(c.Query("page"), 1)
	pageSize := parseQueryInt(c.Query("page_size"), 20)
	items, total, err := h.svc.List(c.Request.Context(), c.Query("status"), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: items, Total: total, Page: page, Size: pageSize})
}

// Detail 登记详情。
func (h *RegistrationHandler) Detail(c *gin.Context) {
	id := parseUint(c.Param("id"))
	reg, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, reg)
}

// UpdateStatus 更新登记状态。
func (h *RegistrationHandler) UpdateStatus(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("登记状态（Registration.status）不合法", err))
		return
	}
	if err := h.svc.UpdateStatus(c.Request.Context(), id, req.Status); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"id": id, "status": req.Status})
}
