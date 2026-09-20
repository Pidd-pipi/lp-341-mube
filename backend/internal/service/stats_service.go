package service

import (
	"context"
	"log/slog"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
)

// StatsService 运营统计服务。
type StatsService struct {
	pkgRepo     *repository.PackageRepository
	regRepo     *repository.RegistrationRepository
	reportRepo  *repository.ReportRepository
	resultRepo  *repository.ExamResultRepository
	metricRepo  *repository.AbnormalMetricRepository
	itemRepo    *repository.PackageItemRepository
	log         *slog.Logger
}

// NewStatsService 构造统计服务。
func NewStatsService(pkgRepo *repository.PackageRepository, regRepo *repository.RegistrationRepository, reportRepo *repository.ReportRepository, resultRepo *repository.ExamResultRepository, metricRepo *repository.AbnormalMetricRepository, itemRepo *repository.PackageItemRepository, log *slog.Logger) *StatsService {
	return &StatsService{pkgRepo: pkgRepo, regRepo: regRepo, reportRepo: reportRepo, resultRepo: resultRepo, metricRepo: metricRepo, itemRepo: itemRepo, log: log}
}

// Dashboard 运营统计。
func (s *StatsService) Dashboard(ctx context.Context) (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}
	stats.PackageCount, _ = s.pkgRepo.Count()
	stats.RegistrationCount, _ = s.regRepo.Count()
	stats.ReportCount, _ = s.reportRepo.Count()
	stats.AbnormalCount, _ = s.metricRepo.Count()
	revenue, _ := s.regRepo.RevenueByMonth()
	total := 0.0
	for _, m := range revenue {
		total += m.Amount
	}
	stats.Revenue = total
	stats.MonthlyRevenue = revenue
	stats.PackageSold, _ = s.regRepo.CountGroupByPackage()
	stats.DeptWorkload, _ = s.itemRepo.CountGroupByDepartment()
	stats.AbnormalTop, _ = s.resultRepo.CountAbnormalGroupByItem()
	s.log.InfoContext(ctx, constants.LOG_STATS_DASHBOARD, "registrations", stats.RegistrationCount)
	return stats, nil
}
