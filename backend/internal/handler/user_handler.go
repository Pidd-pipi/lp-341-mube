package handler

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/blueship581/gbcheckup/internal/dto"
	"github.com/blueship581/gbcheckup/internal/middleware"
	"github.com/blueship581/gbcheckup/internal/service"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
)

// UserHandler 用户接口。
type UserHandler struct {
	svc *service.UserService
	log *slog.Logger
}

// NewUserHandler 构造用户接口。
func NewUserHandler(svc *service.UserService, log *slog.Logger) *UserHandler {
	return &UserHandler{svc: svc, log: log}
}

func userID(c *gin.Context) uint {
	v, _ := c.Get(middleware.UserIDKey)
	return v.(uint)
}

func parseUint(s string) uint {
	v, _ := strconv.ParseUint(s, 10, 64)
	return uint(v)
}

func parseQueryInt(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil && v > 0 {
		return v
	}
	return def
}

// Register 注册。
func (h *UserHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("注册参数（User）不合法", err))
		return
	}
	user, token, err := h.svc.Register(c.Request.Context(), req.Phone, req.Password, req.Name)
	if err != nil {
		c.Error(err)
		return
	}
	util.Created(c, dto.TokenResponse{Token: token, User: user})
}

// Login 登录。
func (h *UserHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("登录参数（User）不合法", err))
		return
	}
	user, token, err := h.svc.Login(c.Request.Context(), req.Phone, req.Password)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, dto.TokenResponse{Token: token, User: user})
}

// Me 当前用户。
func (h *UserHandler) Me(c *gin.Context) {
	user, err := h.svc.GetByID(c.Request.Context(), userID(c))
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, user)
}

// UpdateProfile 修改资料。
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(util.BadRequest("资料参数（User）不合法", err))
		return
	}
	user, err := h.svc.UpdateProfile(c.Request.Context(), userID(c), req.Name, req.Avatar, req.Department)
	if err != nil {
		c.Error(err)
		return
	}
	util.OK(c, user)
}

var _ = errors.New
