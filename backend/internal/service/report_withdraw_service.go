package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

// ReportWithdrawService 报告撤回重签闭环服务：
// 已发布 24 小时内可申请撤回 → 审批期间冻结下载与新版本生成
// → 批准：原版归档(保留)并生成待重签草稿版本；驳回：恢复已发布。
type ReportWithdrawService struct {
	withdrawRepo *repository.ReportWithdrawRepository
	reportRepo   *repository.ReportRepository
	log          *slog.Logger
}

// NewReportWithdrawService 构造撤回重签服务。
func NewReportWithdrawService(withdrawRepo *repository.ReportWithdrawRepository, reportRepo *repository.ReportRepository, log *slog.Logger) *ReportWithdrawService {
	return &ReportWithdrawService{withdrawRepo: withdrawRepo, reportRepo: reportRepo, log: log}
}

// Apply 申请撤回。同一报告只能有一份待审批申请；超出发布 24 小时窗口冲突。
func (s *ReportWithdrawService) Apply(ctx context.Context, reportID, applicantID uint, reason string) (*model.ReportWithdraw, error) {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, util.NotFoundError(constants.MsgReportNotFound, err)
	}
	if report.Status != constants.ReportPublished {
		return nil, util.NewAppError(constants.CodeReportWithdraw, 409,
			fmt.Sprintf("Report[no=%s] withdraw apply failed for applicant[role submit]: %s", report.ReportNo, constants.MsgWithdrawNotPublished),
			errors.New("status not published"))
	}
	if s.isWindowExpired(report) {
		s.log.WarnContext(ctx, constants.LOG_REPORT_WITHDRAW_EXPIRED, "report_id", reportID, "report_no", report.ReportNo)
		return nil, util.NewAppError(constants.CodeReportWithdraw, 409,
			fmt.Sprintf("Report[no=%s] withdraw apply failed: %s", report.ReportNo, constants.MsgWithdrawWindowExpired),
			errors.New("withdraw window expired"))
	}
	exists, err := s.withdrawRepo.ExistsPendingByReport(reportID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, util.NewAppError(constants.CodeReportWithdraw, 409,
			fmt.Sprintf("Report[no=%s] withdraw apply failed: %s", report.ReportNo, constants.MsgWithdrawDuplicate),
			errors.New("pending withdraw exists"))
	}

	w := &model.ReportWithdraw{
		ReportID:    reportID,
		ApplicantID: applicantID,
		Reason:      reason,
		Status:      constants.WithdrawPending,
		ExpiresAt:   s.windowDeadline(report),
	}
	// 事务内先 CAS 报告 published→withdrawing，再落申请；
	// 并发重复申请时只有一方 CAS 成功，另一方得到冲突且报告状态不变。
	txErr := s.reportRepo.DB().Transaction(func(tx *gorm.DB) error {
		rows, err := repository.NewReportRepository(tx).CASStatus(reportID,
			[]string{constants.ReportPublished}, constants.ReportWithdrawing)
		if err != nil {
			return err
		}
		if rows == 0 {
			return util.NewAppError(constants.CodeReportWithdraw, 409,
				fmt.Sprintf("Report[no=%s] withdraw apply conflict: %s", report.ReportNo, constants.MsgWithdrawDuplicate),
				errors.New("cas published->withdrawing affected 0 rows"))
		}
		if err := repository.NewReportWithdrawRepository(tx).Create(w); err != nil {
			return fmt.Errorf("create withdraw: %w", err)
		}
		return nil
	})
	if txErr != nil {
		var appErr *util.AppError
		if errors.As(txErr, &appErr) {
			s.log.WarnContext(ctx, constants.LOG_REPORT_WITHDRAW_DUPLICATE, "report_id", reportID)
			return nil, appErr
		}
		return nil, util.LogError(s.log, constants.LOG_REPORT_WITHDRAW_DUPLICATE, txErr)
	}
	s.log.InfoContext(ctx, constants.LOG_REPORT_WITHDRAW_APPLIED,
		"withdraw_id", w.ID, "report_id", reportID, "applicant_id", applicantID)
	return w, nil
}

// Approve 管理员批准：生成待重签版本并保留原版（原版置 withdrawn 归档）。
func (s *ReportWithdrawService) Approve(ctx context.Context, withdrawID, reviewerID uint, comment string) (*ApproveResult, error) {
	return s.review(ctx, withdrawID, reviewerID, comment, true)
}

