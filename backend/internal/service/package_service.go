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

// PackageService 套餐服务（含检查项目维护）。
type PackageService struct {
	pkgRepo  *repository.PackageRepository
	itemRepo *repository.PackageItemRepository
	log      *slog.Logger
}

// NewPackageService 构造套餐服务。
func NewPackageService(pkgRepo *repository.PackageRepository, itemRepo *repository.PackageItemRepository, log *slog.Logger) *PackageService {
	return &PackageService{pkgRepo: pkgRepo, itemRepo: itemRepo, log: log}
}

// Create 创建套餐。
func (s *PackageService) Create(ctx context.Context, name, packageType string, price float64, status, description string) (*model.Package, error) {
	pkg := &model.Package{Name: name, PackageType: packageType, Price: price, Status: status, Description: description}
	if err := s.pkgRepo.Create(pkg); err != nil {
		return nil, util.LogError(s.log, constants.LOG_PACKAGE_CREATED, fmt.Errorf("create package: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_PACKAGE_CREATED, "package_id", pkg.ID)
	return pkg, nil
}

// Update 更新套餐。
func (s *PackageService) Update(ctx context.Context, id uint, name, packageType string, price float64, status, description string) (*model.Package, error) {
	pkg, err := s.pkgRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError(constants.MsgPackageNotFound, err)
		}
		return nil, err
	}
	if name != "" {
		pkg.Name = name
	}
	if packageType != "" {
		pkg.PackageType = packageType
	}
	pkg.Price = price
	if status != "" {
		pkg.Status = status
	}
	pkg.Description = description
	if err := s.pkgRepo.Update(pkg); err != nil {
		return nil, util.LogError(s.log, constants.LOG_PACKAGE_UPDATED, fmt.Errorf("update package: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_PACKAGE_UPDATED, "package_id", id)
	return pkg, nil
}

// Get 查询套餐详情（含项目）。
func (s *PackageService) Get(ctx context.Context, id uint) (*model.Package, []model.PackageItem, error) {
	pkg, err := s.pkgRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, nil, util.NotFoundError(constants.MsgPackageNotFound, err)
		}
		return nil, nil, err
	}
	items, err := s.itemRepo.ListByPackage(id)
	if err != nil {
		return nil, nil, err
	}
	return pkg, items, nil
}

// List 分页查询套餐。
func (s *PackageService) List(ctx context.Context, status string, page, pageSize int) ([]model.Package, int64, error) {
	return s.pkgRepo.List(status, page, pageSize)
}

// AddItem 添加检查项目。
func (s *PackageService) AddItem(ctx context.Context, packageID uint, item *model.PackageItem) (*model.PackageItem, error) {
	if _, err := s.pkgRepo.FindByID(packageID); err != nil {
		return nil, util.NotFoundError(constants.MsgPackageNotFound, err)
	}
	item.PackageID = packageID
	if err := s.itemRepo.Create(item); err != nil {
		return nil, util.LogError(s.log, constants.LOG_PACKAGE_ITEM_ADDED, fmt.Errorf("create package item: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_PACKAGE_ITEM_ADDED, "item_id", item.ID, "package_id", packageID)
	return item, nil
}

// ListItems 查询套餐项目。
func (s *PackageService) ListItems(ctx context.Context, packageID uint) ([]model.PackageItem, error) {
	return s.itemRepo.ListByPackage(packageID)
}

// UpdateItem 更新检查项目。
func (s *PackageService) UpdateItem(ctx context.Context, id uint, data *model.PackageItem) (*model.PackageItem, error) {
	item, err := s.itemRepo.FindByID(id)
	if err != nil {
		return nil, util.NotFoundError("检查项目（PackageItem）不存在", err)
	}
	if data.ItemName != "" {
		item.ItemName = data.ItemName
	}
	if data.ItemGroup != "" {
		item.ItemGroup = data.ItemGroup
	}
	if data.RefValueRange != "" {
		item.RefValueRange = data.RefValueRange
	}
	if data.Department != "" {
		item.Department = data.Department
	}
	if data.SortOrder != 0 {
		item.SortOrder = data.SortOrder
	}
	if err := s.itemRepo.Update(item); err != nil {
		return nil, util.LogError(s.log, constants.LOG_PACKAGE_ITEM_UPDATED, fmt.Errorf("update package item: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_PACKAGE_ITEM_UPDATED, "item_id", id)
	return item, nil
}

// DeleteItem 删除检查项目。
func (s *PackageService) DeleteItem(ctx context.Context, id uint) error {
	return s.itemRepo.Delete(id)
}
