package xorm

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite 驱动
	"github.com/xudefa/go-boot/actuator"
	"github.com/xudefa/go-boot/health"
)

// TestXormHealthIndicator_Successful 测试健康检查成功时返回 UP 状态
func TestXormHealthIndicator_Successful(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	indicator := actuator.NewDatabaseHealthIndicator(func(ctx context.Context) error {
		sqlDB := db.Engine().DB()
		if sqlDB == nil {
			return nil
		}
		return sqlDB.PingContext(ctx)
	})

	h := indicator.Health(context.Background())
	if h.Status != health.StatusUp {
		t.Fatalf("expected UP, got %s", h.Status)
	}
}

// TestXormHealthIndicator_NilCheck 测试检查函数为 nil 时返回 UNKNOWN 状态
func TestXormHealthIndicator_NilCheck(t *testing.T) {
	t.Parallel()

	indicator := actuator.NewDatabaseHealthIndicator(nil)
	h := indicator.Health(context.Background())
	if h.Status != health.StatusUnknown {
		t.Fatalf("expected UNKNOWN, got %s", h.Status)
	}
}

// TestXormHealthIndicator_Failed 测试检查失败时返回 DOWN 状态
func TestXormHealthIndicator_Failed(t *testing.T) {
	t.Parallel()

	indicator := actuator.NewDatabaseHealthIndicator(func(ctx context.Context) error {
		return &testError{msg: "connection failed"}
	})
	h := indicator.Health(context.Background())
	if h.Status != health.StatusDown {
		t.Fatalf("expected DOWN, got %s", h.Status)
	}
	if h.Details == nil {
		t.Fatal("expected details")
	}
}

// TestXormHealthIndicator_Name 测试健康指标名称
func TestXormHealthIndicator_Name(t *testing.T) {
	t.Parallel()

	indicator := actuator.NewDatabaseHealthIndicator(nil)
	if indicator.Name() != "database" {
		t.Fatalf("expected 'database', got %s", indicator.Name())
	}
}

// testError 测试用的错误类型
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
