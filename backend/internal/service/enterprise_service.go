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

// EnterpriseService 团检企业与订单服务。
type EnterpriseService struct {
	entRepo *repository.EnterpriseRepository
	orderRepo *repository.GroupOrderRepository
	pkgRepo *repository.PackageRepository
	log     *slog.Logger
}

// NewEnterpriseService 构造团检服务。
func NewEnterpriseService(entRepo *repository.EnterpriseRepository, orderRepo *repository.GroupOrderRepository, pkgRepo *repository.PackageRepository, log *slog.Logger) *EnterpriseService {
	return &EnterpriseService{entRepo: entRepo, orderRepo: orderRepo, pkgRepo: pkgRepo, log: log}
}

// CreateEnterprise 创建企业。
func (s *EnterpriseService) CreateEnterprise(ctx context.Context, e *model.Enterprise) (*model.Enterprise, error) {
	if err := s.entRepo.Create(e); err != nil {
		return nil, util.LogError(s.log, constants.LOG_ENTERPRISE_CREATED, fmt.Errorf("create enterprise: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_ENTERPRISE_CREATED, "enterprise_id", e.ID)
	return e, nil
}

// ListEnterprises 企业列表。
func (s *EnterpriseService) ListEnterprises(ctx context.Context, page, pageSize int) ([]model.Enterprise, int64, error) {
	return s.entRepo.List(page, pageSize)
}

// CreateOrder 创建团检订单。
func (s *EnterpriseService) CreateOrder(ctx context.Context, enterpriseID, packageID uint, count int) (*model.GroupOrder, error) {
	if _, err := s.entRepo.FindByID(enterpriseID); err != nil {
		return nil, util.NotFoundError("团检企业（Enterprise）不存在", err)
	}
	if _, err := s.pkgRepo.FindByID(packageID); err != nil {
		return nil, util.NotFoundError(constants.MsgPackageNotFound, err)
	}
	order := &model.GroupOrder{EnterpriseID: enterpriseID, PackageID: packageID, ExamineeCount: count, Status: constants.GroupOrderPending}
	if err := s.orderRepo.Create(order); err != nil {
		return nil, util.LogError(s.log, constants.LOG_GROUP_ORDER_CREATED, fmt.Errorf("create group order: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_GROUP_ORDER_CREATED, "order_id", order.ID)
	return order, nil
}

// ListOrders 团检订单列表。
func (s *EnterpriseService) ListOrders(ctx context.Context, page, pageSize int) ([]model.GroupOrder, int64, error) {
	return s.orderRepo.List(page, pageSize)
}

// DeliverReports 报告批量交付。
func (s *EnterpriseService) DeliverReports(ctx context.Context, orderID uint) (*model.GroupOrder, error) {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError("团检订单（GroupOrder）不存在", err)
		}
		return nil, err
	}
	order.ReportDeliveryStatus = "delivered"
	order.Status = constants.GroupOrderDone
	if err := s.orderRepo.Update(order); err != nil {
		return nil, util.LogError(s.log, constants.LOG_GROUP_ORDER_DELIVERED, fmt.Errorf("deliver reports: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_GROUP_ORDER_DELIVERED, "order_id", orderID)
	return order, nil
}
