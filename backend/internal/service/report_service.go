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
)

// ReportService 报告服务：生成、审核、发布、PDF。
type ReportService struct {
	repo       *repository.ReportRepository
	resultRepo *repository.ExamResultRepository
	regRepo    *repository.RegistrationRepository
	log        *slog.Logger
}

// NewReportService 构造报告服务。
func NewReportService(repo *repository.ReportRepository, resultRepo *repository.ExamResultRepository, regRepo *repository.RegistrationRepository, log *slog.Logger) *ReportService {
	return &ReportService{repo: repo, resultRepo: resultRepo, regRepo: regRepo, log: log}
}

// DraftOrGet 获取/创建草稿报告。
func (s *ReportService) DraftOrGet(ctx context.Context, registrationID uint) (*model.Report, error) {
	if report, err := s.repo.FindByRegistration(registrationID); err == nil {
		return report, nil
	}
	reg, err := s.regRepo.FindByID(registrationID)
	if err != nil {
		return nil, util.NotFoundError(constants.MsgRegNotFound, err)
	}
	seq, _ := s.repo.Count()
	report := &model.Report{
		RegistrationID: registrationID, ExamineeID: reg.ExamineeID,
		ReportNo: fmt.Sprintf("GB%s%04d", time.Now().Format("20060102"), seq+1),
		Status:   constants.ReportDraft, DoctorID: reg.RegisterUserID,
	}
	if err := s.repo.Create(report); err != nil {
		return nil, util.LogError(s.log, constants.LOG_REPORT_DRAFTED, fmt.Errorf("create report: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_REPORT_DRAFTED, "report_id", report.ID, "report_no", report.ReportNo)
	return report, nil
}

// Generate 汇总结果生成报告。
func (s *ReportService) Generate(ctx context.Context, reportID, doctorID uint, conclusion, healthAdvice, followUp string) (*model.Report, error) {
	report, err := s.repo.FindByID(reportID)
	if err != nil {
		return nil, util.NotFoundError(constants.MsgReportNotFound, err)
	}
	if report.Status != constants.ReportDraft && report.Status != constants.ReportGenerated {
		return nil, util.NewAppError(constants.CodeReportStatus, 409, constants.MsgReportStatusInvalid, errors.New("status not draft"))
	}
	results, err := s.resultRepo.ListByRegistration(report.RegistrationID)
	if err != nil {
		return nil, err
	}
	// 汇总结论（缺省生成）
	if conclusion == "" {
		abnormal := 0
		for _, r := range results {
			if r.IsAbnormal {
				abnormal++
			}
		}
		if abnormal > 0 {
			conclusion = fmt.Sprintf("本次体检共发现 %d 项异常指标，建议复查并咨询专科医生。", abnormal)
		} else {
			conclusion = "本次体检各项指标未见明显异常，请继续保持健康生活方式。"
		}
	}
	report.Conclusion = conclusion
	report.HealthAdvice = healthAdvice
	report.FollowUpReminder = followUp
	report.DoctorID = doctorID
	report.Status = constants.ReportGenerated
	now := time.Now()
	report.GeneratedAt = &now
	// 生成 PDF
	pdfBytes, err := util.GeneratePDF(util.PDFReport{
		Title:    "体检报告 " + report.ReportNo,
		SubTitle: fmt.Sprintf("体检人：%s", report.Examinee.Name),
		Header:   []string{"项目", "结果", "参考值", "异常"},
		Rows:     reportRows(results),
		Footer:   []string{"结论：" + conclusion, "健康建议：" + healthAdvice, fmt.Sprintf("生成时间：%s", util.FormatTime(now))},
	})
	if err != nil {
		return nil, util.LogError(s.log, constants.LOG_REPORT_PDF_GENERATED, fmt.Errorf("generate pdf: %w", err))
	}
	// 保存 PDF（内存落盘到 uploads 目录）
	if err := savePDF(report.ReportNo, pdfBytes); err != nil {
		return nil, util.LogError(s.log, constants.LOG_REPORT_PDF_GENERATED, fmt.Errorf("save pdf: %w", err))
	}
	report.PDFURL = "/reports/" + report.ReportNo + ".pdf"
	if err := s.repo.Update(report); err != nil {
		return nil, util.LogError(s.log, constants.LOG_REPORT_GENERATED, fmt.Errorf("update report: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_REPORT_GENERATED, "report_id", reportID)
	return report, nil
}

// Review 审核报告。
func (s *ReportService) Review(ctx context.Context, reportID uint) (*model.Report, error) {
	report, err := s.repo.FindByID(reportID)
	if err != nil {
		return nil, util.NotFoundError(constants.MsgReportNotFound, err)
	}
	if report.Status != constants.ReportGenerated {
		return nil, util.NewAppError(constants.CodeReportStatus, 409, constants.MsgReportStatusInvalid, errors.New("status not generated"))
	}
	report.Status = constants.ReportReviewed
	if err := s.repo.Update(report); err != nil {
		return nil, util.LogError(s.log, constants.LOG_REPORT_REVIEWED, fmt.Errorf("review report: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_REPORT_REVIEWED, "report_id", reportID)
	return report, nil
}

// Publish 发布报告。
func (s *ReportService) Publish(ctx context.Context, reportID uint) (*model.Report, error) {
	report, err := s.repo.FindByID(reportID)
	if err != nil {
		return nil, util.NotFoundError(constants.MsgReportNotFound, err)
	}
	if report.Status != constants.ReportReviewed {
		return nil, util.NewAppError(constants.CodeReportStatus, 409, constants.MsgReportStatusInvalid, errors.New("status not reviewed"))
	}
	report.Status = constants.ReportPublished
	if err := s.repo.Update(report); err != nil {
		return nil, util.LogError(s.log, constants.LOG_REPORT_PUBLISH_FAILED, fmt.Errorf("publish report: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_REPORT_PUBLISHED, "report_id", reportID)
	return report, nil
}

// List 分页查询报告。
func (s *ReportService) List(ctx context.Context, status string, page, pageSize int) ([]model.Report, int64, error) {
	return s.repo.List(status, page, pageSize)
}

// Get 查询报告详情。
func (s *ReportService) Get(ctx context.Context, id uint) (*model.Report, error) {
	report, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError(constants.MsgReportNotFound, err)
		}
		return nil, err
	}
	return report, nil
}

// PDFBytes 返回报告 PDF 内容。
func (s *ReportService) PDFBytes(ctx context.Context, id uint) ([]byte, error) {
	report, err := s.repo.FindByID(id)
	if err != nil {
		return nil, util.NotFoundError(constants.MsgReportNotFound, err)
	}
	results, err := s.resultRepo.ListByRegistration(report.RegistrationID)
	if err != nil {
		return nil, err
	}
	return util.GeneratePDF(util.PDFReport{
		Title:    "体检报告 " + report.ReportNo,
		SubTitle: fmt.Sprintf("体检人：%s", report.Examinee.Name),
		Header:   []string{"项目", "结果", "参考值", "异常"},
		Rows:     reportRows(results),
		Footer:   []string{"结论：" + report.Conclusion},
	})
}

func reportRows(results []model.ExamResult) []util.PDFRow {
	rows := make([]util.PDFRow, 0, len(results))
	for _, r := range results {
		abnormal := "否"
		if r.IsAbnormal {
			abnormal = "是"
		}
		rows = append(rows, util.PDFRow{Columns: []string{r.PackageItem.ItemName, r.ResultValue, r.PackageItem.RefValueRange, abnormal}})
	}
	return rows
}

func savePDF(reportNo string, content []byte) error {
	return util.WriteFile("/app/reports/"+reportNo+".pdf", content)
}
