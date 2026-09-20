package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/handler"
	"github.com/blueship581/gbcheckup/internal/middleware"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
	"github.com/blueship581/gbcheckup/internal/service"
	"github.com/blueship581/gbcheckup/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupWithdrawRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&model.User{}, &model.Package{}, &model.PackageItem{}, &model.Examinee{},
		&model.Registration{}, &model.ExamResult{}, &model.Report{}, &model.ReportWithdrawRequest{},
		&model.AbnormalMetric{}, &model.Enterprise{}, &model.GroupOrder{},
	); err != nil {
		t.Fatal(err)
	}
	examinee := model.Examinee{Name: "钱七", IDCardNo: "110101199505051234"}
	if err := db.Create(&examinee).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	report := model.Report{
		ExamineeID: examinee.ID, ReportNo: "GB202609200099",
		Status: constants.ReportPublished, DoctorID: 2, Version: 1, PublishedAt: &now,
		Conclusion: "体检结论",
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatal(err)
	}

	reportRepo := repository.NewReportRepository(db)
	withdrawRepo := repository.NewReportWithdrawRepository(db)
	resultRepo := repository.NewExamResultRepository(db)
	regRepo := repository.NewRegistrationRepository(db)
	log := util.NewLogger()
	reportSvc := service.NewReportService(reportRepo, resultRepo, regRepo, t.TempDir(), log)
	withdrawSvc := service.NewReportWithdrawService(withdrawRepo, reportRepo, log)

	r := gin.New()
	r.Use(middleware.ErrorHandler(log))
	api := r.Group("/api/v1")
	api.Use(middleware.AuthRequired("test-secret"))
	h := Handlers{
		Report:         handler.NewReportHandler(reportSvc, log),
		ReportWithdraw: handler.NewReportWithdrawHandler(withdrawSvc, log),
	}
	registerReportRoutes(api, h)
	return r, db
}

func token(t *testing.T, userID uint, role string) string {
	t.Helper()
	tok, err := util.GenerateToken("test-secret", 24, userID, "13800000000", role)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func doRequest(t *testing.T, r *gin.Engine, method, path, jwtToken string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if jwtToken != "" {
		req.Header.Set("Authorization", "Bearer "+jwtToken)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestWithdrawHTTPLoop_FreezeConflictRBACAndVersionChain(t *testing.T) {
	r, db := setupWithdrawRouter(t)
	doctor := token(t, 2, constants.RoleDoctor)
	admin := token(t, 1, constants.RoleAdmin)

	// 1. 医生申请撤回成功
	w := doRequest(t, r, http.MethodPost, "/api/v1/reports/1/withdraw", doctor, map[string]string{"reason": "信息更正"})
	if w.Code != http.StatusCreated {
		t.Fatalf("withdraw request status = %d, want 201, body=%s", w.Code, w.Body.String())
	}
	var created struct {
		Data struct {
			ID     uint   `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.Data.Status != constants.WithdrawPending || created.Data.ID == 0 {
		t.Fatalf("unexpected withdraw payload: %s", w.Body.String())
	}

	// 2. 重复申请冲突
	if w := doRequest(t, r, http.MethodPost, "/api/v1/reports/1/withdraw", doctor, map[string]string{}); w.Code != http.StatusConflict {
		t.Fatalf("duplicate withdraw = %d, want 409, body=%s", w.Code, w.Body.String())
	}

	// 3. 审批期间下载冻结
	if w := doRequest(t, r, http.MethodGet, "/api/v1/reports/1/pdf", doctor, nil); w.Code != http.StatusConflict {
		t.Fatalf("pdf during withdrawing = %d, want 409", w.Code)
	}

	// 4. 医生无权审批（RBAC 嵌套管理员组）
	if w := doRequest(t, r, http.MethodPost, "/api/v1/reports/withdrawals/1/approve", doctor, map[string]string{}); w.Code != http.StatusForbidden {
		t.Fatalf("doctor approve = %d, want 403", w.Code)
	}

	// 5. 管理员批准成功
	if w := doRequest(t, r, http.MethodPost, "/api/v1/reports/withdrawals/1/approve", admin, map[string]string{"comment": "同意"}); w.Code != http.StatusOK {
		t.Fatalf("admin approve = %d, want 200, body=%s", w.Code, w.Body.String())
	}

	// 6. 重复/并发审批冲突且状态不变
	if w := doRequest(t, r, http.MethodPost, "/api/v1/reports/withdrawals/1/approve", admin, map[string]string{}); w.Code != http.StatusConflict {
		t.Fatalf("duplicate approve = %d, want 409", w.Code)
	}
	var original model.Report
	db.First(&original, 1)
	if original.Status != constants.ReportWithdrawn {
		t.Fatalf("original = %s, want withdrawn", original.Status)
	}
	var newCount int64
	db.Model(&model.Report{}).Where("parent_report_id = 1").Count(&newCount)
	if newCount != 1 {
		t.Fatalf("new versions = %d, want 1", newCount)
	}

	// 7. 版本链
	w = doRequest(t, r, http.MethodGet, "/api/v1/reports/1/versions", admin, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("version chain = %d, body=%s", w.Code, w.Body.String())
	}
	var chain struct {
		Data struct {
			Versions []model.Report `json:"versions"`
			Requests []struct {
				Status string `json:"status"`
			} `json:"requests"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &chain); err != nil {
		t.Fatal(err)
	}
	if len(chain.Data.Versions) != 2 {
		t.Fatalf("versions = %d, want 2", len(chain.Data.Versions))
	}
	if len(chain.Data.Requests) != 1 || chain.Data.Requests[0].Status != constants.WithdrawApproved {
		t.Fatalf("requests mismatch: %+v", chain.Data.Requests)
	}
}

func TestWithdrawHTTP_RejectRestoresPublished(t *testing.T) {
	r, db := setupWithdrawRouter(t)
	doctor := token(t, 2, constants.RoleDoctor)
	admin := token(t, 1, constants.RoleAdmin)

	if w := doRequest(t, r, http.MethodPost, "/api/v1/reports/1/withdraw", doctor, map[string]string{}); w.Code != http.StatusCreated {
		t.Fatalf("withdraw = %d body=%s", w.Code, w.Body.String())
	}
	if w := doRequest(t, r, http.MethodPost, "/api/v1/reports/withdrawals/1/reject", admin, map[string]string{"comment": "不符合"}); w.Code != http.StatusOK {
		t.Fatalf("reject = %d body=%s", w.Code, w.Body.String())
	}
	var report model.Report
	db.First(&report, 1)
	if report.Status != constants.ReportPublished {
		t.Fatalf("status after reject = %s, want published", report.Status)
	}
	// 驳回后窗口内可再次申请
	if w := doRequest(t, r, http.MethodPost, "/api/v1/reports/1/withdraw", doctor, map[string]string{}); w.Code != http.StatusCreated {
		t.Fatalf("re-request after reject = %d, want 201", w.Code)
	}
}

func TestWithdrawHTTP_UnauthenticatedRejected(t *testing.T) {
	r, _ := setupWithdrawRouter(t)
	if w := doRequest(t, r, http.MethodPost, "/api/v1/reports/1/withdraw", "", nil); w.Code != http.StatusUnauthorized {
		t.Fatalf("no token = %d, want 401", w.Code)
	}
}
