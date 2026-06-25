// xorm 集成模块测试
// 测试 xorm 数据库连接、事务、Repository CRUD 操作和 DSN 生成等功能
package xorm

import (
	"context"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite 驱动
	"github.com/xudefa/go-boot/data"
)

// TestUser 测试用用户结构体
type TestUser struct {
	ID   uint64 `xorm:"id pk autoincr"`
	Name string `xorm:"name"`
}

// TestOpenSQLite 测试使用 SQLite 内存数据库打开连接，验证 db 对象和 engine 不为空
func TestOpenSQLite(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	if db == nil {
		t.Fatal("OpenSQLite() returned nil")
	}
	if db.engine == nil {
		t.Error("engine should not be nil")
	}
}

// TestOpenWithOptions 测试使用 Open 通用接口通过选项创建数据库连接，验证连接成功
func TestOpenWithOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts []Option
	}{
		{
			name: "sqlite with memory database",
			opts: []Option{
				WithDBType(SQLite),
				WithDBName(":memory:"),
			},
		},
		{
			name: "sqlite with custom DSN",
			opts: []Option{
				WithDBType(SQLite),
				WithDSN(":memory:"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			db, err := Open(tt.opts...)
			if err != nil {
				t.Fatalf("Open failed: %v", err)
			}
			if db == nil {
				t.Fatal("Open() returned nil")
			}
		})
	}
}

