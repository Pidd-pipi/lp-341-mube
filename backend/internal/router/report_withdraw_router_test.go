package router_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
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
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type apiResp struct {
	Code int             `json:"code"`
	Msg  string          `json:"message"`
	Data json.RawMessage `json:"data"`
}

// newEngine 用 SQLite + 完整 Gin 路由装配端到端测试环境，返回数据库句柄与引擎。
func newEngine(t *testing.T) (*gorm.DB, http.Handler) {
	t.Helper()
	// 共享缓存模式确保本测试拿到的 *gorm.DB 与路由内仓储访问同一个内存库。
	dsn := fmt.Sprintf("file:e2e-%p?mode=memory&cache=shared", t)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Package{}, &model.PackageItem{}, &model.Examinee{},
		&model.Registration{}, &model.ExamResult{}, &model.Report{}, &model.ReportWithdraw{},
		&model.AbnormalMetric{}, &model.Enterprise{}, &model.GroupOrder{},
	); err != nil {
		t.Fatal(err)
	}
	if err := repository.NewReportWithdrawRepository(db).CreatePendingIndex(); err != nil {
		t.Fatal(err)
	}
	// 外键依赖的基础数据。
	if err := db.Create(&model.Examinee{Name: "测试人", IDCardNo: "310101199001019999"}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Registration{ExamineeID: 1, GuideNo: "GUIDE-E2E-0001", Status: constants.RegistrationCompleted}).Error; err != nil {
		t.Fatal(err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	reportDir := t.TempDir()
	reportRepo := repository.NewReportRepository(db)
	withdrawRepo := repository.NewReportWithdrawRepository(db)
	reportSvc := service.NewReportService(reportRepo,
		repository.NewExamResultRepository(db), repository.NewRegistrationRepository(db), log)
	reportSvc.SetReportDir(reportDir)
	withdrawSvc := service.NewReportWithdrawService(withdrawRepo, reportRepo, log)

	h := router.Handlers{
		Report:   handler.NewReportHandler(reportSvc, log),
		Withdraw: handler.NewReportWithdrawHandler(withdrawSvc, log),
	}
	cfg := config.Config{JWTSecret: "test-secret", RateLimitPerMin: 100000}
	return db, router.New(cfg, log, h, middleware.NewRateLimiter(cfg.RateLimitPerMin), t.TempDir(), reportDir)
}

func authToken(t *testing.T, userID uint, role string) string {
	t.Helper()
	tok, err := util.GenerateToken("test-secret", 24, userID, "13800000000", role)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func doJSON(t *testing.T, h http.Handler, method, path, tok string, body any) (int, apiResp) {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, rdr)
	req.Header.Set("Content-Type", "application/json")
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var resp apiResp
	if rec.Body.Len() > 0 {
		_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	}
	return rec.Code, resp
}

func seedPublished(t *testing.T, db *gorm.DB, no string, publishedAt time.Time) *model.Report {
	t.Helper()
	r := &model.Report{
		RegistrationID: 1, ExamineeID: 1, ReportNo: no,
		Status: constants.ReportPublished, Conclusion: "正常", PDFURL: "/reports/x.pdf",
		VersionNo: 1, PublishedAt: &publishedAt,
	}
	if err := db.Create(r).Error; err != nil {
		t.Fatal(err)
	}
	return r
}

// TestHTTP_WithdrawClosedLoop 真实路由：申请→冻结下载→重复/越权拦截→管理员批准→待重签版本→版本链。
func TestHTTP_WithdrawClosedLoop(t *testing.T) {
	db, h := newEngine(t)
	admin := authToken(t, 1, constants.RoleAdmin)
	doctor := authToken(t, 2, constants.RoleDoctor)
	report := seedPublished(t, db, "GB-HTTP-0001", time.Now().Add(-2*time.Hour))

	// 医生申请撤回 → 201。
	status, resp := doJSON(t, h, http.MethodPost, fmt.Sprintf("/api/v1/reports/%d/withdraw", report.ID), doctor,
		map[string]string{"reason": "结论有误需重签"})
	if status != http.StatusCreated || resp.Code != 0 {
		t.Fatalf("apply status=%d resp=%+v", status, resp)
	}
	var withdraw model.ReportWithdraw
	_ = json.Unmarshal(resp.Data, &withdraw)
	if withdraw.Status != constants.WithdrawPending {
		t.Fatalf("withdraw status = %s", withdraw.Status)
	}

	// 审批期间下载冻结：409。
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/pdf", report.ID), nil)
	req.Header.Set("Authorization", "Bearer "+doctor)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("pdf during withdrawing code = %d, want 409", rec.Code)
	}

	// 重复申请：409。
	if code, _ := doJSON(t, h, http.MethodPost, fmt.Sprintf("/api/v1/reports/%d/withdraw", report.ID), doctor,
		map[string]string{"reason": "再次申请"}); code != http.StatusConflict {
		t.Fatalf("duplicate apply code = %d, want 409", code)
	}

	// 医生无权审批：403。
	if code, _ := doJSON(t, h, http.MethodPost, fmt.Sprintf("/api/v1/reports/withdraws/%d/approve", withdraw.ID), doctor,
		map[string]any{"approve": true}); code != http.StatusForbidden {
		t.Fatalf("doctor approve code = %d, want 403", code)
	}

	// 管理员批准：原版 withdrawn、生成 V2 草稿。
	status, resp = doJSON(t, h, http.MethodPost, fmt.Sprintf("/api/v1/reports/withdraws/%d/approve", withdraw.ID), admin,
		map[string]any{"approve": true, "comment": "同意"})
	if status != http.StatusOK || resp.Code != 0 {
		t.Fatalf("approve status=%d resp=%+v", status, resp)
	}
	var approveBody struct {
		Original     model.Report `json:"original"`
		ResignReport model.Report `json:"resign_report"`
	}
	_ = json.Unmarshal(resp.Data, &approveBody)
	if approveBody.Original.Status != constants.ReportWithdrawn {
		t.Fatalf("original status = %s, want withdrawn", approveBody.Original.Status)
	}
	if approveBody.ResignReport.Status != constants.ReportDraft || approveBody.ResignReport.VersionNo != 2 ||
		approveBody.ResignReport.ParentReportID != report.ID {
		t.Fatalf("resign report invalid: %+v", approveBody.ResignReport)
	}

	// 并发/重复审批：409 且状态不变。
	if code, _ := doJSON(t, h, http.MethodPost, fmt.Sprintf("/api/v1/reports/withdraws/%d/reject", withdraw.ID), admin,
		map[string]any{"approve": false}); code != http.StatusConflict {
		t.Fatalf("double review code = %d, want 409", code)
	}

	// 版本链两版。
	status, resp = doJSON(t, h, http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/versions", report.ID), doctor, nil)
	if status != http.StatusOK {
		t.Fatalf("versions status=%d", status)
	}
	var chain struct {
		Versions []model.Report `json:"versions"`
	}
	_ = json.Unmarshal(resp.Data, &chain)
	if len(chain.Versions) != 2 || chain.Versions[0].Status != constants.ReportWithdrawn {
		t.Fatalf("chain invalid: %+v", chain.Versions)
	}

	// 审批台列表可按 approved 过滤。
	if code, r2 := doJSON(t, h, http.MethodGet, "/api/v1/reports/withdraws?status=approved", admin, nil); code != http.StatusOK || r2.Code != 0 {
		t.Fatalf("list approved code=%d resp=%+v", code, r2)
	}
}

// TestHTTP_WithdrawRejectAndExpiry 超 24h 申请冲突；驳回后恢复已发布并恢复下载。
func TestHTTP_WithdrawRejectAndExpiry(t *testing.T) {
	db, h := newEngine(t)
	admin := authToken(t, 1, constants.RoleAdmin)
	doctor := authToken(t, 2, constants.RoleDoctor)

	expired := seedPublished(t, db, "GB-HTTP-OLD", time.Now().Add(-25*time.Hour))
	if code, _ := doJSON(t, h, http.MethodPost, fmt.Sprintf("/api/v1/reports/%d/withdraw", expired.ID), doctor,
		map[string]string{"reason": "超期撤回"}); code != http.StatusConflict {
		t.Fatalf("expired apply code = %d, want 409", code)
	}

	report := seedPublished(t, db, "GB-HTTP-NEW", time.Now().Add(-time.Hour))
	_, resp := doJSON(t, h, http.MethodPost, fmt.Sprintf("/api/v1/reports/%d/withdraw", report.ID), doctor,
		map[string]string{"reason": "申请撤回"})
	var w model.ReportWithdraw
	_ = json.Unmarshal(resp.Data, &w)
	if code, _ := doJSON(t, h, http.MethodPost, fmt.Sprintf("/api/v1/reports/withdraws/%d/reject", w.ID), admin,
		map[string]any{"approve": false, "comment": "驳回"}); code != http.StatusOK {
		t.Fatalf("reject code = %d", code)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/reports/%d/pdf", report.ID), nil)
	req.Header.Set("Authorization", "Bearer "+doctor)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code == http.StatusConflict {
		t.Fatalf("pdf still frozen after reject: %d", rec.Code)
	}
}
