package repository

import (
	"errors"

	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

// AbnormalMetricRepository 异常指标仓储。
type AbnormalMetricRepository struct{ db *gorm.DB }

// NewAbnormalMetricRepository 构造异常指标仓储。
func NewAbnormalMetricRepository(db *gorm.DB) *AbnormalMetricRepository {
	return &AbnormalMetricRepository{db: db}
}

// WithTx 使用事务连接构造仓储。
func (r *AbnormalMetricRepository) WithTx(tx *gorm.DB) *AbnormalMetricRepository {
	return &AbnormalMetricRepository{db: tx}
}

func (r *AbnormalMetricRepository) Create(m *model.AbnormalMetric) error { return r.db.Create(m).Error }

func (r *AbnormalMetricRepository) List(examineeID uint, page, pageSize int) ([]model.AbnormalMetric, int64, error) {
	q := r.db.Model(&model.AbnormalMetric{})
	if examineeID > 0 {
		q = q.Where("examinee_id = ?", examineeID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.AbnormalMetric
	err := r.db.Preload("PackageItem").Where("examinee_id = ?", examineeID).Order("id desc").Offset((page-1)*pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *AbnormalMetricRepository) FindByID(id uint) (*model.AbnormalMetric, error) {
	var m model.AbnormalMetric
	if err := r.db.Preload("PackageItem").First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *AbnormalMetricRepository) UpdateFollowUp(id uint, status, advice string) error {
	return r.db.Model(&model.AbnormalMetric{}).Where("id = ?", id).Updates(map[string]any{
		"follow_up_status": status, "specialist_advice": advice,
	}).Error
}

func (r *AbnormalMetricRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.AbnormalMetric{}).Count(&count).Error
	return count, err
}
