package repository

import (
	"errors"

	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

// ReportRepository 报告仓储。
type ReportRepository struct{ db *gorm.DB }

// NewReportRepository 构造报告仓储。
func NewReportRepository(db *gorm.DB) *ReportRepository { return &ReportRepository{db: db} }

func (r *ReportRepository) Create(report *model.Report) error { return r.db.Create(report).Error }

func (r *ReportRepository) FindByID(id uint) (*model.Report, error) {
	var report model.Report
	if err := r.db.Preload("Examinee").First(&report, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &report, nil
}

func (r *ReportRepository) FindByRegistration(regID uint) (*model.Report, error) {
	var report model.Report
	if err := r.db.Preload("Examinee").Where("registration_id = ?", regID).First(&report).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &report, nil
}

func (r *ReportRepository) List(status string, page, pageSize int) ([]model.Report, int64, error) {
	q := r.db.Model(&model.Report{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Report
	q2 := r.db.Preload("Examinee").Order("id desc")
	if status != "" {
		q2 = q2.Where("status = ?", status)
	}
	err := q2.Offset((page-1)*pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *ReportRepository) Update(report *model.Report) error { return r.db.Save(report).Error }

func (r *ReportRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Report{}).Where("id = ?", id).Update("status", status).Error
}

func (r *ReportRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Report{}).Count(&count).Error
	return count, err
}
