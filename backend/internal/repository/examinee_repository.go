package repository

import (
	"errors"

	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

// ExamineeRepository 体检人仓储。
type ExamineeRepository struct{ db *gorm.DB }

// NewExamineeRepository 构造体检人仓储。
func NewExamineeRepository(db *gorm.DB) *ExamineeRepository { return &ExamineeRepository{db: db} }

func (r *ExamineeRepository) Create(e *model.Examinee) error { return r.db.Create(e).Error }

func (r *ExamineeRepository) CreateBatch(items []model.Examinee) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

func (r *ExamineeRepository) FindByID(id uint) (*model.Examinee, error) {
	var e model.Examinee
	if err := r.db.First(&e, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *ExamineeRepository) FindByIDCard(idCard string) (*model.Examinee, error) {
	var e model.Examinee
	if err := r.db.Where("id_card_no = ?", idCard).First(&e).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, util.ErrNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *ExamineeRepository) List(keyword string, page, pageSize int) ([]model.Examinee, int64, error) {
	q := r.db.Model(&model.Examinee{})
	if keyword != "" {
		q = q.Where("name ILIKE ? OR id_card_no ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Examinee
	err := r.db.Where("name ILIKE ? OR id_card_no ILIKE ?", "%"+keyword+"%", "%"+keyword+"%").Order("id desc").Offset((page-1)*pageSize).Limit(pageSize).Find(&items).Error
	return items, total, err
}

func (r *ExamineeRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Examinee{}).Count(&count).Error
	return count, err
}
