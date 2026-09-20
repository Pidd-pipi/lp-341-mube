package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

// ReportWithdrawService 报告撤回重签闭环：申请、审批、版本链。
type ReportWithdrawService struct {
	withdrawRepo *repository.ReportWithdrawRepository
	reportRepo   *repository.ReportRepository
	log          *slog.Logger
}

// NewReportWithdrawService 构造撤回重签服务。
func NewReportWithdrawService(withdrawRepo *repository.ReportWithdrawRepository, reportRepo *repository.ReportRepository, log *slog.Logger) *ReportWithdrawService {
	return &ReportWithdrawService{withdrawRepo: withdrawRepo, reportRepo: reportRepo, log: log}
}

// Request 已发布报告 24 小时内申请撤回；同一报告只能有一份待审批申请。
func (s *ReportWithdrawService) Request(ctx context.Context, reportID, applicantID uint, reason string) (*model.ReportWithdrawRequest, error) {
	var created *model.ReportWithdrawRequest
	err := s.reportRepo.DB().Transaction(func(tx *gorm.DB) error {
		reportTx := s.reportRepo.WithDB(tx)
		withdrawTx := s.withdrawRepo.WithDB(tx)

		// 锁定报告行，保证两个并发申请串行执行；后者必能看到前者已提交的 pending 申请。
		report, err := reportTx.FindByIDForUpdate(reportID)
		if err != nil {
			return util.NotFoundError(constants.MsgReportNotFound, err)
		}
		if report.Status != constants.ReportPublished {
			return util.NewAppError(constants.CodeWithdrawConflict, http.StatusConflict, constants.MsgWithdrawNotPublished,
				fmt.Errorf("report[no=%s] withdraw failed: status=%s", report.ReportNo, report.Status))
		}
		// 撤回窗口：自发布起 24 小时；历史数据无发布时间时回退到更新时间。
		since := report.UpdatedAt
		if report.PublishedAt != nil {
			since = *report.PublishedAt
		}
		if time.Since(since) > constants.ReportWithdrawWindow {
			return util.NewAppError(constants.CodeWithdrawConflict, http.StatusConflict, constants.MsgWithdrawWindowExpired,
				fmt.Errorf("report[no=%s] withdraw failed: published at %s", report.ReportNo, util.FormatTime(since)))
		}
		pending, err := withdrawTx.CountPendingByReport(reportID)
		if err != nil {
			return err
		}
		if pending > 0 {
			return util.NewAppError(constants.CodeWithdrawConflict, http.StatusConflict, constants.MsgWithdrawDuplicate,
				fmt.Errorf("report[no=%s] withdraw conflict: pending=%d", report.ReportNo, pending))
		}
		now := time.Now()
		req := &model.ReportWithdrawRequest{
			ReportID:    reportID,
			ApplicantID: applicantID,
			Reason:      reason,
			Status:      constants.WithdrawPending,
			ExpiresAt:   now.Add(constants.ReportWithdrawWindow),
		}
		// published -> withdrawing，进入冻结态；条件更新失败说明并发改动，整事务回滚。
		if affected, err := reportTx.UpdateStatusIf(reportID,
			[]string{constants.ReportPublished}, constants.ReportWithdrawing); err != nil {
			return err
		} else if affected == 0 {
			return util.NewAppError(constants.CodeWithdrawConflict, http.StatusConflict, constants.MsgWithdrawDuplicate,
				errors.New("report status changed concurrently"))
		}
		if err := withdrawTx.Create(req); err != nil {
			return err
		}
		created = req
		return nil
	})
	if err != nil {
		return nil, util.LogError(s.log, constants.LOG_REPORT_WITHDRAW_REQUEST_FAILED, err, "report_id", reportID, "applicant_id", applicantID)
	}
	s.log.InfoContext(ctx, constants.LOG_REPORT_WITHDRAW_REQUESTED,
		"withdraw_id", created.ID, "report_id", reportID, "applicant_id", applicantID)
	return created, nil
}