// TestTransaction_Commit 测试提交事务，验证提交无错误
func TestTransaction_Commit(t *testing.T) {
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	err = tx.Commit()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestTransaction_Rollback 测试回滚事务，验证回滚无错误
func TestTransaction_Rollback(t *testing.T) {
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	err = tx.Rollback()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestTransaction_Close 测试关闭事务，验证关闭无错误
func TestTransaction_Close(t *testing.T) {
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	err = tx.Close()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

// TestTransaction_ImplementsTransactionInterface 编译时检查 Transaction 是否实现了 data.Transaction 接口
func TestTransaction_ImplementsTransactionInterface(t *testing.T) {
	t.Parallel()

	var _ data.Transaction = (*Transaction)(nil)
}

// TestRepository_Create 测试 Repository 的 Create 方法，验证创建记录后 ID 被自动填充
func TestRepository_Create(t *testing.T) {
	db := createTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewRepository[TestUser](db.Engine())
	user := &TestUser{Name: "John"}

	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if user.ID == 0 {
		t.Error("user ID should be set after create")
	}
}

// TestRepository_FindByID 测试 Repository 的 FindByID 方法，验证根据 ID 查找记录并返回正确字段
func TestRepository_FindByID(t *testing.T) {
	db := createTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewRepository[TestUser](db.Engine())
	user := &TestUser{Name: "John"}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := repo.FindByID(user.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found == nil {
		t.Fatal("found should not be nil")
	}
	if found.Name != "John" {
		t.Errorf("expected name 'John', got %s", found.Name)
	}
}

// TestRepository_Update 测试 Repository 的 Update 方法，验证更新记录后重新查询得到新值
func TestRepository_Update(t *testing.T) {
	db := createTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewRepository[TestUser](db.Engine())
	user := &TestUser{Name: "John"}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	user.Name = "Jane"
	err := repo.Update(user)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByID(user.ID)
	if found.Name != "Jane" {
		t.Errorf("expected name 'Jane', got %s", found.Name)
	}
}

// TestRepository_Delete 测试 Repository 的 Delete 方法
func TestRepository_Delete(t *testing.T) {
	db := createTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewRepository[TestUser](db.Engine())
	user := &TestUser{Name: "John"}
	if err := repo.Create(user); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err := repo.Delete(user.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	found, _ := repo.FindByID(user.ID)
	if found != nil {
		t.Error("user should be deleted")
	}
}

// TestRepository_FindAll 测试 Repository 的 FindAll 方法，验证插入两条记录后查询到全部结果
func TestRepository_FindAll(t *testing.T) {
	db := createTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewRepository[TestUser](db.Engine())
	if err := repo.Create(&TestUser{Name: "John"}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := repo.Create(&TestUser{Name: "Jane"}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	results, err := repo.FindAll(nil)
	if err != nil {
		t.Fatalf("FindAll failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 users, got %d", len(results))
	}
}

// TestRepository_Count 测试 Repository 的 Count 方法，验证插入两条记录后计数为 2
func TestRepository_Count(t *testing.T) {
	db := createTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewRepository[TestUser](db.Engine())
	if err := repo.Create(&TestUser{Name: "John"}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := repo.Create(&TestUser{Name: "Jane"}); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	count, err := repo.Count(nil)
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected count 2, got %d", count)
	}
}

// TestRepository_ImplementsRepositoryInterface 编译时检查 Repository[TestUser] 是否实现了 data.Repository[TestUser] 接口
func TestRepository_ImplementsRepositoryInterface(t *testing.T) {
	t.Parallel()

	var _ data.Repository[TestUser] = (*Repository[TestUser])(nil)
}

// TestClient 测试创建 xorm Client 包装对象，验证不为空
func TestClient(t *testing.T) {
	t.Parallel()

	db, _ := OpenSQLite(WithDBName(":memory:"))
	defer func() { _ = db.Close() }()

	client := NewClient(db.Engine())
	if client == nil {
		t.Error("NewClient should return client")
	}
}

// TestConfig_DSNForMySQL 测试生成 MySQL 的 DSN 连接串，验证不为空且长度合理
func TestConfig_DSNForMySQL(t *testing.T) {
	t.Parallel()

	cfg := &config{
		Host:     "localhost",
		Port:     3306,
		User:     "gate",
		Password: "123456",
		DBName:   "gate",
		Charset:  "utf8",
	}
	dsn := cfg.DSNForMySQL()
	if dsn == "" {
		t.Error("DSN should not be empty")
	}
	if len(dsn) < 20 {
		t.Errorf("DSN seems too short: %s", dsn)
	}
}

// TestConfig_DSNForPostgres 测试生成 PostgreSQL 的 DSN 连接串，验证不为空
func TestConfig_DSNForPostgres(t *testing.T) {
	t.Parallel()

	cfg := &config{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "testdb",
		SslMode:  "disable",
	}
	dsn := cfg.DSNForPostgres()
	if dsn == "" {
		t.Error("DSN should not be empty")
	}
}

// TestConfig_DSNForSQLite 测试生成 SQLite 的 DSN 连接串，验证直接返回数据库文件名
func TestConfig_DSNForSQLite(t *testing.T) {
	t.Parallel()

	cfg := &config{
		DBName: "test.db",
	}
	dsn := cfg.DSNForSQLite()
	if dsn != "test.db" {
		t.Errorf("expected 'test.db', got '%s'", dsn)
	}
}

// TestConfig_DSNForMSSQL 测试生成 MSSQL 的 DSN 连接串，验证不为空
func TestConfig_DSNForMSSQL(t *testing.T) {
	t.Parallel()

	cfg := &config{
		Host:     "localhost",
		Port:     1433,
		User:     "sa",
		Password: "password",
		DBName:   "testdb",
	}
	dsn := cfg.DSNForMSSQL()
	if dsn == "" {
		t.Error("DSN should not be empty")
	}
}

// TestConfig_DSNOverride 测试自定义 DSN 覆盖
func TestConfig_DSNOverride(t *testing.T) {
	t.Parallel()

	cfg := &config{
		DSN: "custom-dsn-string",
	}

	if cfg.DSNForMySQL() != "custom-dsn-string" {
		t.Error("DSN should be overridden")
	}
	if cfg.DSNForPostgres() != "custom-dsn-string" {
		t.Error("DSN should be overridden")
	}
	if cfg.DSNForMSSQL() != "custom-dsn-string" {
		t.Error("DSN should be overridden")
	}
	if cfg.DSNForSQLite() != "custom-dsn-string" {
		t.Error("DSN should be overridden")
	}
}

// createTestDB 创建测试数据库并返回 DB 实例
func createTestDB(t *testing.T) *DB {
	// 清理旧数据库文件
	_ = os.Remove("/tmp/test_xorm.db")

	db, err := OpenSQLite(WithDBName("/tmp/test_xorm.db"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}

	// 使用原生 SQL 创建表
	_, err = db.Engine().Exec(`CREATE TABLE test_user (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("Create table failed: %v", err)
	}

	return db
}

// cleanupTestDB 清理测试数据库
func cleanupTestDB(t *testing.T, db *DB) {
	_ = db.Close()
	_ = os.Remove("/tmp/test_xorm.db")
}
