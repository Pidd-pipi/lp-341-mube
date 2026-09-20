package router

import (
	"log/slog"
	"net/http"

	"github.com/blueship581/gbcheckup/internal/config"
	"github.com/blueship581/gbcheckup/internal/handler"
	"github.com/blueship581/gbcheckup/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Handlers 全部接口处理器集合。
type Handlers struct {
	User         *handler.UserHandler
	Package      *handler.PackageHandler
	Examinee     *handler.ExamineeHandler
	Registration *handler.RegistrationHandler
	ExamResult   *handler.ExamResultHandler
	Report       *handler.ReportHandler
	Abnormal     *handler.AbnormalMetricHandler
	Enterprise   *handler.EnterpriseHandler
	Stats        *handler.StatsHandler
}

// New 装配 Gin 路由。
func New(cfg config.Config, log *slog.Logger, h Handlers, limiter *middleware.RateLimiter, uploadDir string) *gin.Engine {
	allowOrigins := cfg.CORSOriginsList()
	if len(allowOrigins) == 0 {
		allowOrigins = []string{"http://localhost:18941"}
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.ErrorHandler(log))
	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger(log))
	r.Use(middleware.UploadMiddleware())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		AllowCredentials: true,
	}))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"status": "up"}})
	})
	r.GET("/api/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": gin.H{"status": "up"}})
	})
	r.Static("/uploads", uploadDir)
	r.Static("/reports", "/app/reports")

	auth := middleware.AuthRequired(cfg.JWTSecret)
	rate := limiter.Limit()

	api := r.Group("/api/v1")
	api.POST("/auth/register", rate, h.User.Register)
	api.POST("/auth/login", rate, h.User.Login)

	authed := api.Group("")
	authed.Use(auth, rate)
	registerUserRoutes(authed, h)
	registerPackageRoutes(authed, h)
	registerExamineeRoutes(authed, h)
	registerRegistrationRoutes(authed, h)
	registerExamResultRoutes(authed, h)
	registerReportRoutes(authed, h)
	registerAbnormalMetricRoutes(authed, h)
	registerEnterpriseRoutes(authed, h)
	registerStatsRoutes(authed, h)
	return r
}
