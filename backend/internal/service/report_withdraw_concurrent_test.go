package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/model"
)

// TestReportWithdraw_ConcurrentApproval 并发审批同一申请：恰好一方成功，另一方冲突且状态不变。
func TestReportWithdraw_ConcurrentApproval(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	wsvc, _, reportRepo := newWithdrawSvc(t, db)
	report := seedPublishedReport(t, db, time.Now().Add(-time.Hour))
	w, err := wsvc.Apply(ctx, report.ID, 2, "并发审批测试")
	if err != nil {
		t.Fatal(err)
	}

	const n = 8
	var wg sync.WaitGroup
	var successMu sync.Mutex
	successes, conflicts := 0, 0
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			_, err := wsvc.Approve(ctx, w.ID, 1, "并发审批")
			_ = i
			successMu.Lock()
			defer successMu.Unlock()
			if err == nil {
				successes++
			} else if isAppStatus(err, constants.CodeReportWithdraw) {
				conflicts++
			} else {
				t.Errorf("unexpected err: %v", err)
			}
		}()
	}
	wg.Wait()

	if successes != 1 {
		t.Fatalf("successes = %d, want exactly 1", successes)
	}
	if conflicts != n-1 {
		t.Fatalf("conflicts = %d, want %d", conflicts, n-1)
	}
	got, _ := wsvc.Get(ctx, w.ID)
	if got.Status != constants.WithdrawApproved {
		t.Fatalf("withdraw status = %s, want approved", got.Status)
	}
	rg, _ := reportRepo.FindByID(report.ID)
	if rg.Status != constants.ReportWithdrawn {
		t.Fatalf("report status = %s, want withdrawn", rg.Status)
	}
	var versionCount int64
	db.Model(&model.Report{}).Count(&versionCount)
	if versionCount != 2 {
		t.Fatalf("report version count = %d, want 2 (original + one resign)", versionCount)
	}
}

// TestReportWithdraw_ConcurrentApply 并发申请同一报告：恰好一份 pending。
func TestReportWithdraw_ConcurrentApply(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	wsvc, _, _ := newWithdrawSvc(t, db)
	report := seedPublishedReport(t, db, time.Now().Add(-30*time.Minute))

	const n = 8
	var wg sync.WaitGroup
	var mu sync.Mutex
	successes, conflicts := 0, 0
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			_, err := wsvc.Apply(ctx, report.ID, 2, "并发申请撤回")
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				successes++
			} else if isAppStatus(err, constants.CodeReportWithdraw) {
				conflicts++
			} else {
				t.Errorf("unexpected err: %v", err)
			}
		}()
	}
	wg.Wait()

	if successes != 1 {
		t.Fatalf("apply successes = %d, want 1", successes)
	}
	if conflicts != n-1 {
		t.Fatalf("apply conflicts = %d, want %d", conflicts, n-1)
	}
	var pending int64
	db.Model(&model.ReportWithdraw{}).Where("report_id = ? AND status = ?", report.ID, constants.WithdrawPending).Count(&pending)
	if pending != 1 {
		t.Fatalf("pending withdraws = %d, want 1", pending)
	}
}
