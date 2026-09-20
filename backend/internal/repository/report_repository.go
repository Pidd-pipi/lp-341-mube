package repository

import (
	"errors"

	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReportRepository 报告仓储。
type ReportRepository struct{ db *gorm.DB }

// NewReportRepository 构造报告仓储。
func NewReportRepository(db *gorm.DB) *ReportRepository { return &ReportRepository{db: db} }

// WithDB 基于指定事务句柄构造同构仓储（service 事务内使用）。
func (r *ReportRepository) WithDB(tx *gorm.DB) *ReportRepository {
	return &ReportRepository{db: tx}
}

// DB 暴露底层句柄，供 service 组装跨实体事务。
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

// FindByIDForUpdate 事务内对报告行加排他锁（SELECT ... FOR UPDATE），串行化并发撤回申请/审批。
func (r *ReportRepository) FindByIDForUpdate(id uint) (*model.Report, error) {
	var report model.Report
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&report, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &report, nil
}

// FindByRegistration 返回该登记下的最新版本报告。
func (r *ReportRepository) FindByRegistration(regID uint) (*model.Report, error) {
	var report model.Report
	if err := r.db.Preload("Examinee").Where("registration_id = ?", regID).
		Order("version desc, id desc").First(&report).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &report, nil
}

// ListVersionChain 返回版本链：根版本为自身的初代及其全部子孙版本。
func (r *ReportRepository) ListVersionChain(rootID uint) ([]model.Report, error) {
	var items []model.Report
	err := r.db.Preload("Examinee").
		Where("root_report_id = ? OR id = ?", rootID, rootID).
		Order("version asc, id asc").Find(&items).Error
	return items, err
}

// CountByRoot 返回版本链版本数（含根版本）。
func (r *ReportRepository) CountByRoot(rootID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Report{}).
		Where("root_report_id = ? OR id = ?", rootID, rootID).Count(&count).Error
	return count, err
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

// UpdateStatusIf 条件更新状态；返回受影响行数（0 表示状态已被并发改动，调用方应返回冲突）。
func (r *ReportRepository) UpdateStatusIf(id uint, fromStatuses []string, toStatus string) (int64, error) {
	res := r.db.Model(&model.Report{}).
		Where("id = ? AND status IN ?", id, fromStatuses).
		Update("status", toStatus)
	return res.RowsAffected, res.Error
}

// RestoreStatusIf 条件恢复状态（驳回撤回时恢复已发布）；返回受影响行数。
func (r *ReportRepository) RestoreStatusIf(id uint, expectStatus, toStatus string) (int64, error) {
	res := r.db.Model(&model.Report{}).
		Where("id = ? AND status = ?", id, expectStatus).
		Update("status", toStatus)
	return res.RowsAffected, res.Error
}

func (r *ReportRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Report{}).Count(&count).Error
	return count, err
}

// CountActive 有效报告数：撤回批准后原版归档为 withdrawn，不计入，避免版本链重复计数。
func (r *ReportRepository) CountActive() (int64, error) {
	var count int64
	err := r.db.Model(&model.Report{}).Where("status <> ?", "withdrawn").Count(&count).Error
	return count, err
}
