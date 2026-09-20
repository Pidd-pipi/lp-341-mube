package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

// newWithdrawSvc 装配撤回重签闭环所需的报告/申请仓储与服务。
func newWithdrawSvc(t *testing.T, db *gorm.DB) (*ReportWithdrawService, *ReportService, *repository.ReportRepository) {
	reportRepo := repository.NewReportRepository(db)
	withdrawRepo := repository.NewReportWithdrawRepository(db)
	rsvc := NewReportService(reportRepo, repository.NewExamResultRepository(db),
		repository.NewRegistrationRepository(db), testLogger())
	rsvc.SetReportDir(t.TempDir())
	return NewReportWithdrawService(withdrawRepo, reportRepo, testLogger()), rsvc, reportRepo
}

// seedPublishedReport 落一份已发布报告（发布时间可回溯，用于 24h 窗口测试）。
func seedPublishedReport(t *testing.T, db *gorm.DB, publishedAt time.Time) *model.Report {
	t.Helper()
	n := time.Now().UnixNano()
	examinee := model.Examinee{Name: "王五", IDCardNo: fmt.Sprintf("310101%012d", n%1e12)}
	if err := db.Create(&examinee).Error; err != nil {
		t.Fatal(err)
	}
	reg := model.Registration{ExamineeID: examinee.ID, GuideNo: fmt.Sprintf("GUIDE-W-%d", n), Status: constants.RegistrationCompleted}
	if err := db.Create(&reg).Error; err != nil {
		t.Fatal(err)
	}
	report := &model.Report{
		RegistrationID: reg.ID, ExamineeID: examinee.ID, ReportNo: fmt.Sprintf("GB-W-%d", n),
		Status: constants.ReportPublished, Conclusion: "一切正常", PDFURL: "/reports/gb-w.pdf",
		VersionNo: 1, PublishedAt: &publishedAt,
	}
	if err := db.Create(report).Error; err != nil {
		t.Fatal(err)
	}
	return report
}

// TestReportWithdraw_ApplyApproveResignFlow 申请→冻结→批准生成待重签版本→原版归档保留→版本链。
func TestReportWithdraw_ApplyApproveResignFlow(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	wsvc, rsvc, reportRepo := newWithdrawSvc(t, db)
	publishedAt := time.Now().Add(-2 * time.Hour)
	report := seedPublishedReport(t, db, publishedAt)

	w, err := wsvc.Apply(ctx, report.ID, 2, "结论填写有误，需重签")
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if w.Status != constants.WithdrawPending || w.ExpiresAt.Sub(publishedAt) != constants.WithdrawWindowHours*time.Hour {
		t.Fatalf("pending withdraw invalid: %+v", w)
	}
	got, _ := reportRepo.FindByID(report.ID)
	if got.Status != constants.ReportWithdrawing {
		t.Fatalf("report status = %s, want withdrawing", got.Status)
	}

	// 审批期间冻结下载。
	if _, err := rsvc.PDFBytes(ctx, report.ID); !isAppStatus(err, constants.CodeReportWithdraw) {
		t.Fatalf("PDFBytes during withdrawing err = %v, want withdraw conflict", err)
	}
	// 审批期间冻结新版本生成（Generate）。
	if _, err := rsvc.Generate(ctx, report.ID, 2, "", "", ""); err == nil || !isAppStatus(err, constants.CodeReportWithdraw) {
		t.Fatalf("Generate during withdrawing err = %v, want withdraw conflict", err)
	}

	// 批准。
	res, err := wsvc.Approve(ctx, w.ID, 1, "同意撤回重签")
	if err != nil {
		t.Fatalf("Approve() error = %v", err)
	}
	original, _ := reportRepo.FindByID(report.ID)
	if original.Status != constants.ReportWithdrawn {
		t.Fatalf("original status = %s, want withdrawn(archived)", original.Status)
	}
	if res.Resign == nil || res.Resign.Status != constants.ReportDraft {
		t.Fatalf("resign report invalid: %+v", res.Resign)
	}
	if res.Resign.ID == report.ID || res.Resign.ParentReportID != report.ID ||
		res.Resign.RootReportID != report.ID || res.Resign.VersionNo != 2 {
		t.Fatalf("resign version chain invalid: %+v", res.Resign)
	}
	if res.Resign.PDFURL != "" || res.Resign.ReportNo == report.ReportNo {
		t.Fatalf("resign report should drop pdf and have new no: %+v", res.Resign)
	}
	// 原版内容保留。
	if original.Conclusion != "一切正常" || original.PDFURL != "/reports/gb-w.pdf" {
		t.Fatalf("original content not preserved: %+v", original)
	}

	// 待重签版本可继续走原生成→审核→发布流程。
	regenerated, err := rsvc.Generate(ctx, res.Resign.ID, 2, "", "", "")
	if err != nil {
		t.Fatalf("regenerate resign report: %v", err)
	}
	if regenerated.Status != constants.ReportGenerated {
		t.Fatalf("regenerated status = %s", regenerated.Status)
	}
	if _, err := rsvc.Review(ctx, res.Resign.ID); err != nil {
		t.Fatalf("review resign: %v", err)
	}
	republished, err := rsvc.Publish(ctx, res.Resign.ID)
	if err != nil {
		t.Fatalf("publish resign: %v", err)
	}
	if republished.Status != constants.ReportPublished || republished.PublishedAt == nil {
		t.Fatalf("republished invalid: %+v", republished)
	}

	// 版本链包含原版与重签版。
	chain, err := wsvc.Chain(ctx, report.ID)
	if err != nil {
		t.Fatalf("Chain() error = %v", err)
	}
	if len(chain.Versions) != 2 || chain.Versions[0].ID != report.ID ||
		chain.Versions[1].ID != res.Resign.ID {
		t.Fatalf("chain versions invalid: %+v", chain.Versions)
	}
	if len(chain.Withdraws) != 1 || chain.Withdraws[0].ID != w.ID {
		t.Fatalf("chain withdraws invalid: %+v", chain.Withdraws)
	}
}

