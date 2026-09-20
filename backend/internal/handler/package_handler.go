package handler

import (
	"log/slog"

	"github.com/blueship581/gbcheckup/internal/dto"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/service"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
)

// PackageHandler 套餐接口。
type PackageHandler struct {
	svc *service.PackageService
	log *slog.Logger
}

// NewPackageHandler 构造套餐接口。
func NewPackageHandler(svc *service.PackageService, log *slog.Logger) *PackageHandler {
	return &PackageHandler{svc: svc, log: log}
}

// Create 创建套餐。
func (h *PackageHandler) Create(c *gin.Context) {
	var req dto.PackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("套餐（Package）参数不合法", err))
		return
	}
	pkg, err := h.svc.Create(c.Request.Context(), req.Name, req.PackageType, req.Price, req.Status, req.Description)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, pkg)
}

// List 套餐列表。
func (h *PackageHandler) List(c *gin.Context) {
	page := parseQueryInt(c.Query("page"), 1)
	pageSize := parseQueryInt(c.Query("page_size"), 20)
	pkgs, total, err := h.svc.List(c.Request.Context(), c.Query("status"), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: pkgs, Total: total, Page: page, Size: pageSize})
}

// Detail 套餐详情（含项目）。
func (h *PackageHandler) Detail(c *gin.Context) {
	id := parseUint(c.Param("id"))
	pkg, items, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"package": pkg, "items": items})
}

// Update 更新套餐。
func (h *PackageHandler) Update(c *gin.Context) {
	id := parseUint(c.Param("id"))
	var req dto.PackageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("套餐（Package）参数不合法", err))
		return
	}
	pkg, err := h.svc.Update(c.Request.Context(), id, req.Name, req.PackageType, req.Price, req.Status, req.Description)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, pkg)
}

// AddItem 添加检查项目。
func (h *PackageHandler) AddItem(c *gin.Context) {
	packageID := parseUint(c.Param("id"))
	var req dto.PackageItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("检查项目（PackageItem）参数不合法", err))
		return
	}
	item := &model.PackageItem{ItemName: req.ItemName, ItemGroup: req.ItemGroup, RefValueRange: req.RefValueRange, Department: req.Department, SortOrder: req.SortOrder}
	created, err := h.svc.AddItem(c.Request.Context(), packageID, item)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, created)
}

// ListItems 套餐项目列表。
func (h *PackageHandler) ListItems(c *gin.Context) {
	packageID := parseUint(c.Param("id"))
	items, err := h.svc.ListItems(c.Request.Context(), packageID)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, items)
}

// UpdateItem 更新检查项目。
func (h *PackageHandler) UpdateItem(c *gin.Context) {
	id := parseUint(c.Param("itemId"))
	var req dto.PackageItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("检查项目（PackageItem）参数不合法", err))
		return
	}
	item, err := h.svc.UpdateItem(c.Request.Context(), id, &model.PackageItem{ItemName: req.ItemName, ItemGroup: req.ItemGroup, RefValueRange: req.RefValueRange, Department: req.Department, SortOrder: req.SortOrder})
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, item)
}

// DeleteItem 删除检查项目。
func (h *PackageHandler) DeleteItem(c *gin.Context) {
	id := parseUint(c.Param("itemId"))
	if err := h.svc.DeleteItem(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}
	util.OK(c, gin.H{"deleted": id})
}
