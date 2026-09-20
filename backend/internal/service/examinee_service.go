package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/blueship581/gbcheckup/internal/constants"
	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
	"github.com/blueship581/gbcheckup/internal/util"
)

// ExamineeService 体检人服务。
type ExamineeService struct {
	repo *repository.ExamineeRepository
	log  *slog.Logger
}

// NewExamineeService 构造体检人服务。
func NewExamineeService(repo *repository.ExamineeRepository, log *slog.Logger) *ExamineeService {
	return &ExamineeService{repo: repo, log: log}
}

// Create 登记体检人。
func (s *ExamineeService) Create(ctx context.Context, e *model.Examinee) (*model.Examinee, error) {
	if existing, err := s.repo.FindByIDCard(e.IDCardNo); err == nil && existing != nil {
		return nil, util.ConflictError("身份证号（Examinee.id_card_no）已存在", errors.New("duplicate id card"))
	}
	if err := s.repo.Create(e); err != nil {
		return nil, util.LogError(s.log, constants.LOG_EXAMINEE_CREATED, fmt.Errorf("create examinee: %w", err))
	}
	s.log.InfoContext(ctx, constants.LOG_EXAMINEE_CREATED, "examinee_id", e.ID)
	return e, nil
}

// Get 查询体检人。
func (s *ExamineeService) Get(ctx context.Context, id uint) (*model.Examinee, error) {
	e, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, util.ErrNotFound) {
			return nil, util.NotFoundError(constants.MsgExamineeNotFound, err)
		}
		return nil, err
	}
	return e, nil
}

// List 分页查询体检人。
func (s *ExamineeService) List(ctx context.Context, keyword string, page, pageSize int) ([]model.Examinee, int64, error) {
	return s.repo.List(keyword, page, pageSize)
}

// BatchImport 团体批量导入（每行：姓名,身份证号,手机号,性别,年龄）。
func (s *ExamineeService) BatchImport(ctx context.Context, enterpriseID *uint, csvText string) (int, []model.Examinee, error) {
	lines := strings.Split(strings.TrimSpace(csvText), "\n")
	created := make([]model.Examinee, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "姓名") {
			continue
		}
		cols := strings.Split(line, ",")
		if len(cols) < 2 {
			continue
		}
		name := strings.TrimSpace(cols[0])
		idCard := strings.TrimSpace(cols[1])
		if name == "" || idCard == "" {
			continue
		}
		e := &model.Examinee{Name: name, IDCardNo: idCard, SourceType: "group", EnterpriseID: enterpriseID}
		if len(cols) > 2 {
			e.Phone = strings.TrimSpace(cols[2])
		}
		if len(cols) > 3 {
			e.Gender = strings.TrimSpace(cols[3])
		}
		if len(cols) > 4 {
			if age, err := strconv.Atoi(strings.TrimSpace(cols[4])); err == nil {
				e.Age = age
			}
		}
		if existing, err := s.repo.FindByIDCard(idCard); err == nil && existing != nil {
			continue
		}
		if err := s.repo.Create(e); err != nil {
			return 0, nil, util.LogError(s.log, constants.LOG_EXAMINEE_IMPORTED, fmt.Errorf("import examinee: %w", err))
		}
		created = append(created, *e)
	}
	s.log.InfoContext(ctx, constants.LOG_EXAMINEE_IMPORTED, "imported", len(created))
	return len(created), created, nil
}
