package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

func setupPublishedReport(t *testing.T, db *gorm.DB) *model.Report {
	t.Helper()
	examinee := model.Examinee{Name: "李四", IDCardNo: "310101199202021234"}
	if err := db.Create(&examinee).Error; err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	report := model.Report{
		RegistrationID: 1, ExamineeID: examinee.ID, ReportNo: "GB202609200001",
		Status: constants.ReportPublished, DoctorID: 2,
		Conclusion: "未见异常", Version: 1, PublishedAt: &now,
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatal(err)
	}
	return &report
}

func newWithdrawService(db *gorm.DB) *ReportWithdrawService {
	reportRepo := repository.NewReportRepository(db)
	withdrawRepo := repository.NewReportWithdrawRepository(db)
	return NewReportWithdrawService(withdrawRepo, reportRepo, testLogger())
}

func reloadReport(t *testing.T, db *gorm.DB, id uint) model.Report {
	t.Helper()
	var r model.Report
	if err := db.First(&r, id).Error; err != nil {
		t.Fatal(err)
	}
	return r
}

func TestWithdraw_RequestWithinWindow(t *testing.T) {
	db := newTestDB(t)
	svc := newWithdrawService(db)
	report := setupPublishedReport(t, db)

	req, err := svc.Request(context.Background(), report.ID, 2, "信息有误")
	if err != nil {
		t.Fatalf("Request() error = %v", err)
	}
	if req.Status != constants.WithdrawPending {
		t.Fatalf("request status = %s, want pending", req.Status)
	}
	got := reloadReport(t, db, report.ID)
	if got.Status != constants.ReportWithdrawing {
		t.Fatalf("report status = %s, want withdrawing", got.Status)
	}
}

func TestWithdraw_RejectRestoresPublished(t *testing.T) {
	db := newTestDB(t)
	svc := newWithdrawService(db)
	report := setupPublishedReport(t, db)
	ctx := context.Background()

	req, err := svc.Request(ctx, report.ID, 2, "信息有误")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Reject(ctx, req.ID, 1, "理由不充分"); err != nil {
		t.Fatalf("Reject() error = %v", err)
	}
	got := reloadReport(t, db, report.ID)
	if got.Status != constants.ReportPublished {
		t.Fatalf("report status = %s, want published after reject", got.Status)
	}
}

func TestWithdraw_ApproveCreatesResignVersionAndKeepsOriginal(t *testing.T) {
	db := newTestDB(t)
	svc := newWithdrawService(db)
	report := setupPublishedReport(t, db)
	ctx := context.Background()

	req, err := svc.Request(ctx, report.ID, 2, "信息有误")
	if err != nil {
		t.Fatal(err)
	}
	handled, err := svc.Approve(ctx, req.ID, 1, "同意")
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	if handled.NewReportID == 0 {
		t.Fatal("expected new resign report id")
	}

	original := reloadReport(t, db, report.ID)
	if original.Status != constants.ReportWithdrawn {
		t.Fatalf("original status = %s, want withdrawn (保留原版)", original.Status)
	}
	var newReport model.Report
	if err := db.First(&newReport, handled.NewReportID).Error; err != nil {
		t.Fatal(err)
	}
	if newReport.Status != constants.ReportResignPending {
		t.Fatalf("new report status = %s, want resign_pending", newReport.Status)
	}
	if newReport.Version != 2 || newReport.ParentReportID != report.ID {
		t.Fatalf("new version link mismatch: version=%d parent=%d", newReport.Version, newReport.ParentReportID)
	}
	if newReport.PDFURL != "" || newReport.GeneratedAt != nil {
		t.Fatal("resign version should not carry generated pdf artifacts")
	}

	// 版本链包含原版与新版本
	chain, err := svc.VersionChain(ctx, report.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(chain.Versions) != 2 {
		t.Fatalf("chain versions = %d, want 2", len(chain.Versions))
	}
	if chain.Versions[0].ID != report.ID || chain.Versions[1].ID != newReport.ID {
		t.Fatal("version chain order mismatch")
	}
}

func TestWithdraw_OnlyPublishedCanRequest(t *testing.T) {
	db := newTestDB(t)
	svc := newWithdrawService(db)
	ctx := context.Background()
	examinee := model.Examinee{Name: "王五", IDCardNo: "110101199303031234"}
	if err := db.Create(&examinee).Error; err != nil {
		t.Fatal(err)
	}
	draft := model.Report{ExamineeID: examinee.ID, ReportNo: "GB202609200002", Status: constants.ReportDraft, Version: 1}
	if err := db.Create(&draft).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Request(ctx, draft.ID, 2, ""); err == nil {
		t.Fatal("expected conflict when withdrawing non-published report")
	}
}

func TestWithdraw_ExpiredWindowConflict(t *testing.T) {
	db := newTestDB(t)
	svc := newWithdrawService(db)
	ctx := context.Background()
	examinee := model.Examinee{Name: "赵六", IDCardNo: "440101199404041234"}
	if err := db.Create(&examinee).Error; err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-25 * time.Hour)
	report := model.Report{
		ExamineeID: examinee.ID, ReportNo: "GB202609200003",
		Status: constants.ReportPublished, Version: 1, PublishedAt: &old,
	}
	if err := db.Create(&report).Error; err != nil {
		t.Fatal(err)
	}
	// 显式把 updated_at 也置旧，避免 UpdatedAt 回退分支掩盖窗口校验。
	if err := db.Model(&model.Report{}).Where("id = ?", report.ID).Update("updated_at", old).Error; err != nil {
		t.Fatal(err)
	}
	_, err := svc.Request(ctx, report.ID, 2, "")
	if err == nil {
		t.Fatal("expected withdraw window expired conflict")
	}
	var appErr *util.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != 409 || appErr.Code != constants.CodeWithdrawConflict {
		t.Fatalf("err = %v, want 409 code=%d", err, constants.CodeWithdrawConflict)
	}
	still := reloadReport(t, db, report.ID)
	if still.Status != constants.ReportPublished {
		t.Fatalf("status = %s, want unchanged published", still.Status)
	}
}

