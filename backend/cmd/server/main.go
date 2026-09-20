package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blueship581/gbcheckup/internal/config"
	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/handler"
	"github.com/blueship581/gbcheckup/internal/middleware"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
	"github.com/blueship581/gbcheckup/internal/router"
	"github.com/blueship581/gbcheckup/internal/service"
	"github.com/blueship581/gbcheckup/internal/util"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}
	log := util.NewLogger()

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)})
	if err != nil {
		panic(fmt.Errorf("open database: %w", err))
	}
	if err := migrateAndSeed(db, log); err != nil {
		panic(fmt.Errorf("migrate database: %w", err))
	}
	log.Info(constants.LOG_DB_INITIALIZED)

	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Warn("mkdir upload dir", "error", err)
	}
	if err := os.MkdirAll("/app/reports", 0o755); err != nil {
		log.Warn("mkdir reports dir", "error", err)
	}

	userRepo := repository.NewUserRepository(db)
	pkgRepo := repository.NewPackageRepository(db)
	itemRepo := repository.NewPackageItemRepository(db)
	examineeRepo := repository.NewExamineeRepository(db)
	regRepo := repository.NewRegistrationRepository(db)
	resultRepo := repository.NewExamResultRepository(db)
	reportRepo := repository.NewReportRepository(db)
	metricRepo := repository.NewAbnormalMetricRepository(db)
	entRepo := repository.NewEnterpriseRepository(db)
	orderRepo := repository.NewGroupOrderRepository(db)

	userSvc := service.NewUserService(userRepo, cfg.JWTSecret, cfg.JWTExpireHours, log)
	pkgSvc := service.NewPackageService(pkgRepo, itemRepo, log)
	examineeSvc := service.NewExamineeService(examineeRepo, log)
	regSvc := service.NewRegistrationService(regRepo, examineeRepo, pkgRepo, itemRepo, resultRepo, log)
	resultSvc := service.NewExamResultService(resultRepo, regRepo, metricRepo, log)
	reportSvc := service.NewReportService(reportRepo, resultRepo, regRepo, log)
	metricSvc := service.NewAbnormalMetricService(metricRepo, log)
	entSvc := service.NewEnterpriseService(entRepo, orderRepo, pkgRepo, log)
	statsSvc := service.NewStatsService(pkgRepo, regRepo, reportRepo, resultRepo, metricRepo, itemRepo, log)

	h := router.Handlers{
		User:         handler.NewUserHandler(userSvc, log),
		Package:      handler.NewPackageHandler(pkgSvc, log),
		Examinee:     handler.NewExamineeHandler(examineeSvc, log),
		Registration: handler.NewRegistrationHandler(regSvc, log),
		ExamResult:   handler.NewExamResultHandler(resultSvc, log),
		Report:       handler.NewReportHandler(reportSvc, log),
		Abnormal:     handler.NewAbnormalMetricHandler(metricSvc, log),
		Enterprise:   handler.NewEnterpriseHandler(entSvc, log),
		Stats:        handler.NewStatsHandler(statsSvc, log),
	}
	r := router.New(cfg, log, h, middleware.NewRateLimiter(cfg.RateLimitPerMin), cfg.UploadDir)

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Info(constants.LOG_SERVER_STARTED, "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(constants.LOG_SERVER_STARTED, "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error(constants.LOG_SERVER_SHUTDOWN, "error", err)
	}
}

