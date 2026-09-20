package service

import (
	"context"
	"testing"

	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
)

func TestRegistrationService_Register(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	examinee := model.Examinee{Name: "张三", IDCardNo: "110101199001011234", Gender: "男", Age: 35}
	if err := db.Create(&examinee).Error; err != nil {
		t.Fatal(err)
	}
	pkg := model.Package{Name: "入职体检", PackageType: "basic", Price: 300, Status: "active"}
	if err := db.Create(&pkg).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create([]model.PackageItem{
		{PackageID: pkg.ID, ItemName: "血常规", RefValueRange: "3.5-9.5", Department: "检验科"},
		{PackageID: pkg.ID, ItemName: "肝功能", RefValueRange: "0-40", Department: "检验科"},
	}).Error; err != nil {
		t.Fatal(err)
	}

	svc := NewRegistrationService(
		repository.NewRegistrationRepository(db), repository.NewExamineeRepository(db),
		repository.NewPackageRepository(db), repository.NewPackageItemRepository(db),
		repository.NewExamResultRepository(db), testLogger(),
	)
	reg, err := svc.Register(ctx, examinee.ID, pkg.ID, 1)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if reg.ID == 0 || reg.GuideNo == "" {
		t.Fatalf("registration invalid: %+v", reg)
	}
	results, err := repository.NewExamResultRepository(db).ListByRegistration(reg.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("exam results = %d, want 2", len(results))
	}
	for _, r := range results {
		if r.Status != "pending" || r.RegistrationID != reg.ID {
			t.Fatalf("result invalid: %+v", r)
		}
	}
}
