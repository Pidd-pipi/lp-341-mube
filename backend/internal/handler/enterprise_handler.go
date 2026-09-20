package handler

import (
	"log/slog"

	"github.com/blueship581/gbcheckup/internal/dto"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/service"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
)

// EnterpriseHandler 团检接口。
type EnterpriseHandler struct {
	svc *service.EnterpriseService
	log *slog.Logger
}

// NewEnterpriseHandler 构造团检接口。
func NewEnterpriseHandler(svc *service.EnterpriseService, log *slog.Logger) *EnterpriseHandler {
	return &EnterpriseHandler{svc: svc, log: log}
}

// CreateEnterprise 创建企业。
func (h *EnterpriseHandler) CreateEnterprise(c *gin.Context) {
	var req dto.EnterpriseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("团检企业（Enterprise）参数不合法", err))
		return
	}
	e := &model.Enterprise{Name: req.Name, Contact: req.Contact, Phone: req.Phone, Address: req.Address}
	created, err := h.svc.CreateEnterprise(c.Request.Context(), e)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, created)
}

// ListEnterprises 企业列表。
func (h *EnterpriseHandler) ListEnterprises(c *gin.Context) {
	page := parseQueryInt(c.Query("page"), 1)
	pageSize := parseQueryInt(c.Query("page_size"), 20)
	items, total, err := h.svc.ListEnterprises(c.Request.Context(), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: items, Total: total, Page: page, Size: pageSize})
}

// CreateOrder 创建团检订单。
func (h *EnterpriseHandler) CreateOrder(c *gin.Context) {
	var req dto.GroupOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("团检订单（GroupOrder）参数不合法", err))
		return
	}
	order, err := h.svc.CreateOrder(c.Request.Context(), req.EnterpriseID, req.PackageID, req.ExamineeCount)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, order)
}

// ListOrders 团检订单列表。
func (h *EnterpriseHandler) ListOrders(c *gin.Context) {
	page := parseQueryInt(c.Query("page"), 1)
	pageSize := parseQueryInt(c.Query("page_size"), 20)
	items, total, err := h.svc.ListOrders(c.Request.Context(), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: items, Total: total, Page: page, Size: pageSize})
}

// DeliverReports 报告批量交付。
func (h *EnterpriseHandler) DeliverReports(c *gin.Context) {
	id := parseUint(c.Param("id"))
	order, err := h.svc.DeliverReports(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, order)
}
