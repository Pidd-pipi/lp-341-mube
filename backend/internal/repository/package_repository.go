package repository

import (
	"errors"

	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

// PackageRepository 套餐仓储。
type PackageRepository struct{ db *gorm.DB }

// NewPackageRepository 构造套餐仓储。
func NewPackageRepository(db *gorm.DB) *PackageRepository { return &PackageRepository{db: db} }

func (r *PackageRepository) Create(pkg *model.Package) error { return r.db.Create(pkg).Error }

func (r *PackageRepository) FindByID(id uint) (*model.Package, error) {
	var pkg model.Package
	if err := r.db.First(&pkg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &pkg, nil
}

func (r *PackageRepository) List(status string, page, pageSize int) ([]model.Package, int64, error) {
	q := r.db.Model(&model.Package{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var pkgs []model.Package
	q2 := r.db.Order("id asc")
	if status != "" {
		q2 = q2.Where("status = ?", status)
	}
	err := q2.Offset((page-1)*pageSize).Limit(pageSize).Find(&pkgs).Error
	return pkgs, total, err
}

func (r *PackageRepository) Update(pkg *model.Package) error { return r.db.Save(pkg).Error }

func (r *PackageRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Package{}).Count(&count).Error
	return count, err
}

// CountGroupByType 按套餐类型统计成交量。
func (r *PackageRepository) CountGroupByType() ([]model.NameCount, error) {
	var rows []model.NameCount
	err := r.db.Model(&model.Package{}).Select("name, count(*) as count").Group("name").Scan(&rows).Error
	return rows, err
}