func TestWithdraw_DuplicatePendingConflict(t *testing.T) {
	db := newTestDB(t)
	svc := newWithdrawService(db)
	ctx := context.Background()
	report := setupPublishedReport(t, db)

	if _, err := svc.Request(ctx, report.ID, 2, "第一次"); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Request(ctx, report.ID, 2, "第二次")
	if err == nil {
		t.Fatal("expected duplicate pending conflict")
	}
	var appErr *util.AppError
	if !errors.As(err, &appErr) || appErr.HTTPStatus != 409 {
		t.Fatalf("err = %v, want 409", err)
	}
	still := reloadReport(t, db, report.ID)
	if still.Status != constants.ReportWithdrawing {
		t.Fatalf("status = %s, want unchanged withdrawing", still.Status)
	}
	var count int64
	db.Model(&model.ReportWithdrawRequest{}).Where("report_id = ? AND status = ?", report.ID, constants.WithdrawPending).Count(&count)
	if count != 1 {
		t.Fatalf("pending requests = %d, want 1", count)
	}
}

func TestWithdraw_ConcurrentApproveOnlyOneSucceeds(t *testing.T) {
	db := newTestDB(t)
	svc := newWithdrawService(db)
	ctx := context.Background()
	report := setupPublishedReport(t, db)
	req, err := svc.Request(ctx, report.ID, 2, "并发")
	if err != nil {
		t.Fatal(err)
	}

	// 两次针对同一申请的并发/重复审批：先批准成功，第二次必须冲突且状态不再变化。
	if _, err := svc.Approve(ctx, req.ID, 1, ""); err != nil {
		t.Fatalf("first Approve() error = %v", err)
	}
	if _, err := svc.Approve(ctx, req.ID, 1, ""); err == nil {
		t.Fatal("expected conflict on duplicate approve")
	}
	original := reloadReport(t, db, report.ID)
	if original.Status != constants.ReportWithdrawn {
		t.Fatalf("original status = %s, want withdrawn", original.Status)
	}
	var newCount int64
	db.Model(&model.Report{}).Where("parent_report_id = ?", report.ID).Count(&newCount)
	if newCount != 1 {
		t.Fatalf("resign versions = %d, want exactly 1", newCount)
	}
}

func TestWithdraw_ApproveAfterExpiryConflictAndStateUnchanged(t *testing.T) {
	db := newTestDB(t)
	svc := newWithdrawService(db)
	ctx := context.Background()
	report := setupPublishedReport(t, db)
	req, err := svc.Request(ctx, report.ID, 2, "超时")
	if err != nil {
		t.Fatal(err)
	}
	// 手动将申请置为已超时（未审批）。
	if err := db.Model(&model.ReportWithdrawRequest{}).Where("id = ?", req.ID).
		Update("expires_at", time.Now().Add(-time.Minute)).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Approve(ctx, req.ID, 1, ""); err == nil {
		t.Fatal("expected timeout conflict on approve")
	}
	if _, err := svc.Reject(ctx, req.ID, 1, ""); err == nil {
		t.Fatal("expected timeout conflict on reject")
	}
	original := reloadReport(t, db, report.ID)
	if original.Status != constants.ReportWithdrawing {
		t.Fatalf("status = %s, want unchanged withdrawing after timeout", original.Status)
	}
	var newCount int64
	db.Model(&model.Report{}).Where("parent_report_id = ?", report.ID).Count(&newCount)
	if newCount != 0 {
		t.Fatalf("resign versions = %d, want 0", newCount)
	}
}

