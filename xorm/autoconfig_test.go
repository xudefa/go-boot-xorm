package xorm

import (
	"testing"

	_ "github.com/go-sql-driver/mysql" // MySQL 驱动
	_ "github.com/mattn/go-sqlite3"    // SQLite 驱动
	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/data"
	"github.com/xudefa/go-boot/environment"
	"github.com/xudefa/go-boot/event"
)

// TestXormAutoConfiguration_Configure_WithDefaults 测试使用默认配置
func TestXormAutoConfiguration_Configure_WithDefaults(t *testing.T) {
	t.Parallel()

	container := core.New()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
		constants.XORMEnabled: "true",
	}))

	ctx := &mockApplicationContext{
		container: container,
		env:       env,
	}

	config := &XormAutoConfiguration{}
	err := config.Configure(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !container.Has(constants.XORMDBBeanID) {
		t.Fatal("expected xormDB bean to be registered")
	}

	bean, err := container.Get(constants.XORMDBBeanID)
	if err != nil {
		t.Fatalf("failed to get bean: %v", err)
	}

	_, ok := bean.(*DB)
	if !ok {
		t.Fatalf("expected *DB, got %T", bean)
	}

	if !container.Has(constants.XORMDatabaseHealthIndicatorBeanID) {
		t.Fatal("expected xormDatabaseHealthIndicator bean to be registered")
	}
}

// TestXormAutoConfiguration_Configure_WithCustomConfig 测试自定义配置
func TestXormAutoConfiguration_Configure_WithCustomConfig(t *testing.T) {
	t.Parallel()

	container := core.New()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test", environment.PriorityNormal, map[string]any{
		constants.XORMEnabled:      "true",
		constants.XORMType:         "sqlite",
		constants.XORMHost:         "localhost",
		constants.XORMPort:         3306,
		constants.XORMUsername:     "root",
		constants.XORMPassword:     "123456",
		constants.XORMDatabase:     "testdb",
		constants.XORMMaxOpenConns: 50,
		constants.XORMMaxIdleConns: 5,
		constants.XORMShowSQL:      false,
	}))

	ctx := &mockApplicationContext{
		container: container,
		env:       env,
	}

	config := &XormAutoConfiguration{}
	err := config.Configure(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bean, err := container.Get(constants.XORMDBBeanID)
	if err != nil {
		t.Fatalf("failed to get bean: %v", err)
	}

	db := bean.(*DB)
	if db == nil {
		t.Fatal("expected non-nil DB")
	}
}

// TestXormAutoConfiguration_ImplementsAutoConfiguration 验证实现了 boot.AutoConfiguration 接口
func TestXormAutoConfiguration_ImplementsAutoConfiguration(t *testing.T) {
	t.Parallel()

	var _ boot.AutoConfiguration = (*XormAutoConfiguration)(nil)
}

// TestXormAutoConfiguration_DBImplementsTransactor 验证 DB 实现了 data.Transactor 接口
func TestXormAutoConfiguration_DBImplementsTransactor(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	var _ data.Transactor = db
}

// mockApplicationContext 模拟 ApplicationContext
type mockApplicationContext struct {
	container core.Container
	env       *environment.Environment
}

func (m *mockApplicationContext) Container() core.Container {
	return m.container
}

func (m *mockApplicationContext) Environment() *environment.Environment {
	return m.env
}

func (m *mockApplicationContext) Register(name string, opts ...core.BuilderOption) error {
	return m.container.Register(name, opts...)
}

func (m *mockApplicationContext) Get(name string) (any, error) {
	return m.container.Get(name)
}

func (m *mockApplicationContext) EventBus() interface {
	Publish(event event.ApplicationEvent)
} {
	return &mockEventBus{}
}

// mockEventBus 模拟事件总线
type mockEventBus struct{}

func (m *mockEventBus) Publish(event event.ApplicationEvent) {
	// 空实现，仅用于测试
}

// 验证接口实现
var _ boot.ApplicationContext = (*mockApplicationContext)(nil)
