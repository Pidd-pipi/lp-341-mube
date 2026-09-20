package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OK 统一成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": data})
}

// Created 创建成功响应。
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "ok", "data": data})
}

// PageData 分页数据。
type PageData struct {
	List  any   `json:"list"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"page_size"`
}
