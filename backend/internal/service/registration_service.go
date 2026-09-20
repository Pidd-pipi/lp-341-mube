package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
	"github.com/blueship581/gbcheckup/internal/util"
	"gorm.io/gorm"
)

// RegistrationService 体检登记/导检服务。
type RegistrationService struct {
	repo       *repository.RegistrationRepository
	examinee   *repository.ExamineeRepository
	pkg        *repository.PackageRepository
	itemRepo   *repository.PackageItemRepository
	resultRepo *repository.ExamResultRepository
	log        *slog.Logger
}

// NewRegistrationService 构造登记服务。
func NewRegistrationService(repo *repository.RegistrationRepository, examinee *repository.ExamineeRepository, pkg *repository.PackageRepository, itemRepo *repository.PackageItemRepository, resultRepo *repository.ExamResultRepository, log *slog.Logger) *RegistrationService {
	return &RegistrationService{repo: repo, examinee: examinee, pkg: pkg, itemRepo: itemRepo, resultRepo: resultRepo, log: log}
}

// Register 登记体检人并分配套餐，生成导检单号与待录入结果。
func (s *RegistrationService) Register(ctx context.Context, examineeID, packageID, registerUserID uint) (*model.Registration, error) {
	if _, err := s.examinee.FindByID(examineeID); err != nil {
		return nil, util.NotFoundError(constants.MsgExamineeNotFound, err)
	}
	pkg, err := s.pkg.FindByID(packageID)
	if err != nil {
		return nil, util.NotFoundError(constants.MsgPackageNotFound, err)
	}
	seq, _ := s.repo.Count()
	reg := &model.Registration{
		ExamineeID: examineeID, PackageID: packageID,
		GuideNo:       fmt.Sprintf("GUIDE%s%04d", time.Now().Format("20060102"), seq+1),
		Status:        constants.RegistrationRegistered,
		RegisteredAt:  time.Now(), RegisterUserID: registerUserID,
	}
	// 初始化待录入检查结果
	items, err := s.itemRepo.ListByPackage(packageID)
	if err != nil {
		return nil, err
	}
	results := make([]model.ExamResult, 0, len(items))
	for _, it := range items {
		results = append(results, model.ExamResult{
			ExamineeID: examineeID, PackageItemID: it.ID,
			Status: constants.ResultPending, DoctorID: registerUserID,
		})
	}
	err = s.repo.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.WithTx(tx).Create(reg); err != nil {
			return fmt.Errorf("create registration: %w", err)
		}
		for i := range results {
			results[i].RegistrationID = reg.ID
		}
		if err := s.resultRepo.WithTx(tx).CreateBatch(results); err != nil {
			return fmt.Errorf("init exam results: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, util.LogError(s.log, constants.LOG_REGISTRATION_CREATED, err)
	}
	s.log.InfoContext(ctx, constants.LOG_REGISTRATION_CREATED, "registration_id", reg.ID, "package", pkg.Name)
	return reg, nil
}

// List 分页查询登记。
func (s *RegistrationService) List(ctx context.Context, status string, page, pageSize int) ([]model.Registration, int64, error) {
	return s.repo.List(status, page, pageSize)
}

// Get 查询登记详情。
func (s *RegistrationService) Get(ctx context.Context, id uint) (*model.Registration, error) {
	reg, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError(constants.MsgRegNotFound, err)
		}
		return nil, err
	}
	return reg, nil
}

// UpdateStatus 更新登记状态。
func (s *RegistrationService) UpdateStatus(ctx context.Context, id uint, status string) error {
	valid := map[string]bool{constants.RegistrationRegistered: true, constants.RegistrationInProgress: true, constants.RegistrationCompleted: true}
	if !valid[status] {
		return util.BadRequest("登记状态（Registration.status）不合法", errors.New("invalid status"))
	}
	if err := s.repo.UpdateStatus(id, status); err != nil {
		return util.LogError(s.log, constants.LOG_REGISTRATION_STATUS_CHANGED, fmt.Errorf("update registration status: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_REGISTRATION_STATUS_CHANGED, "registration_id", id, "status", status)
	return nil
}