// Approve 管理员批准：申请关闭、原版归档保留、生成待重签新版本。
func (s *ReportWithdrawService) Approve(ctx context.Context, requestID, reviewerID uint, comment string) (*model.ReportWithdrawRequest, error) {
	var result *model.ReportWithdrawRequest
	err := s.reportRepo.DB().Transaction(func(tx *gorm.DB) error {
		reportTx := s.reportRepo.WithDB(tx)
		withdrawTx := s.withdrawRepo.WithDB(tx)

		req, err := withdrawTx.FindByIDForUpdate(requestID)
		if err != nil {
			return util.NotFoundError(constants.MsgWithdrawNotFound, err)
		}
		if req.Status != constants.WithdrawPending {
			return util.NewAppError(constants.CodeWithdrawConflict, http.StatusConflict, constants.MsgWithdrawAlreadyHandled,
				fmt.Errorf("withdraw[id=%d] approve conflict: status=%s", requestID, req.Status))
		}
		if time.Now().After(req.ExpiresAt) {
			return util.NewAppError(constants.CodeWithdrawConflict, http.StatusConflict, constants.MsgWithdrawTimeout,
				fmt.Errorf("withdraw[id=%d] approve conflict: expired at %s", requestID, util.FormatTime(req.ExpiresAt)))
		}
		report, err := reportTx.FindByIDForUpdate(req.ReportID)
		if err != nil {
			return util.NotFoundError(constants.MsgReportNotFound, err)
		}
		if report.Status != constants.ReportWithdrawing {
			// 报告状态已被并发改动：不动状态、不动申请，返回冲突。
			return util.NewAppError(constants.CodeWithdrawConflict, http.StatusConflict, constants.MsgWithdrawAlreadyHandled,
				fmt.Errorf("report[id=%d] approve conflict: status=%s", report.ID, report.Status))
		}

		// 构造待重签新版本：内容沿用原版，清空生成/发布产物，等待重新生成 PDF 与发布。
		rootID := report.RootReportID
		if rootID == 0 {
			rootID = report.ID
		}
		chainCount, err := reportTx.CountByRoot(rootID)
		if err != nil {
			return err
		}
		newVersion := int(chainCount) + 1
		newReport := &model.Report{
			RegistrationID:   report.RegistrationID,
			ExamineeID:       report.ExamineeID,
			ReportNo:         fmt.Sprintf("%s-R%d", reportOriginReportNo(report.ReportNo), newVersion),
			Status:           constants.ReportResignPending,
			Conclusion:       report.Conclusion,
			HealthAdvice:     report.HealthAdvice,
			FollowUpReminder: report.FollowUpReminder,
			DoctorID:         reviewerID,
			RootReportID:     rootID,
			ParentReportID:   report.ID,
			Version:          newVersion,
		}
		if err := reportTx.Create(newReport); err != nil {
			return fmt.Errorf("create resign report: %w", err)
		}

		affected, err := withdrawTx.ApproveIfPending(requestID, reviewerID, newReport.ID, comment, time.Now())
		if err != nil {
			return err
		}
		if affected == 0 {
			// 并发/超时：条件更新未命中，回滚（新版本创建一并撤销，状态不变）。
			return util.NewAppError(constants.CodeWithdrawConflict, http.StatusConflict, constants.MsgWithdrawAlreadyHandled,
				errors.New("withdraw request changed concurrently or expired"))
		}
		// 原版归档保留（withdrawing -> withdrawn），冻结解除转由新版本承载。
		if affected, err := reportTx.UpdateStatusIf(report.ID,
			[]string{constants.ReportWithdrawing}, constants.ReportWithdrawn); err != nil {
			return err
		} else if affected == 0 {
			return util.NewAppError(constants.CodeWithdrawConflict, http.StatusConflict, constants.MsgWithdrawAlreadyHandled,
				errors.New("report status changed concurrently"))
		}

		result = req
		result.Status = constants.WithdrawApproved
		result.ReviewerID = reviewerID
		result.ReviewComment = comment
		result.NewReportID = newReport.ID
		return nil
	})
	if err != nil {
		return nil, util.LogError(s.log, constants.LOG_REPORT_WITHDRAW_APPROVE_FAILED, err,
			"withdraw_id", requestID, "reviewer_id", reviewerID)
	}
	s.log.InfoContext(ctx, constants.LOG_REPORT_WITHDRAW_APPROVED,
		"withdraw_id", requestID, "reviewer_id", reviewerID, "report_id", result.ReportID, "new_report_id", result.NewReportID)
	return result, nil
}

