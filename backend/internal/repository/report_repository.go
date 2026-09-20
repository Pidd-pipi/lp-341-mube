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

// DB 暴露底层句柄，供需要多仓储协同的事务使用。
func (r *ReportRepository) DB() *gorm.DB { return r.db }

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
	// 优先返回非归档版本（已撤回原版排到最后），保证撤回批准后拿到待重签新版本。
	if err := r.db.Preload("Examinee").Where("registration_id = ?", regID).
		Order("CASE status WHEN 'withdrawn' THEN 1 ELSE 0 END").
		Order("id desc").First(&report).Error; err != nil {
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
	err := q2.Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *ReportRepository) Update(report *model.Report) error { return r.db.Save(report).Error }

func (r *ReportRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Report{}).Where("id = ?", id).Update("status", status).Error
}

// CASStatus 以条件更新实现状态机 CAS：仅当报告当前状态属于 expect 之一时才更新，
// 返回 rowsAffected==0 表示状态已被并发改动（冲突）。
func (r *ReportRepository) CASStatus(id uint, expect []string, status string) (int64, error) {
	res := r.db.Model(&model.Report{}).
		Where("id = ? AND status IN ?", id, expect).
		Update("status", status)
	return res.RowsAffected, res.Error
}

// CountActiveByRegistration 统计某登记下非归档（未撤回）的报告数，防止审批冻结期间生成新版本。
func (r *ReportRepository) CountActiveByRegistration(regID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Report{}).
		Where("registration_id = ? AND status <> ?", regID, "withdrawn").
		Count(&count).Error
	return count, err
}

// ListChain 返回版本链：root_report_id 命中（含根本身）的全部版本，按版本号升序。
func (r *ReportRepository) ListChain(rootID uint) ([]model.Report, error) {
	var items []model.Report
	err := r.db.Preload("Examinee").
		Where("root_report_id = ? OR id = ?", rootID, rootID).
		Order("version_no asc, id asc").Find(&items).Error
	return items, err
}

func (r *ReportRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Report{}).Count(&count).Error
	return count, err
}
