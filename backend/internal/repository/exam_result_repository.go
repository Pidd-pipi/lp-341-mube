package repository

import (
	"errors"

	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

// ExamResultRepository 检查结果仓储。
type ExamResultRepository struct{ db *gorm.DB }

// NewExamResultRepository 构造检查结果仓储。
func NewExamResultRepository(db *gorm.DB) *ExamResultRepository { return &ExamResultRepository{db: db} }

// WithTx 使用事务连接构造仓储。
func (r *ExamResultRepository) WithTx(tx *gorm.DB) *ExamResultRepository {
	return &ExamResultRepository{db: tx}
}

// Transaction 在事务内执行 fn，任一步返回 error 则整体回滚。
func (r *ExamResultRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *ExamResultRepository) Create(res *model.ExamResult) error { return r.db.Create(res).Error }

func (r *ExamResultRepository) CreateBatch(items []model.ExamResult) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

func (r *ExamResultRepository) FindByID(id uint) (*model.ExamResult, error) {
	var res model.ExamResult
	if err := r.db.Preload("PackageItem").First(&res, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &res, nil
}

func (r *ExamResultRepository) ListByRegistration(regID uint) ([]model.ExamResult, error) {
	var items []model.ExamResult
	err := r.db.Preload("PackageItem").Where("registration_id = ?", regID).Order("id asc").Find(&items).Error
	return items, err
}

func (r *ExamResultRepository) ListPending(page, pageSize int) ([]model.ExamResult, int64, error) {
	var total int64
	if err := r.db.Model(&model.ExamResult{}).Where("status != ?", "reviewed").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.ExamResult
	err := r.db.Preload("PackageItem").Where("status != ?", "reviewed").Order("id asc").Offset((page-1)*pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *ExamResultRepository) Update(res *model.ExamResult) error { return r.db.Save(res).Error }

func (r *ExamResultRepository) CountAbnormal() (int64, error) {
	var count int64
	err := r.db.Model(&model.ExamResult{}).Where("is_abnormal = ?", true).Count(&count).Error
	return count, err
}

// CountAbnormalGroupByItem 异常指标 TOP 统计。
func (r *ExamResultRepository) CountAbnormalGroupByItem() ([]model.NameCount, error) {
	var rows []model.NameCount
	err := r.db.Model(&model.ExamResult{}).
		Select("package_items.item_name as name, count(exam_results.id) as count").
		Joins("JOIN package_items ON package_items.id = exam_results.package_item_id").
		Where("exam_results.is_abnormal = ?", true).
		Group("package_items.item_name").Order("count desc").Limit(10).Scan(&rows).Error
	return rows, err
}
