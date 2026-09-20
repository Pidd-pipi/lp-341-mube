package repository

import (
	"errors"

	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

// ReportWithdrawRepository 撤回重签申请仓储。
type ReportWithdrawRepository struct{ db *gorm.DB }

// NewReportWithdrawRepository 构造撤回重签申请仓储。
func NewReportWithdrawRepository(db *gorm.DB) *ReportWithdrawRepository {
	return &ReportWithdrawRepository{db: db}
}

// DB 暴露底层句柄，供审批事务复用。
func (r *ReportWithdrawRepository) DB() *gorm.DB { return r.db }

// CreatePendingIndex 创建“同一报告仅一份待审批申请”的部分唯一索引
// （Postgres 与 SQLite 均支持 WHERE 子句的部分索引）。
func (r *ReportWithdrawRepository) CreatePendingIndex() error {
	return r.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_report_withdraw_one_pending
		ON report_withdraws(report_id) WHERE status = 'pending'`).Error
}

func (r *ReportWithdrawRepository) Create(w *model.ReportWithdraw) error {
	return r.db.Create(w).Error
}

func (r *ReportWithdrawRepository) FindByID(id uint) (*model.ReportWithdraw, error) {
	var w model.ReportWithdraw
	if err := r.db.Preload("Report").Preload("Report.Examinee").First(&w, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &w, nil
}

// FindPendingByReport 返回某报告当前待审批的申请（无则 ErrNotFound）。
func (r *ReportWithdrawRepository) FindPendingByReport(reportID uint) (*model.ReportWithdraw, error) {
	var w model.ReportWithdraw
	if err := r.db.Where("report_id = ? AND status = ?", reportID, "pending").First(&w).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &w, nil
}

// ExistsPendingByReport 判断某报告是否已存在待审批申请（重复申请冲突用）。
func (r *ReportWithdrawRepository) ExistsPendingByReport(reportID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.ReportWithdraw{}).
		Where("report_id = ? AND status = ?", reportID, "pending").Count(&count).Error
	return count > 0, err
}

// ListByReports 按报告分页查询申请并级联报告；reportID>0 时只看该报告。
func (r *ReportWithdrawRepository) ListByReports(status string, reportID uint, page, pageSize int) ([]model.ReportWithdraw, int64, error) {
	q := r.db.Model(&model.ReportWithdraw{})
	if status != "" {
		q = q.Where("report_withdraws.status = ?", status)
	}
	if reportID > 0 {
		q = q.Where("report_withdraws.report_id = ?", reportID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.ReportWithdraw
	err := q.Preload("Report").Preload("Report.Examinee").
		Order("report_withdraws.id desc").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

// CASReview 审批 CAS：仅当申请仍为 pending 时更新为目标状态，
// 返回 rowsAffected==0 表示已被并发审批或状态已变（冲突，状态不变）。
func (r *ReportWithdrawRepository) CASReview(id, reviewerID, newReportID uint, status, comment string, reviewedAt any) (int64, error) {
	updates := map[string]any{
		"status":         status,
		"reviewer_id":    reviewerID,
		"review_comment": comment,
		"reviewed_at":    reviewedAt,
		"new_report_id":  newReportID,
	}
	res := r.db.Model(&model.ReportWithdraw{}).
		Where("id = ? AND status = ?", id, "pending").
		Updates(updates)
	return res.RowsAffected, res.Error
}

// UpdateNewReportID 批准事务内待重签版本创建成功后回写关联。
func (r *ReportWithdrawRepository) UpdateNewReportID(id, newReportID uint) error {
	return r.db.Model(&model.ReportWithdraw{}).
		Where("id = ?", id).Update("new_report_id", newReportID).Error
}