// Reject 管理员驳回：申请关闭，报告恢复 published。
func (s *ReportWithdrawService) Reject(ctx context.Context, withdrawID, reviewerID uint, comment string) (*model.ReportWithdraw, error) {
	res, err := s.review(ctx, withdrawID, reviewerID, comment, false)
	if err != nil {
		return nil, err
	}
	return res.Withdraw, nil
}

// ApproveResult 批准结果：申请、归档原版、待重签新版本。
type ApproveResult struct {
	Withdraw *model.ReportWithdraw `json:"withdraw"`
	Original *model.Report         `json:"original"`
	Resign   *model.Report         `json:"resign_report"`
}

func (s *ReportWithdrawService) review(ctx context.Context, withdrawID, reviewerID uint, comment string, approve bool) (*ApproveResult, error) {
	w, err := s.withdrawRepo.FindByID(withdrawID)
	if err != nil {
		return nil, util.NotFoundError(constants.MsgWithdrawNotFound, err)
	}
	if w.Status != constants.WithdrawPending {
		// 重复或并发审批：冲突且状态不变。
		return nil, s.concurrentConflict(ctx, w.ID, w.Status)
	}
	if time.Now().After(w.ExpiresAt) {
		s.log.WarnContext(ctx, constants.LOG_REPORT_WITHDRAW_EXPIRED, "withdraw_id", withdrawID)
		return nil, util.NewAppError(constants.CodeReportWithdraw, 409,
			fmt.Sprintf("ReportWithdraw[id=%d] review failed: %s", withdrawID, constants.MsgWithdrawWindowExpired),
			errors.New("withdraw window expired"))
	}
	report, err := s.reportRepo.FindByID(w.ReportID)
	if err != nil {
		return nil, util.NotFoundError(constants.MsgReportNotFound, err)
	}
	if report.Status != constants.ReportWithdrawing {
		return nil, s.concurrentConflict(ctx, w.ID, report.Status)
	}

	now := time.Now()
	targetStatus := constants.WithdrawRejected
	reportTarget := constants.ReportPublished // 驳回恢复已发布
	if approve {
		targetStatus = constants.WithdrawApproved
		reportTarget = constants.ReportWithdrawn // 批准原版归档保留
	}

	result := &ApproveResult{Withdraw: w, Original: report}
	txErr := s.reportRepo.DB().Transaction(func(tx *gorm.DB) error {
		txReportRepo := repository.NewReportRepository(tx)
		txWithdrawRepo := repository.NewReportWithdrawRepository(tx)

		// 1. 先占用申请行（pending→终态）：并发审批中只有一方 CAS 成功，
		//    其余立即冲突回滚，绝不能先建待重签版本再判定（避免唯一键冲突）。
		rows, err := txWithdrawRepo.CASReview(withdrawID, reviewerID, 0, targetStatus, comment, now)
		if err != nil {
			return err
		}
		if rows == 0 {
			return util.NewAppError(constants.CodeReportWithdraw, 409,
				fmt.Sprintf("ReportWithdraw[id=%d] concurrent review: %s", withdrawID, constants.MsgWithdrawNotPending),
				errors.New("cas withdraw review affected 0 rows"))
		}
		// 2. 再占用报告状态（withdrawing→终态）。
		reportRows, err := txReportRepo.CASStatus(report.ID,
			[]string{constants.ReportWithdrawing}, reportTarget)
		if err != nil {
			return err
		}
		if reportRows == 0 {
			return util.NewAppError(constants.CodeReportWithdraw, 409,
				fmt.Sprintf("Report[no=%s] concurrent review: %s", report.ReportNo, constants.MsgWithdrawFrozen),
				errors.New("cas report status affected 0 rows"))
		}
		// 3. 两道 CAS 均通过后，批准场景才创建待重签版本并回写申请。
		if approve {
			resign := s.buildResignReport(report)
			if err := txReportRepo.Create(resign); err != nil {
				return fmt.Errorf("create resign report: %w", err)
			}
			if err := txWithdrawRepo.UpdateNewReportID(withdrawID, resign.ID); err != nil {
				return fmt.Errorf("link resign report: %w", err)
			}
			result.Resign = resign
		}
		return nil
	})
	if txErr != nil {
		var appErr *util.AppError
		if errors.As(txErr, &appErr) {
			return nil, s.concurrentConflict(ctx, w.ID, targetStatus)
		}
		return nil, util.LogError(s.log, constants.LOG_REPORT_WITHDRAW_CONFLICT, txErr)
	}

	// 回填审批结果字段供响应使用。
	w.Status = targetStatus
	w.ReviewerID = reviewerID
	w.ReviewComment = comment
	w.ReviewedAt = &now
	if approve {
		w.NewReportID = result.Resign.ID
		s.log.InfoContext(ctx, constants.LOG_REPORT_WITHDRAW_APPROVED,
			"withdraw_id", withdrawID, "report_id", report.ID,
			"original_report_id", report.ID, "resign_report_id", result.Resign.ID, "reviewer_id", reviewerID)
		report.Status = constants.ReportWithdrawn
	} else {
		s.log.InfoContext(ctx, constants.LOG_REPORT_WITHDRAW_REJECTED,
			"withdraw_id", withdrawID, "report_id", report.ID, "reviewer_id", reviewerID)
		report.Status = constants.ReportPublished
	}
	result.Withdraw = w
	result.Original = report
	return result, nil
}

