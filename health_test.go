package xorm

import (
	"context"
	"testing"

	"github.com/xudefa/go-boot/actuator"
	"github.com/xudefa/go-boot/health"

	_ "github.com/mattn/go-sqlite3"
)

// TestXormHealthIndicator 测试 XORM 数据库健康指标
func TestXormHealthIndicator(t *testing.T) {
	t.Run("successful check returns up", func(t *testing.T) {
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
	})

	t.Run("nil check function returns unknown", func(t *testing.T) {
		indicator := actuator.NewDatabaseHealthIndicator(nil)
		h := indicator.Health(context.Background())
		if h.Status != health.StatusUnknown {
			t.Fatalf("expected UNKNOWN, got %s", h.Status)
		}
	})

	t.Run("failed check returns down", func(t *testing.T) {
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
	})
}

// TestXormHealthIndicator_Name 测试健康指标名称
func TestXormHealthIndicator_Name(t *testing.T) {
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