// migrateAndSeed 自动迁移并注入种子数据；init.sql 已建表则跳过。
func migrateAndSeed(db *gorm.DB, log *slog.Logger) error {
	var tableCount int64
	if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name='packages'").Scan(&tableCount).Error; err != nil {
		return err
	}
	if tableCount > 0 {
		return nil
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Package{}, &model.PackageItem{}, &model.Examinee{},
		&model.Registration{}, &model.ExamResult{}, &model.Report{}, &model.AbnormalMetric{},
		&model.Enterprise{}, &model.GroupOrder{},
	); err != nil {
		return err
	}
	seeds := []struct {
		phone string
		pass  string
		name  string
		role  string
		dept  string
	}{
		{"13800000001", "admin123", "系统管理员", constants.RoleAdmin, "管理部"},
		{"13800000002", "doctor123", "张医生", constants.RoleDoctor, "内科"},
		{"13800000003", "front123", "前台小李", constants.RoleFrontDesk, "导检台"},
		{"13800000004", "examinee123", "体检用户", constants.RoleExaminee, ""},
	}
	users := make([]model.User, 0, len(seeds))
	for _, s := range seeds {
		hash, _ := bcrypt.GenerateFromPassword([]byte(s.pass), bcrypt.DefaultCost)
		users = append(users, model.User{Phone: s.phone, PasswordHash: string(hash), Name: s.name, Role: s.role, Department: s.dept})
	}
	if err := db.Create(&users).Error; err != nil {
		return err
	}
	packages := []model.Package{
		{Name: "入职基础体检", PackageType: constants.PackageEntry, Price: 199, Status: constants.PackageActive, Description: "血常规、尿常规、肝功能、胸片"},
		{Name: "年度标准体检", PackageType: constants.PackageAnnual, Price: 599, Status: constants.PackageActive, Description: "内外科、血生化、腹部彩超、心电图"},
		{Name: "高端深度体检", PackageType: constants.PackagePremium, Price: 1999, Status: constants.PackageActive, Description: "含 CT、肿瘤标志物、心脑血管深度筛查"},
	}
	if err := db.Create(&packages).Error; err != nil {
		return err
	}
	items := []model.PackageItem{
		{PackageID: 1, ItemName: "血常规", ItemGroup: "检验", RefValueRange: "3.5-9.5", Department: "检验科", SortOrder: 1},
		{PackageID: 1, ItemName: "肝功能ALT", ItemGroup: "检验", RefValueRange: "9-50", Department: "检验科", SortOrder: 2},
		{PackageID: 1, ItemName: "胸片", ItemGroup: "影像", RefValueRange: "正常", Department: "放射科", SortOrder: 3},
		{PackageID: 2, ItemName: "空腹血糖", ItemGroup: "生化", RefValueRange: "3.9-6.1", Department: "检验科", SortOrder: 1},
		{PackageID: 2, ItemName: "心电图", ItemGroup: "心电", RefValueRange: "正常", Department: "功能科", SortOrder: 2},
		{PackageID: 3, ItemName: "胸部CT", ItemGroup: "影像", RefValueRange: "正常", Department: "放射科", SortOrder: 1},
		{PackageID: 3, ItemName: "肿瘤标志物CEA", ItemGroup: "检验", RefValueRange: "0-5", Department: "检验科", SortOrder: 2},
	}
	if err := db.Create(&items).Error; err != nil {
		return err
	}
	examinees := []model.Examinee{
		{Name: "王小明", IDCardNo: "110101199501011111", Phone: "13911110001", Gender: "male", Age: 30, SourceType: "personal"},
		{Name: "李小红", IDCardNo: "310101198812121222", Phone: "13911110002", Gender: "female", Age: 36, SourceType: "personal"},
		{Name: "赵大强", IDCardNo: "440101199003033333", Phone: "13911110003", Gender: "male", Age: 34, SourceType: "personal"},
	}
	if err := db.Create(&examinees).Error; err != nil {
		return err
	}
	regs := []model.Registration{
		{ExamineeID: 1, PackageID: 1, GuideNo: "GUIDE202608160001", Status: constants.RegistrationInProgress, RegisteredAt: time.Now(), RegisterUserID: 3},
		{ExamineeID: 2, PackageID: 2, GuideNo: "GUIDE202608160002", Status: constants.RegistrationRegistered, RegisteredAt: time.Now(), RegisterUserID: 3},
	}
	if err := db.Create(&regs).Error; err != nil {
		return err
	}
	now := time.Now()
	results := []model.ExamResult{
		{RegistrationID: 1, ExamineeID: 1, PackageItemID: 1, ResultValue: "5.2", Status: constants.ResultEntered, IsAbnormal: false, DoctorID: 2, EnteredAt: &now},
		{RegistrationID: 1, ExamineeID: 1, PackageItemID: 2, ResultValue: "72", Status: constants.ResultEntered, IsAbnormal: true, DoctorID: 2, EnteredAt: &now},
		{RegistrationID: 1, ExamineeID: 1, PackageItemID: 3, ResultValue: "正常", Status: constants.ResultPending, IsAbnormal: false},
		{RegistrationID: 2, ExamineeID: 2, PackageItemID: 4, ResultValue: "5.5", Status: constants.ResultPending, IsAbnormal: false},
		{RegistrationID: 2, ExamineeID: 2, PackageItemID: 5, ResultValue: "正常", Status: constants.ResultPending, IsAbnormal: false},
	}
	if err := db.Create(&results).Error; err != nil {
		return err
	}
	if err := db.Create(&model.AbnormalMetric{
		ExamineeID: 1, PackageItemID: 2, AbnormalLevel: constants.AbnormalModerate,
		Value: "72", RefValueRange: "9-50", TrendJSON: "[]", FollowUpStatus: constants.FollowUpPending,
	}).Error; err != nil {
		return err
	}
	if err := db.Create(&model.Report{
		RegistrationID: 1, ExamineeID: 1, ReportNo: "GB202608160001",
		Status: constants.ReportDraft, DoctorID: 2,
	}).Error; err != nil {
		return err
	}
	if err := db.Create(&model.Enterprise{Name: "华信科技", Contact: "陈经理", Phone: "021-88886666", Address: "上海市浦东新区"}).Error; err != nil {
		return err
	}
	if err := db.Create(&model.GroupOrder{EnterpriseID: 1, PackageID: 2, ExamineeCount: 50, Status: constants.GroupOrderConfirmed}).Error; err != nil {
		return err
	}
	log.Info(constants.LOG_DB_INITIALIZED, "seed", "ok")
	return nil
}