// Get 查询撤回申请详情。
func (s *ReportWithdrawService) Get(ctx context.Context, id uint) (*model.ReportWithdraw, error) {
	w, err := s.withdrawRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError(constants.MsgWithdrawNotFound, err)
		}
		return nil, err
	}
	return w, nil
}

// List 分页查询撤回申请（可按状态/报告过滤）。
func (s *ReportWithdrawService) List(ctx context.Context, status string, reportID uint, page, pageSize int) ([]model.ReportWithdraw, int64, error) {
	return s.withdrawRepo.ListByReports(status, reportID, page, pageSize)
}

// Chain 查看报告版本链（含各版本对应的撤回申请）。
func (s *ReportWithdrawService) Chain(ctx context.Context, reportID uint) (*VersionChain, error) {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		return nil, util.NotFoundError(constants.MsgReportNotFound, err)
	}
	rootID := report.RootReportID
	if rootID == 0 {
		rootID = report.ID
	}
	versions, err := s.reportRepo.ListChain(rootID)
	if err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(versions))
	for i := range versions {
		ids = append(ids, versions[i].ID)
	}
	var withdraws []model.ReportWithdraw
	if len(ids) > 0 {
		if err := s.reportRepo.DB().
			Where("report_id IN ? OR new_report_id IN ?", ids, ids).
			Order("id asc").Find(&withdraws).Error; err != nil {
			return nil, err
		}
	}
	return &VersionChain{RootReportID: rootID, Versions: versions, Withdraws: withdraws}, nil
}

// VersionChain 版本链响应。
type VersionChain struct {
	RootReportID uint                   `json:"root_report_id"`
	Versions     []model.Report         `json:"versions"`
	Withdraws    []model.ReportWithdraw `json:"withdraws"`
}

// buildResignReport 基于原版生成待重签草稿版本：保留结论等内容与版本链关系，
// 不复制 PDF（需重新走生成流程），状态为 draft。
func (s *ReportWithdrawService) buildResignReport(old *model.Report) *model.Report {
	rootID := old.RootReportID
	if rootID == 0 {
		rootID = old.ID
	}
	newVersion := old.VersionNo + 1
	return &model.Report{
		RegistrationID:   old.RegistrationID,
		ExamineeID:       old.ExamineeID,
		ReportNo:         fmt.Sprintf("%s-V%d", old.ReportNo, newVersion),
		Status:           constants.ReportDraft,
		Conclusion:       old.Conclusion,
		HealthAdvice:     old.HealthAdvice,
		FollowUpReminder: old.FollowUpReminder,
		PDFURL:           "",
		DoctorID:         old.DoctorID,
		VersionNo:        newVersion,
		RootReportID:     rootID,
		ParentReportID:   old.ID,
	}
}

func (s *ReportWithdrawService) windowDeadline(r *model.Report) time.Time {
	base := r.PublishedAt
	if base == nil {
		t := r.UpdatedAt
		base = &t
	}
	return base.Add(time.Duration(constants.WithdrawWindowHours) * time.Hour)
}

func (s *ReportWithdrawService) isWindowExpired(r *model.Report) bool {
	return time.Now().After(s.windowDeadline(r))
}

func (s *ReportWithdrawService) concurrentConflict(ctx context.Context, withdrawID uint, current string) error {
	s.log.WarnContext(ctx, constants.LOG_REPORT_WITHDRAW_CONFLICT,
		"withdraw_id", withdrawID, "current_status", current)
	return util.NewAppError(constants.CodeReportWithdraw, 409,
		fmt.Sprintf("ReportWithdraw[id=%d] review failed[status=%s]: %s", withdrawID, current, constants.MsgWithdrawNotPending),
		errors.New("not pending"))
}
