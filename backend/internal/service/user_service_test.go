package service

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/blueship581/gbcheckup/internal/model"
	"github.com/blueship581/gbcheckup/internal/repository"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Package{}, &model.PackageItem{}, &model.Examinee{},
		&model.Registration{}, &model.ExamResult{}, &model.Report{}, &model.AbnormalMetric{},
		&model.Enterprise{}, &model.GroupOrder{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestUserService_RegisterAndLogin(t *testing.T) {
	db := newTestDB(t)
	svc := NewUserService(repository.NewUserRepository(db), "secret", 24, testLogger())
	ctx := context.Background()
	user, token, err := svc.Register(ctx, "13700000001", "pass123", "测试用户")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if user.Role != "examinee" || token == "" {
		t.Fatalf("register result mismatch: role=%s", user.Role)
	}
	if _, _, err := svc.Register(ctx, "13700000001", "pass456", "重复"); err == nil {
		t.Fatal("expected conflict for duplicate phone")
	}
	got, token2, err := svc.Login(ctx, "13700000001", "pass123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if got.ID != user.ID || token2 == "" {
		t.Fatal("login mismatch")
	}
	if _, _, err := svc.Login(ctx, "13700000001", "wrong"); err == nil {
		t.Fatal("expected login failure")
	}
}
