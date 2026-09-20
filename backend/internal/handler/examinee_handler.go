package handler

import (
	"log/slog"

	"github.com/blueship581/gbcheckup/internal/dto"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/service"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
)

// ExamineeHandler 体检人接口。
type ExamineeHandler struct {
	svc *service.ExamineeService
	log *slog.Logger
}

// NewExamineeHandler 构造体检人接口。
func NewExamineeHandler(svc *service.ExamineeService, log *slog.Logger) *ExamineeHandler {
	return &ExamineeHandler{svc: svc, log: log}
}

// Create 登记体检人。
func (h *ExamineeHandler) Create(c *gin.Context) {
	var req dto.ExamineeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("体检人（Examinee）参数不合法", err))
		return
	}
	e := &model.Examinee{Name: req.Name, IDCardNo: req.IDCardNo, Phone: req.Phone, Gender: req.Gender, Age: req.Age, SourceType: req.SourceType, EnterpriseID: req.EnterpriseID}
	if e.SourceType == "" {
		e.SourceType = "personal"
	}
	created, err := h.svc.Create(c.Request.Context(), e)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, created)
}

// List 体检人列表。
func (h *ExamineeHandler) List(c *gin.Context) {
	page := parseQueryInt(c.Query("page"), 1)
	pageSize := parseQueryInt(c.Query("page_size"), 20)
	items, total, err := h.svc.List(c.Request.Context(), c.Query("keyword"), page, pageSize)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, util.PageData{List: items, Total: total, Page: page, Size: pageSize})
}

// Detail 体检人详情。
func (h *ExamineeHandler) Detail(c *gin.Context) {
	id := parseUint(c.Param("id"))
	e, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, e)
}

// BatchImport 团体批量导入。
func (h *ExamineeHandler) BatchImport(c *gin.Context) {
	var req dto.BatchImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("团体导入（Examinee）参数不合法", err))
		return
	}
	count, items, err := h.svc.BatchImport(c.Request.Context(), req.EnterpriseID, req.CSVText)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, gin.H{"imported": count, "items": items})
}