// TestReportWithdraw_RejectRestoresPublished 驳回后恢复已发布，且无新版本产生。
func TestReportWithdraw_RejectRestoresPublished(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	wsvc, rsvc, reportRepo := newWithdrawSvc(t, db)
	report := seedPublishedReport(t, db, time.Now().Add(-time.Hour))

	w, err := wsvc.Apply(ctx, report.ID, 2, "需要修改")
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	updated, err := wsvc.Reject(ctx, w.ID, 1, "驳回，无需重签")
	if err != nil {
		t.Fatalf("Reject() error = %v", err)
	}
	if updated.Status != constants.WithdrawRejected || updated.NewReportID != 0 {
		t.Fatalf("rejected withdraw invalid: %+v", updated)
	}
	got, _ := reportRepo.FindByID(report.ID)
	if got.Status != constants.ReportPublished {
		t.Fatalf("report status = %s, want published restored", got.Status)
	}
	// 恢复后下载不再冻结。
	if _, err := rsvc.PDFBytes(ctx, report.ID); err != nil {
		t.Fatalf("pdf after reject: %v", err)
	}
	var count int64
	db.Model(&model.Report{}).Count(&count)
	if count != 1 {
		t.Fatalf("report count = %d, reject must not create new version", count)
	}
}

// TestReportWithdraw_Conflicts 超时申请、非发布态、重复申请、重复/并发审批均冲突且状态不变。
func TestReportWithdraw_Conflicts(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	wsvc, _, reportRepo := newWithdrawSvc(t, db)

	t.Run("window_expired_apply", func(t *testing.T) {
		report := seedPublishedReport(t, db, time.Now().Add(-25*time.Hour))
		_, err := wsvc.Apply(ctx, report.ID, 2, "超期撤回")
		if !isAppStatus(err, constants.CodeReportWithdraw) {
			t.Fatalf("err = %v, want withdraw conflict", err)
		}
		got, _ := reportRepo.FindByID(report.ID)
		if got.Status != constants.ReportPublished {
			t.Fatalf("status changed after expired apply: %s", got.Status)
		}
	})

	t.Run("not_published_apply", func(t *testing.T) {
		examinee := model.Examinee{Name: "赵六", IDCardNo: "110101199301015678"}
		_ = db.Create(&examinee).Error
		reg := model.Registration{ExamineeID: examinee.ID, GuideNo: "GUIDE-W-0002", Status: constants.RegistrationRegistered}
		_ = db.Create(&reg).Error
		report := model.Report{RegistrationID: reg.ID, ExamineeID: examinee.ID,
			ReportNo: "GB-W-0002", Status: constants.ReportReviewed, VersionNo: 1}
		if err := db.Create(&report).Error; err != nil {
			t.Fatal(err)
		}
		if _, err := wsvc.Apply(ctx, report.ID, 2, "未发布撤回"); !isAppStatus(err, constants.CodeReportWithdraw) {
			t.Fatalf("err = %v, want conflict", err)
		}
	})

	t.Run("duplicate_apply", func(t *testing.T) {
		report := seedPublishedReport(t, db, time.Now().Add(-30*time.Minute))
		if _, err := wsvc.Apply(ctx, report.ID, 2, "第一次申请"); err != nil {
			t.Fatal(err)
		}
		_, err := wsvc.Apply(ctx, report.ID, 2, "重复申请")
		if !isAppStatus(err, constants.CodeReportWithdraw) {
			t.Fatalf("duplicate apply err = %v, want conflict", err)
		}
		var pending int64
		db.Model(&model.ReportWithdraw{}).Where("report_id = ? AND status = ?", report.ID, constants.WithdrawPending).Count(&pending)
		if pending != 1 {
			t.Fatalf("pending count = %d, want 1", pending)
		}
	})

	t.Run("duplicate_review", func(t *testing.T) {
		report := seedPublishedReport(t, db, time.Now().Add(-10*time.Minute))
		w, err := wsvc.Apply(ctx, report.ID, 2, "申请")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := wsvc.Approve(ctx, w.ID, 1, "批准"); err != nil {
			t.Fatal(err)
		}
		// 再次批准/驳回：冲突且状态不变。
		if _, err := wsvc.Approve(ctx, w.ID, 1, "重复批准"); !isAppStatus(err, constants.CodeReportWithdraw) {
			t.Fatalf("double approve err = %v", err)
		}
		if _, err := wsvc.Reject(ctx, w.ID, 1, "重复驳回"); !isAppStatus(err, constants.CodeReportWithdraw) {
			t.Fatalf("approve then reject err = %v", err)
		}
		got, _ := wsvc.Get(ctx, w.ID)
		if got.Status != constants.WithdrawApproved {
			t.Fatalf("withdraw status changed after duplicate review: %s", got.Status)
		}
	})

	t.Run("review_expired_window", func(t *testing.T) {
		// 申请尚在 pending，但已超 24h：审批返回冲突，申请与报告状态不变。
		report := seedPublishedReport(t, db, time.Now().Add(-23*time.Hour))
		w, err := wsvc.Apply(ctx, report.ID, 2, "临界申请")
		if err != nil {
			t.Fatal(err)
		}
		expired := time.Now().Add(time.Hour)
		if err := db.Model(&model.ReportWithdraw{}).Where("id = ?", w.ID).
			Update("expires_at", expired.Add(-25*time.Hour)).Error; err != nil {
			t.Fatal(err)
		}
		if _, err := wsvc.Approve(ctx, w.ID, 1, "超时批准"); !isAppStatus(err, constants.CodeReportWithdraw) {
			t.Fatalf("expired review err = %v", err)
		}
		got, _ := wsvc.Get(ctx, w.ID)
		if got.Status != constants.WithdrawPending {
			t.Fatalf("expired review changed withdraw status: %s", got.Status)
		}
		rg, _ := reportRepo.FindByID(report.ID)
		if rg.Status != constants.ReportWithdrawing {
			t.Fatalf("expired review changed report status: %s", rg.Status)
		}
	})
}

