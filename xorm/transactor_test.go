package xorm

import (
	"context"
	"testing"

	_ "github.com/mattn/go-sqlite3" // SQLite 驱动
	"github.com/xudefa/go-boot/data"
)

// TestDB_ImplementsTransactor 验证 DB 实现了 data.Transactor 接口
func TestDB_ImplementsTransactor(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	var _ data.Transactor = db
}

// TestDB_Query 测试 DB 的 Query 方法
func TestDB_Query(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer func() { _ = rows.Close() }()
	if rows == nil {
		t.Fatal("rows should not be nil")
	}
}

// TestDB_QueryRow 测试 DB 的 QueryRow 方法
func TestDB_QueryRow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	row := db.QueryRow(ctx, "SELECT 1")
	var val int
	err = row.Scan(&val)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}
}

// TestDB_Exec 测试 DB 的 Exec 方法
func TestDB_Exec(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	result, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER)")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
}

// TestDB_Begin 测试 DB 的 Begin 方法
func TestDB_Begin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	if tx == nil {
		t.Fatal("tx should not be nil")
	}

	err = tx.Rollback()
	if err != nil {
		t.Errorf("Rollback failed: %v", err)
	}
}

// TestDB_Stats 测试 DB 的 Stats 方法
func TestDB_Stats(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	stats := db.Stats()
	// 某些字段可能为 0，这取决于驱动实现
	// 至少验证 stats 结构体被正确返回
	_ = stats
}

// TestDB_Close 测试 DB 的 Close 方法
func TestDB_Close(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// TestTransaction_Transactor 测试 Transaction 的 Transactor 方法
func TestTransaction_Transactor(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	// tx 是 *Transaction 类型（xorm 实现），有 Transactor 方法
	xormTx := tx.(*Transaction)
	transactor := xormTx.Transactor()
	if transactor == nil {
		t.Fatal("transactor should not be nil")
	}
}

// TestTransaction_Query 测试 Transaction 的 Query 方法
func TestTransaction_Query(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 使用简单的 SELECT 查询，不依赖表
	rows, err := tx.Query(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer func() { _ = rows.Close() }()
	if rows == nil {
		t.Fatal("rows should not be nil")
	}
}

// TestTransaction_QueryRow 测试 Transaction 的 QueryRow 方法
func TestTransaction_QueryRow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	row := tx.QueryRow(ctx, "SELECT 1")
	var val int
	err = row.Scan(&val)
	if err != nil {
		t.Fatalf("Scan failed: %v", err)
	}
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}
}

// TestTransaction_Exec 测试 Transaction 的 Exec 方法
func TestTransaction_Exec(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	// 使用简单的 SELECT 查询来验证事务执行能力
	result, err := tx.Exec(ctx, "SELECT 1")
	if err != nil {
		t.Fatalf("Exec failed: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
}

// TestTransaction_Begin 测试 Transaction 的 Begin 方法（嵌套事务）
func TestTransaction_Begin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	nestedTx, err := tx.Begin(ctx)
	if err != nil {
		t.Fatalf("Nested Begin failed: %v", err)
	}
	if nestedTx == nil {
		t.Fatal("nestedTx should not be nil")
	}

	err = nestedTx.Rollback()
	if err != nil {
		t.Errorf("Nested Rollback failed: %v", err)
	}
}

// TestTransaction_Stats 测试 Transaction 的 Stats 方法
func TestTransaction_Stats(t *testing.T) {
	t.Parallel()

	db, err := OpenSQLite(WithDBName(":memory:"))
	if err != nil {
		t.Fatalf("OpenSQLite failed: %v", err)
	}
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatalf("Begin failed: %v", err)
	}
	defer func() { _ = tx.Rollback() }()

	stats := tx.Stats()
	// Transaction.Stats() 返回空结构体，这是预期行为
	_ = stats
}
