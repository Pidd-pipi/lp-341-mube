package middleware

import (
	"net/http"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
)

// MaxUploadSize 单文件大小上限（8MB）。
const MaxUploadSize = 8 << 20

// UploadMiddleware 上传限制中间件。
func UploadMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost || c.Request.Method == http.MethodPut {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxUploadSize)
		}
		c.Next()
	}
}

// SaveImage 保存上传的影像图片。
func SaveImage(c *gin.Context, dir string) (string, error) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		return "", util.BadRequest(constants.MsgParamInvalid, err)
	}
	defer file.Close()
	content, err := util.ReadAll(file)
	if err != nil {
		return "", util.BadRequest(constants.MsgParamInvalid, err)
	}
	return util.SaveUploadFile(dir, "image", content, header.Filename)
}