func TestWithdraw_DownloadAndGenerateFrozenWhileWithdrawing(t *testing.T) {
	db := newTestDB(t)
	withdrawSvc := newWithdrawService(db)
	reportRepo := repository.NewReportRepository(db)
	resultRepo := repository.NewExamResultRepository(db)
	regRepo := repository.NewRegistrationRepository(db)
	ctx := context.Background()

	report := setupPublishedReport(t, db)
	if _, err := withdrawSvc.Request(ctx, report.ID, 2, "冻结"); err != nil {
		t.Fatal(err)
	}
	reportSvc := NewReportService(reportRepo, resultRepo, regRepo, t.TempDir(), testLogger())
	if _, err := reportSvc.Generate(ctx, report.ID, 2, "x", "", ""); err == nil {
		t.Fatal("expected generate frozen during withdrawing")
	}
	if _, err := reportSvc.PDFBytes(ctx, report.ID); err == nil {
		t.Fatal("expected pdf download frozen during withdrawing")
	}
}

func TestWithdraw_RejectThenCanRequestAgainWithinWindow(t *testing.T) {
	db := newTestDB(t)
	svc := newWithdrawService(db)
	ctx := context.Background()
	report := setupPublishedReport(t, db)

	first, err := svc.Request(ctx, report.ID, 2, "第一次")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Reject(ctx, first.ID, 1, "驳回"); err != nil {
		t.Fatal(err)
	}
	second, err := svc.Request(ctx, report.ID, 2, "再次申请")
	if err != nil {
		t.Fatalf("re-request after reject error = %v", err)
	}
	if second.Status != constants.WithdrawPending {
		t.Fatalf("second request status = %s, want pending", second.Status)
	}
}

func TestWithdraw_SecondChainApproveIncrementsVersion(t *testing.T) {
	db := newTestDB(t)
	svc := newWithdrawService(db)
	ctx := context.Background()
	report := setupPublishedReport(t, db)

	// 第一次撤回批准：v1 -> v2
	req1, err := svc.Request(ctx, report.ID, 2, "撤回1")
	if err != nil {
		t.Fatal(err)
	}
	res1, err := svc.Approve(ctx, req1.ID, 1, "")
	if err != nil {
		t.Fatal(err)
	}
	// v2 走完重新生成 -> 审核 -> 发布
	for _, step := range []struct {
		status string
		apply  func(uint) error
	}{
		{constants.ReportGenerated, func(id uint) error {
			_, e := newReportServiceForTest(db, t.TempDir()).Generate(ctx, id, 2, "新版结论", "", "")
			return e
		}},
		{constants.ReportReviewed, func(id uint) error { _, e := svcApproveReview(db, ctx, id); return e }},
	} {
		if err := step.apply(res1.NewReportID); err != nil {
			t.Fatalf("advance v2 to %s: %v", step.status, err)
		}
	}
	if _, err := svcApprovePublish(db, ctx, res1.NewReportID); err != nil {
		t.Fatal(err)
	}

	// 第二次撤回批准：v2 -> v3，版本链共 3 版
	req2, err := svc.Request(ctx, res1.NewReportID, 2, "撤回2")
	if err != nil {
		t.Fatalf("request on v2: %v", err)
	}
	res2, err := svc.Approve(ctx, req2.ID, 1, "")
	if err != nil {
		t.Fatalf("approve second: %v", err)
	}
	v3 := reloadReport(t, db, res2.NewReportID)
	if v3.Version != 3 || v3.ParentReportID != res1.NewReportID || v3.RootReportID != report.ID {
		t.Fatalf("v3 link mismatch: version=%d parent=%d root=%d", v3.Version, v3.ParentReportID, v3.RootReportID)
	}
	chain, err := svc.VersionChain(ctx, res2.NewReportID)
	if err != nil {
		t.Fatal(err)
	}
	if len(chain.Versions) != 3 {
		t.Fatalf("chain length = %d, want 3", len(chain.Versions))
	}
}

func newReportServiceForTest(db *gorm.DB, dir string) *ReportService {
	return NewReportService(repository.NewReportRepository(db),
		repository.NewExamResultRepository(db), repository.NewRegistrationRepository(db), dir, testLogger())
}

func svcApproveReview(db *gorm.DB, ctx context.Context, id uint) (*model.Report, error) {
	return newReportServiceForTest(db, "").Review(ctx, id)
}

func svcApprovePublish(db *gorm.DB, ctx context.Context, id uint) (*model.Report, error) {
	return newReportServiceForTest(db, "").Publish(ctx, id)
}