// Reject 管理员驳回：申请关闭，报告恢复已发布。
func (s *ReportWithdrawService) Reject(ctx context.Context, requestID, reviewerID uint, comment string) (*model.ReportWithdrawRequest, error) {
	var result *model.ReportWithdrawRequest
	err := s.reportRepo.DB().Transaction(func(tx *gorm.DB) error {
		reportTx := s.reportRepo.WithDB(tx)
		withdrawTx := s.withdrawRepo.WithDB(tx)

		req, err := withdrawTx.FindByIDForUpdate(requestID)
		if err != nil {
			return util.NotFoundError(constants.MsgWithdrawNotFound, err)
		}
		if req.Status != constants.WithdrawPending {
			return util.NewAppError(constants.CodeWithdrawConflict, http.StatusConflict, constants.MsgWithdrawAlreadyHandled,
				fmt.Errorf("withdraw[id=%d] reject conflict: status=%s", requestID, req.Status))
		}
		if time.Now().After(req.ExpiresAt) {
			return util.NewAppError(constants.CodeWithdrawConflict, http.StatusConflict, constants.MsgWithdrawTimeout,
				fmt.Errorf("withdraw[id=%d] reject conflict: expired at %s", requestID, util.FormatTime(req.ExpiresAt)))
		}

		// 锁定报告行，与撤回申请、批准流程串行，避免并发改状态。
		if _, err := reportTx.FindByIDForUpdate(req.ReportID); err != nil {
			return util.NotFoundError(constants.MsgReportNotFound, err)
		}

		affected, err := withdrawTx.RejectIfPending(requestID, reviewerID, comment, time.Now())
		if err != nil {
			return err
		}
		if affected == 0 {
			return util.NewAppError(constants.CodeWithdrawConflict, http.StatusConflict, constants.MsgWithdrawAlreadyHandled,
				errors.New("withdraw request changed concurrently or expired"))
		}
		// withdrawing -> published，恢复下载。
		if affected, err := reportTx.RestoreStatusIf(req.ReportID,
			constants.ReportWithdrawing, constants.ReportPublished); err != nil {
			return err
		} else if affected == 0 {
			return util.NewAppError(constants.CodeWithdrawConflict, http.StatusConflict, constants.MsgWithdrawAlreadyHandled,
				errors.New("report status changed concurrently"))
		}

		result = req
		result.Status = constants.WithdrawRejected
		result.ReviewerID = reviewerID
		result.ReviewComment = comment
		return nil
	})
	if err != nil {
		return nil, util.LogError(s.log, constants.LOG_REPORT_WITHDRAW_APPROVE_FAILED, err,
			"withdraw_id", requestID, "reviewer_id", reviewerID)
	}
	s.log.InfoContext(ctx, constants.LOG_REPORT_WITHDRAW_REJECTED,
		"withdraw_id", requestID, "reviewer_id", reviewerID, "report_id", result.ReportID)
	return result, nil
}

// VersionChain 版本链视图：报告版本列表 + 撤回申请记录。
func (s *ReportWithdrawService) VersionChain(ctx context.Context, reportID uint) (*VersionChain, error) {
	report, err := s.reportRepo.FindByID(reportID)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError(constants.MsgReportNotFound, err)
		}
		return nil, err
	}
	rootID := report.RootReportID
	if rootID == 0 {
		rootID = report.ID
	}
	versions, err := s.reportRepo.ListVersionChain(rootID)
	if err != nil {
		return nil, err
	}
	requests := make([]model.ReportWithdrawRequest, 0)
	seen := map[uint]bool{}
	for _, v := range versions {
		if seen[v.ID] {
			continue
		}
		seen[v.ID] = true
		list, err := s.withdrawRepo.ListByReport(v.ID)
		if err != nil {
			return nil, err
		}
		requests = append(requests, list...)
	}
	pending, err := s.withdrawRepo.FindPendingByReport(reportID)
	if err != nil && !errors.Is(err, util.ErrNotFound) {
		return nil, err
	}
	s.log.InfoContext(ctx, constants.LOG_REPORT_VERSION_CHAIN_QUERIED, "report_id", reportID, "root_report_id", rootID)
	chain := &VersionChain{RootReportID: rootID, Versions: versions, Requests: requests}
	if pending != nil {
		chain.PendingRequest = pending
	}
	return chain, nil
}

// reportOriginReportNo 去除历史重签后缀：GB202608160001-R2 -> GB202608160001。
func reportOriginReportNo(reportNo string) string {
	if idx := strings.LastIndex(reportNo, "-R"); idx > 0 {
		return reportNo[:idx]
	}
	return reportNo
}
