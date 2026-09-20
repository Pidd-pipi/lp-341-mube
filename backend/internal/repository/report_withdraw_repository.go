package repository

import (
	"errors"
	"time"

	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReportWithdrawRepository 报告撤回申请仓储。
type ReportWithdrawRepository struct{ db *gorm.DB }

// NewReportWithdrawRepository 构造撤回申请仓储。
func NewReportWithdrawRepository(db *gorm.DB) *ReportWithdrawRepository {
	return &ReportWithdrawRepository{db: db}
}

// WithDB 基于指定事务句柄构造同构仓储（service 事务内使用）。
func (r *ReportWithdrawRepository) WithDB(tx *gorm.DB) *ReportWithdrawRepository {
	return &ReportWithdrawRepository{db: tx}
}

func (r *ReportWithdrawRepository) Create(req *model.ReportWithdrawRequest) error {
	return r.db.Create(req).Error
}

// CountPendingByReport 统计某报告的待审批申请数（事务内加锁检查，保证唯一）。
func (r *ReportWithdrawRepository) CountPendingByReport(reportID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.ReportWithdrawRequest{}).
		Where("report_id = ? AND status = ?", reportID, "pending").
		Count(&count).Error
	return count, err
}

// FindPendingByReport 查询某报告当前待审批申请，无则返回 util.ErrNotFound。
func (r *ReportWithdrawRepository) FindPendingByReport(reportID uint) (*model.ReportWithdrawRequest, error) {
	var req model.ReportWithdrawRequest
	if err := r.db.Where("report_id = ? AND status = ?", reportID, "pending").
		Order("id desc").First(&req).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &req, nil
}

func (r *ReportWithdrawRepository) FindByID(id uint) (*model.ReportWithdrawRequest, error) {
	var req model.ReportWithdrawRequest
	if err := r.db.Preload("Report.Examinee").First(&req, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &req, nil
}

// FindByIDForUpdate 事务内对申请行加排他锁，串行化并发审批。
func (r *ReportWithdrawRepository) FindByIDForUpdate(id uint) (*model.ReportWithdrawRequest, error) {
	var req model.ReportWithdrawRequest
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&req, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &req, nil
}

// ApproveIfPending 仅当申请仍为 pending 且未超时时置为 approved；返回受影响行数。
// affected=0 表示重复/并发审批或已超时，调用方返回冲突且状态不变。
func (r *ReportWithdrawRepository) ApproveIfPending(id, reviewerID, newReportID uint, comment string, now time.Time) (int64, error) {
	res := r.db.Model(&model.ReportWithdrawRequest{}).
		Where("id = ? AND status = ? AND expires_at > ?", id, "pending", now).
		Updates(map[string]any{
			"status":         "approved",
			"reviewer_id":    reviewerID,
			"new_report_id":  newReportID,
			"review_comment": comment,
			"reviewed_at":    now,
		})
	return res.RowsAffected, res.Error
}

// RejectIfPending 仅当申请仍为 pending 且未超时时置为 rejected；返回受影响行数。
func (r *ReportWithdrawRepository) RejectIfPending(id, reviewerID uint, comment string, now time.Time) (int64, error) {
	res := r.db.Model(&model.ReportWithdrawRequest{}).
		Where("id = ? AND status = ? AND expires_at > ?", id, "pending", now).
		Updates(map[string]any{
			"status":         "rejected",
			"reviewer_id":    reviewerID,
			"review_comment": comment,
			"reviewed_at":    now,
		})
	return res.RowsAffected, res.Error
}

// ListByReport 查询某报告的全部撤回申请（按时间倒序）。
func (r *ReportWithdrawRepository) ListByReport(reportID uint) ([]model.ReportWithdrawRequest, error) {
	var items []model.ReportWithdrawRequest
	err := r.db.Where("report_id = ?", reportID).Order("id desc").Find(&items).Error
	return items, err
}

// ListExpiredPending 查询已超时但仍为 pending 的申请（可选运维兜底）。
func (r *ReportWithdrawRepository) ListExpiredPending(now time.Time, limit int) ([]model.ReportWithdrawRequest, error) {
	var items []model.ReportWithdrawRequest
	err := r.db.Where("status = ? AND expires_at <= ?", "pending", now).
		Order("id asc").Limit(limit).Find(&items).Error
	return items, err
}
