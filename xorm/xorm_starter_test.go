package xorm

import (
	"testing"

	src "github.com/xudefa/go-boot"
)

// TestStarter 测试用模型结构体
type TestStarterModel struct {
	ID   uint64 `xorm:"id pk autoincr"`
	Name string `xorm:"name"`
}

// TestStarter_Starter_WithNilEngine 测试 engine 为 nil 时 Starter 返回 nil
func TestStarter_Starter_WithNilEngine(t *testing.T) {
	t.Parallel()

	starter := &Starter{
		engine:      nil,
		autoMigrate: nil,
	}

	err := starter.Starter()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// TestStarter_Starter_WithEmptyModels 测试空模型列表时 Starter 返回 nil
func TestStarter_Starter_WithEmptyModels(t *testing.T) {
	t.Parallel()

	// 创建 mock engine
	engine := &mockSyncEngine{sync2Called: false}

	starter := &Starter{
		engine:      engine,
		autoMigrate: []any{},
	}

	err := starter.Starter()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if engine.sync2Called {
		t.Error("Sync2 should not be called with empty models")
	}
}

// TestStarter_Starter_WithModels 测试有模型时调用 Sync2
func TestStarter_Starter_WithModels(t *testing.T) {
	t.Parallel()

	engine := &mockSyncEngine{sync2Called: false}
	models := []any{&TestStarterModel{}}

	starter := &Starter{
		engine:      engine,
		autoMigrate: models,
	}

	err := starter.Starter()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	if !engine.sync2Called {
		t.Error("expected Sync2 to be called")
	}
}

// TestNewStarter 测试创建启动器
func TestNewStarter(t *testing.T) {
	t.Parallel()

	engine := &mockSyncEngine{}
	models := []any{&TestStarterModel{}}

	starter := NewStarter(engine, models)
	if starter == nil {
		t.Fatal("expected non-nil starter")
	}

	if len(starter.autoMigrate) != 1 {
		t.Errorf("expected 1 model, got %d", len(starter.autoMigrate))
	}
}

// TestNewAutoMigrateStarter 测试创建自动迁移启动器
func TestNewAutoMigrateStarter(t *testing.T) {
	t.Parallel()

	engine := &mockSyncEngine{}
	models := []any{&TestStarterModel{}, &TestStarterModel{}}

	starter := NewAutoMigrateStarter(engine, models...)
	if starter == nil {
		t.Fatal("expected non-nil starter")
	}

	// 验证返回类型
	var _ src.Starter = starter
}

// TestStarter_ImplementsStarterInterface 验证实现了 boot.Starter 接口
func TestStarter_ImplementsStarterInterface(t *testing.T) {
	t.Parallel()

	var _ src.Starter = (*Starter)(nil)
}

// mockSyncEngine 模拟支持 Sync2 的引擎
type mockSyncEngine struct {
	sync2Called bool
}

func (m *mockSyncEngine) Sync2(beans ...any) error {
	m.sync2Called = true
	return nil
}
