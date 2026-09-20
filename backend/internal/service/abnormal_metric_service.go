package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
	"github.com/blueship581/gbcheckup/internal/util"
)

// AbnormalMetricService 异常指标服务：记录、趋势对比、复查跟踪。
type AbnormalMetricService struct {
	repo *repository.AbnormalMetricRepository
	log  *slog.Logger
}

// NewAbnormalMetricService 构造异常指标服务。
func NewAbnormalMetricService(repo *repository.AbnormalMetricRepository, log *slog.Logger) *AbnormalMetricService {
	return &AbnormalMetricService{repo: repo, log: log}
}

// List 分页查询异常指标。
func (s *AbnormalMetricService) List(ctx context.Context, examineeID uint, page, pageSize int) ([]model.AbnormalMetric, int64, error) {
	return s.repo.List(examineeID, page, pageSize)
}

// UpdateFollowUp 更新复查跟踪与专科建议。
func (s *AbnormalMetricService) UpdateFollowUp(ctx context.Context, id uint, status, advice string) (*model.AbnormalMetric, error) {
	if status != constants.FollowUpPending && status != constants.FollowUpDone {
		return nil, util.BadRequest("复查状态（AbnormalMetric.follow_up_status）不合法", errors.New("invalid status"))
	}
	if err := s.repo.UpdateFollowUp(id, status, advice); err != nil {
		return nil, util.LogError(s.log, constants.LOG_ABNORMAL_METRIC_FOLLOWUP, fmt.Errorf("update follow-up: %w", err))
	}
	m, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	s.log.InfoContext(ctx, constants.LOG_ABNORMAL_METRIC_FOLLOWUP, "metric_id", id, "status", status)
	return m, nil
}
