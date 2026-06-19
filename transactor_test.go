package xorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xudefa/go-boot/data"
)

func TestDB_ImplementsTransactor(t *testing.T) {
	db, err := OpenSQLite(WithDBName(":memory:"))
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	var _ data.Transactor = db
}

func TestDB_Query(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	rows, err := db.Query(ctx, "SELECT 1")
	assert.NoError(t, err)
	defer func() { _ = rows.Close() }()
	assert.NotNil(t, rows)
}

func TestDB_QueryRow(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	row := db.QueryRow(ctx, "SELECT 1")
	var val int
	err = row.Scan(&val)
	assert.NoError(t, err)
	assert.Equal(t, 1, val)
}

func TestDB_Exec(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	result, err := db.Exec(ctx, "CREATE TABLE test (id INTEGER)")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDB_Begin(t *testing.T) {
	ctx := context.Background()
	db, err := OpenSQLite(WithDBName(":memory:"))
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	tx, err := db.Begin(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	err = tx.Rollback()
	assert.NoError(t, err)
}

func TestDB_Stats(t *testing.T) {
	db, err := OpenSQLite(WithDBName(":memory:"))
	assert.NoError(t, err)
	defer func() { _ = db.Close() }()

	stats := db.Stats()
	assert.NotNil(t, stats)
}

func TestDB_Close(t *testing.T) {
	db, err := OpenSQLite(WithDBName(":memory:"))
	assert.NoError(t, err)

	err = db.Close()
	assert.NoError(t, err)
}
