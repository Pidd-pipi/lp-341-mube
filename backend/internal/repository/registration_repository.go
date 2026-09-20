package repository

import (
	"errors"

	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

// RegistrationRepository 登记仓储。
type RegistrationRepository struct{ db *gorm.DB }

// NewRegistrationRepository 构造登记仓储。
func NewRegistrationRepository(db *gorm.DB) *RegistrationRepository { return &RegistrationRepository{db: db} }

// WithTx 使用事务连接构造仓储。
func (r *RegistrationRepository) WithTx(tx *gorm.DB) *RegistrationRepository {
	return &RegistrationRepository{db: tx}
}

// Transaction 在事务内执行 fn，任一步返回 error 则整体回滚。
func (r *RegistrationRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return r.db.Transaction(fn)
}

func (r *RegistrationRepository) Create(reg *model.Registration) error { return r.db.Create(reg).Error }

func (r *RegistrationRepository) FindByID(id uint) (*model.Registration, error) {
	var reg model.Registration
	if err := r.db.Preload("Examinee").Preload("Package").First(&reg, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &reg, nil
}

func (r *RegistrationRepository) List(status string, page, pageSize int) ([]model.Registration, int64, error) {
	q := r.db.Model(&model.Registration{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Registration
	q2 := r.db.Preload("Examinee").Preload("Package").Order("id desc")
	if status != "" {
		q2 = q2.Where("status = ?", status)
	}
	err := q2.Offset((page-1)*pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *RegistrationRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Registration{}).Where("id = ?", id).Update("status", status).Error
}

func (r *RegistrationRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Registration{}).Count(&count).Error
	return count, err
}

// CountToday 今日登记人数。
func (r *RegistrationRepository) CountToday() (int64, error) {
	var count int64
	err := r.db.Model(&model.Registration{}).Where("registered_at::date = CURRENT_DATE").Count(&count).Error
	return count, err
}

// CountGroupByPackage 按套餐统计成交量。
func (r *RegistrationRepository) CountGroupByPackage() ([]model.NameCount, error) {
	var rows []model.NameCount
	err := r.db.Model(&model.Registration{}).
		Select("packages.name as name, count(registrations.id) as count").
		Joins("JOIN packages ON packages.id = registrations.package_id").
		Group("packages.name").Scan(&rows).Error
	return rows, err
}

// RevenueByMonth 月度收入。
func (r *RegistrationRepository) RevenueByMonth() ([]model.MonthAmount, error) {
	var rows []model.MonthAmount
	err := r.db.Model(&model.Registration{}).
		Select("to_char(registered_at, 'YYYY-MM') as month, COALESCE(sum(packages.price),0) as amount").
		Joins("JOIN packages ON packages.id = registrations.package_id").
		Group("to_char(registered_at, 'YYYY-MM')").Order("month desc").Limit(6).Scan(&rows).Error
	return rows, err
}