// TestReportWithdraw_ApprovedVersionCanWithdrawAgain 重签发布后可再次撤回，版本链继续延伸。
func TestReportWithdraw_ApprovedVersionCanWithdrawAgain(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	wsvc, rsvc, _ := newWithdrawSvc(t, db)
	report := seedPublishedReport(t, db, time.Now().Add(-time.Hour))

	w1, _ := wsvc.Apply(ctx, report.ID, 2, "第一次撤回")
	res1, err := wsvc.Approve(ctx, w1.ID, 1, "批准")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := rsvc.Generate(ctx, res1.Resign.ID, 2, "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := rsvc.Review(ctx, res1.Resign.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := rsvc.Publish(ctx, res1.Resign.ID); err != nil {
		t.Fatal(err)
	}

	w2, err := wsvc.Apply(ctx, res1.Resign.ID, 2, "第二次撤回")
	if err != nil {
		t.Fatalf("second withdraw apply: %v", err)
	}
	res2, err := wsvc.Approve(ctx, w2.ID, 1, "再次批准")
	if err != nil {
		t.Fatalf("second approve: %v", err)
	}
	if res2.Resign.VersionNo != 3 || res2.Resign.RootReportID != report.ID ||
		res2.Resign.ParentReportID != res1.Resign.ID {
		t.Fatalf("v3 chain invalid: %+v", res2.Resign)
	}
	chain, _ := wsvc.Chain(ctx, res2.Resign.ID)
	if len(chain.Versions) != 3 {
		t.Fatalf("chain len = %d, want 3", len(chain.Versions))
	}
}

func isAppStatus(err error, code int) bool {
	var appErr *util.AppError
	return errors.As(err, &appErr) && appErr.Code == code
}
