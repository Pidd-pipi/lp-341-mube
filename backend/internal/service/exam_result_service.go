package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

// ExamResultService 检查结果服务：录入、参考值比对、异常标记。
type ExamResultService struct {
	repo       *repository.ExamResultRepository
	regRepo    *repository.RegistrationRepository
	metricRepo *repository.AbnormalMetricRepository
	log        *slog.Logger
}

// NewExamResultService 构造检查结果服务。
func NewExamResultService(repo *repository.ExamResultRepository, regRepo *repository.RegistrationRepository, metricRepo *repository.AbnormalMetricRepository, log *slog.Logger) *ExamResultService {
	return &ExamResultService{repo: repo, regRepo: regRepo, metricRepo: metricRepo, log: log}
}

// EnterInput 结果录入入参。
type EnterInput struct {
	ResultValue string `json:"result_value"`
	ResultText  string `json:"result_text"`
	ImageURL    string `json:"image_url"`
}

// Enter 录入结果并与参考值比对标红异常。
func (s *ExamResultService) Enter(ctx context.Context, resultID, doctorID uint, input EnterInput) (*model.ExamResult, error) {
	res, err := s.repo.FindByID(resultID)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError(constants.MsgResultNotFound, err)
		}
		return nil, err
	}
	res.ResultValue = input.ResultValue
	res.ResultText = input.ResultText
	if input.ImageURL != "" {
		res.ImageURL = input.ImageURL
	}
	res.DoctorID = doctorID
	res.Status = constants.ResultEntered
	now := time.Now()
	res.EnteredAt = &now
	res.IsAbnormal = IsAbnormal(res.PackageItem.RefValueRange, input.ResultValue)
	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.WithTx(tx).Update(res); err != nil {
			return fmt.Errorf("update exam result: %w", err)
		}
		if res.IsAbnormal {
			metric := &model.AbnormalMetric{
				ExamineeID: res.ExamineeID, PackageItemID: res.PackageItemID,
				AbnormalLevel: GuessAbnormalLevel(input.ResultValue), Value: input.ResultValue, RefValueRange: res.PackageItem.RefValueRange,
				TrendJSON: "[]", FollowUpStatus: constants.FollowUpPending,
			}
			if err := s.metricRepo.WithTx(tx).Create(metric); err != nil {
				return fmt.Errorf("create abnormal metric: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return nil, util.LogError(s.log, constants.LOG_EXAM_RESULT_ENTERED, err)
	}
	s.log.InfoContext(ctx, constants.LOG_EXAM_RESULT_ENTERED, "result_id", resultID, "abnormal", res.IsAbnormal)
	if res.IsAbnormal {
		s.log.InfoContext(ctx, constants.LOG_EXAM_RESULT_ABNORMAL, "result_id", resultID, "item", res.PackageItem.ItemName)
	}
	return res, nil
}

// Review 审核结果。
func (s *ExamResultService) Review(ctx context.Context, resultID uint) error {
	res, err := s.repo.FindByID(resultID)
	if err != nil {
		return util.NotFoundError(constants.MsgResultNotFound, err)
	}
	res.Status = constants.ResultReviewed
	if err := s.repo.Update(res); err != nil {
		return util.LogError(s.log, constants.LOG_EXAM_RESULT_REVIEWED, fmt.Errorf("review result: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_EXAM_RESULT_REVIEWED, "result_id", resultID)
	return nil
}

// ListByRegistration 查询登记下全部结果。
func (s *ExamResultService) ListByRegistration(ctx context.Context, regID uint) ([]model.ExamResult, error) {
	return s.repo.ListByRegistration(regID)
}

// ListPending 待录入/待审核工作台。
func (s *ExamResultService) ListPending(ctx context.Context, page, pageSize int) ([]model.ExamResult, int64, error) {
	return s.repo.ListPending(page, pageSize)
}

// IsAbnormal 数值结果与参考值范围比对（支持 "10-20" / ">10" / "<5" 等格式）。
func IsAbnormal(refRange, value string) bool {
	refRange = strings.TrimSpace(refRange)
	value = strings.TrimSpace(value)
	if refRange == "" || value == "" {
		return false
	}
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}
	// 范围 10-20
	if idx := strings.Index(refRange, "-"); idx > 0 && !strings.Contains(refRange, ">") && !strings.Contains(refRange, "<") {
		lo, err1 := strconv.ParseFloat(strings.TrimSpace(refRange[:idx]), 64)
		hi, err2 := strconv.ParseFloat(strings.TrimSpace(refRange[idx+1:]), 64)
		if err1 == nil && err2 == nil {
			return v < lo || v > hi
		}
	}
	// >10 或 <5
	if strings.HasPrefix(refRange, ">") {
		if lim, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(refRange, ">")), 64); err == nil {
			return v <= lim
		}
	}
	if strings.HasPrefix(refRange, "<") {
		if lim, err := strconv.ParseFloat(strings.TrimSpace(strings.TrimPrefix(refRange, "<")), 64); err == nil {
			return v >= lim
		}
	}
	return false
}

// GuessAbnormalLevel 根据偏离程度猜测异常等级。
func GuessAbnormalLevel(value string) string {
	v, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return constants.AbnormalMild
	}
	abs := v
	if abs < 0 {
		abs = -abs
	}
	switch {
	case abs > 1000:
		return constants.AbnormalSevere
	case abs > 200:
		return constants.AbnormalModerate
	default:
		return constants.AbnormalMild
	}
}
