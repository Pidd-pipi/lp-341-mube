package repository

import (
	"errors"

	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

// EnterpriseRepository 企业仓储。
type EnterpriseRepository struct{ db *gorm.DB }

// NewEnterpriseRepository 构造企业仓储。
func NewEnterpriseRepository(db *gorm.DB) *EnterpriseRepository { return &EnterpriseRepository{db: db} }

func (r *EnterpriseRepository) Create(e *model.Enterprise) error { return r.db.Create(e).Error }

func (r *EnterpriseRepository) List(page, pageSize int) ([]model.Enterprise, int64, error) {
	var total int64
	if err := r.db.Model(&model.Enterprise{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Enterprise
	err := r.db.Order("id desc").Offset((page-1)*pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *EnterpriseRepository) FindByID(id uint) (*model.Enterprise, error) {
	var e model.Enterprise
	if err := r.db.First(&e, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}
