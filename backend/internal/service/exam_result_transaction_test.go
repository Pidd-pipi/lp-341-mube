package service

import (
	"context"
	"testing"

	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
)

func TestExamResultService_EnterAbnormalCreatesMetric(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	item := model.PackageItem{ItemName: "血糖", RefValueRange: "3.9-6.1", Department: "检验科"}
	if err := db.Create(&item).Error; err != nil {
		t.Fatal(err)
	}
	examinee := model.Examinee{Name: "张三", IDCardNo: "110101199001011234"}
	if err := db.Create(&examinee).Error; err != nil {
		t.Fatal(err)
	}
	res := model.ExamResult{ExamineeID: examinee.ID, PackageItemID: item.ID, Status: "pending", DoctorID: 1}
	if err := db.Create(&res).Error; err != nil {
		t.Fatal(err)
	}

	resultRepo := repository.NewExamResultRepository(db)
	metricRepo := repository.NewAbnormalMetricRepository(db)
	svc := NewExamResultService(resultRepo, repository.NewRegistrationRepository(db), metricRepo, testLogger())

	entered, err := svc.Enter(ctx, res.ID, 1, EnterInput{ResultValue: "9.9"})
	if err != nil {
		t.Fatalf("Enter() error = %v", err)
	}
	if !entered.IsAbnormal {
		t.Fatal("expected abnormal result")
	}
	metrics, _, err := metricRepo.List(examinee.ID, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(metrics) != 1 {
		t.Fatalf("abnormal metrics = %d, want 1", len(metrics))
	}
	if metrics[0].Value != "9.9" {
		t.Fatalf("metric value = %s, want 9.9", metrics[0].Value)
	}
}
