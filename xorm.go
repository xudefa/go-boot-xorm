// Package xorm 基于 XORM 提供数据库访问层实现。
//
// 该包将 XORM 与 go-boot 数据访问层接口集成，
// 支持多种数据库、事务和 Repository 模式。
//
// 定义：
//
//   - DB: 数据库连接实现了 data.Transactor 接口
//   - Transaction: 事务实现了 data.Transaction 接口
//   - Repository[T]: 泛型 Repository 实现了 data.Repository[T] 接口
//   - Option: 数据库配置选项
//
// 支持的数据库类型：
//
//   - MySQL: OpenMySQL()
//   - PostgreSQL: OpenPostgreSQL()
//   - MSSQL: OpenMSSQL()
//   - SQLite: OpenSQLite() (默认)
//
// 快速开始:
//
//	// 创建数据库连接
//	db, _ := xorm.OpenMySQL(
//	    xorm.WithHost("localhost"),
//	    xorm.WithPort(3306),
//	    xorm.WithUser("gate"),
//	    xorm.WithPassword("123456"),
//	    xorm.WithDBName("gate"),
//	)
//
//	// 创建 Repository
//	repo := xorm.NewRepository[User](db.Engine())
//
//	// CRUD 操作
//	user := &User{Name: "John"}
//	repo.Create(user)
package xorm

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/xudefa/go-boot/data"

	"xorm.io/xorm"
)

// DB 是 XORM 数据库连接。
//
// 字段说明:
//   - engine: XORM 引擎实例
type DB struct {
	engine *xorm.Engine
}

// Open 打开数据库连接。
//
// 参数:
//   - opts: 可选的配置选项
//
// 返回值:
//   - *DB: 数据库连接实例
//   - error: 连接错误
func Open(opts ...Option) (*DB, error) {
	cfg := &config{Type: string(SQLite), DBName: ":memory:"}
	for _, opt := range opts {
		opt(cfg)
	}

	var driver, dsn string
	switch cfg.Type {
	case string(MySQL):
		driver = "mysql"
		dsn = cfg.DSNForMySQL()
	case string(PostgreSQL):
		driver = "postgres"
		dsn = cfg.DSNForPostgres()
	case string(MSSQL):
		driver = "mssql"
		dsn = cfg.DSNForMSSQL()
	case string(SQLite):
		driver = "sqlite3"
		dsn = cfg.DSNForSQLite()
	default:
		driver = "mysql"
		dsn = cfg.DSNForMySQL()
	}

	engine, err := xorm.NewEngine(driver, dsn)
	if err != nil {
		return nil, err
	}

	engine.ShowSQL(cfg.ShowSQL)
	engine.SetMaxIdleConns(cfg.MaxIdleConns)
	engine.SetMaxOpenConns(cfg.MaxOpenConns)

	return &DB{engine: engine}, nil
}

type config struct {
	Type         string
	DSN          string
	Host         string
	Port         int
	User         string
	Password     string
	DBName       string
	MaxIdleConns int
	MaxOpenConns int
	ShowSQL      bool
	TimeZone     string
	SslMode      string
	Charset      string
}

// DSNForMySQL 生成 MySQL 数据源名称。
//
// 返回值:
//   - string: MySQL DSN 字符串
func (c *config) DSNForMySQL() string {
	if c.DSN != "" {
		return c.DSN
	}
	return c.User + ":" + c.Password + "@tcp(" + c.Host + ":" + itoa(c.Port) + ")/" + c.DBName + "?charset=" + c.Charset
}

// DSNForPostgres 生成 PostgreSQL 数据源名称。
//
// 返回值:
//   - string: PostgreSQL DSN 字符串
func (c *config) DSNForPostgres() string {
	if c.DSN != "" {
		return c.DSN
	}
	return "host=" + c.Host + " user=" + c.User + " password=" + c.Password + " dbname=" + c.DBName + " port=" + itoa(c.Port) + " sslmode=" + c.SslMode
}

// DSNForMSSQL 生成 MSSQL 数据源名称。
//
// 返回值:
//   - string: MSSQL DSN 字符串
func (c *config) DSNForMSSQL() string {
	if c.DSN != "" {
		return c.DSN
	}
	return "server=" + c.Host + ";port=" + itoa(c.Port) + ";database=" + c.DBName + ";user id=" + c.User + ";password=" + c.Password + ";encrypt=disable"
}

// DSNForSQLite 生成 SQLite 数据源名称。
//
// 返回值:
//   - string: SQLite DSN 字符串
func (c *config) DSNForSQLite() string {
	if c.DSN != "" {
		return c.DSN
	}
	return c.DBName
}

// DBType 定义数据库类型。
type DBType string

const (
	MySQL      DBType = "mysql"    // MySQL 数据库
	PostgreSQL DBType = "postgres" // PostgreSQL 数据库
	MSSQL      DBType = "mssql"    // MSSQL 数据库
	SQLite     DBType = "sqlite"   // SQLite 数据库
)

// Option 是数据库配置选项函数。
type Option func(*config)

// WithDSN 设置自定义数据源名称。
//
// 参数:
//   - dsn: 数据源名称字符串
//
// 返回值:
//   - Option: 配置选项函数
func WithDSN(dsn string) Option {
	return func(c *config) {
		c.DSN = dsn
	}
}

// WithDBType 设置数据库类型。
//
// 参数:
//   - t: 数据库类型
//
// 返回值:
//   - Option: 配置选项函数
func WithDBType(t DBType) Option {
	return func(c *config) {
		c.Type = string(t)
	}
}

// WithHost 设置数据库主机地址。
//
// 参数:
//   - host: 主机地址
//
// 返回值:
//   - Option: 配置选项函数
func WithHost(host string) Option {
	return func(c *config) {
		c.Host = host
	}
}

// WithPort 设置数据库端口号。
//
// 参数:
//   - port: 端口号
//
// 返回值:
//   - Option: 配置选项函数
func WithPort(port int) Option {
	return func(c *config) {
		c.Port = port
	}
}

// WithUser 设置数据库用户名。
//
// 参数:
//   - user: 用户名
//
// 返回值:
//   - Option: 配置选项函数
func WithUser(user string) Option {
	return func(c *config) {
		c.User = user
	}
}

// WithPassword 设置数据库密码。
//
// 参数:
//   - password: 密码
//
// 返回值:
//   - Option: 配置选项函数
func WithPassword(password string) Option {
	return func(c *config) {
		c.Password = password
	}
}

// WithDBName 设置数据库名称。
//
// 参数:
//   - dbname: 数据库名称
//
// 返回值:
//   - Option: 配置选项函数
func WithDBName(dbname string) Option {
	return func(c *config) {
		c.DBName = dbname
	}
}

// WithMaxIdleConns 设置最大空闲连接数。
//
// 参数:
//   - n: 空闲连接数
//
// 返回值:
//   - Option: 配置选项函数
func WithMaxIdleConns(n int) Option {
	return func(c *config) {
		c.MaxIdleConns = n
	}
}

// WithMaxOpenConns 设置最大打开连接数。
//
// 参数:
//   - n: 打开连接数
//
// 返回值:
//   - Option: 配置选项函数
func WithMaxOpenConns(n int) Option {
	return func(c *config) {
		c.MaxOpenConns = n
	}
}

// WithShowSQL 设置是否显示 SQL。
//
// 参数:
//   - show: 是否显示 SQL
//
// 返回值:
//   - Option: 配置选项函数
func WithShowSQL(show bool) Option {
	return func(c *config) {
		c.ShowSQL = show
	}
}

// WithSSLMode 设置 SSL 模式。
//
// 参数:
//   - mode: SSL 模式
//
// 返回值:
//   - Option: 配置选项函数
func WithSSLMode(mode string) Option {
	return func(c *config) {
		c.SslMode = mode
	}
}

// WithCharset 设置字符集。
//
// 参数:
//   - charset: 字符集
//
// 返回值:
//   - Option: 配置选项函数
func WithCharset(charset string) Option {
	return func(c *config) {
		c.Charset = charset
	}
}

// WithTimeZone 设置时区。
//
// 参数:
//   - tz: 时区字符串
//
// 返回值:
//   - Option: 配置选项函数
func WithTimeZone(tz string) Option {
	return func(c *config) {
		c.TimeZone = tz
	}
}

// OpenMySQL 打开 MySQL 数据库连接。
//
// 参数:
//   - opts: 可选的配置选项
//
// 返回值:
//   - *DB: 数据库连接实例
//   - error: 连接错误
func OpenMySQL(opts ...Option) (*DB, error) {
	cfg := &config{Type: string(MySQL)}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.DSN = cfg.DSNForMySQL()

	engine, err := xorm.NewEngine("mysql", cfg.DSN)
	if err != nil {
		return nil, err
	}
	engine.ShowSQL(cfg.ShowSQL)
	engine.SetMaxIdleConns(cfg.MaxIdleConns)
	engine.SetMaxOpenConns(cfg.MaxOpenConns)
	return &DB{engine: engine}, nil
}

// OpenPostgreSQL 打开 PostgreSQL 数据库连接。
//
// 参数:
//   - opts: 可选的配置选项
//
// 返回值:
//   - *DB: 数据库连接实例
//   - error: 连接错误
func OpenPostgreSQL(opts ...Option) (*DB, error) {
	cfg := &config{Type: string(PostgreSQL)}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.DSN = cfg.DSNForPostgres()

	engine, err := xorm.NewEngine("postgres", cfg.DSN)
	if err != nil {
		return nil, err
	}
	engine.ShowSQL(cfg.ShowSQL)
	engine.SetMaxIdleConns(cfg.MaxIdleConns)
	engine.SetMaxOpenConns(cfg.MaxOpenConns)
	return &DB{engine: engine}, nil
}

// OpenMSSQL 打开 MSSQL 数据库连接。
//
// 参数:
//   - opts: 可选的配置选项
//
// 返回值:
//   - *DB: 数据库连接实例
//   - error: 连接错误
func OpenMSSQL(opts ...Option) (*DB, error) {
	cfg := &config{Type: string(MSSQL)}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.DSN = cfg.DSNForMSSQL()

	engine, err := xorm.NewEngine("mssql", cfg.DSN)
	if err != nil {
		return nil, err
	}
	engine.ShowSQL(cfg.ShowSQL)
	engine.SetMaxIdleConns(cfg.MaxIdleConns)
	engine.SetMaxOpenConns(cfg.MaxOpenConns)
	return &DB{engine: engine}, nil
}

// OpenSQLite 打开 SQLite 数据库连接。
//
// 参数:
//   - opts: 可选的配置选项
//
// 返回值:
//   - *DB: 数据库连接实例
//   - error: 连接错误
func OpenSQLite(opts ...Option) (*DB, error) {
	cfg := &config{Type: string(SQLite)}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.DSN = cfg.DSNForSQLite()

	engine, err := xorm.NewEngine("sqlite3", cfg.DSN)
	if err != nil {
		return nil, err
	}
	engine.ShowSQL(cfg.ShowSQL)
	engine.SetMaxIdleConns(cfg.MaxIdleConns)
	engine.SetMaxOpenConns(cfg.MaxOpenConns)
	return &DB{engine: engine}, nil
}

// Engine 返回底层的 XORM 引擎实例。
//
// 返回值:
//   - *xorm.Engine: XORM 引擎实例
func (d *DB) Engine() *xorm.Engine {
	return d.engine
}

// Transactor 返回 data.Transactor 接口，用于事务操作。
//
// 返回值:
//   - data.Transactor: 事务执行器
func (d *DB) Transactor() data.Transactor {
	return d
}

// Query 执行查询并返回多行结果。
//
// 参数:
//   - ctx: 上下文
//   - query: SQL 查询语句
//   - args: 查询参数
//
// 返回值:
//   - data.Rows: 查询结果行
//   - error: 错误
func (d *DB) Query(ctx context.Context, query string, args ...any) (data.Rows, error) {
	session := d.engine.Context(ctx)
	rows, err := session.Query(append([]any{query}, convertArgs(args)...)...)
	if err != nil {
		return nil, err
	}
	return &xormRows{rows: rows}, nil
}

// QueryRow 执行查询并返回单行结果。
//
// 参数:
//   - ctx: 上下文
//   - query: SQL 查询语句
//   - args: 查询参数
//
// 返回值:
//   - data.Row: 单行查询结果
func (d *DB) QueryRow(ctx context.Context, query string, args ...any) data.Row {
	return &xormRow{session: d.engine.Context(ctx), query: query, args: args}
}

// Exec 执行 SQL 并返回结果。
//
// 参数:
//   - ctx: 上下文
//   - query: SQL 语句
//   - args: 参数
//
// 返回值:
//   - data.Result: 执行结果
//   - error: 错误
func (d *DB) Exec(ctx context.Context, query string, args ...any) (data.Result, error) {
	session := d.engine.Context(ctx)
	result, err := session.Exec(append([]any{query}, convertArgs(args)...)...)
	if err != nil {
		return nil, err
	}
	return &xormResult{result: result}, nil
}

// Begin 开始一个新事务。
//
// 参数:
//   - ctx: 上下文
//
// 返回值:
//   - data.Transaction: 事务实例
//   - error: 错误
func (d *DB) Begin(ctx context.Context) (data.Transaction, error) {
	session := d.engine.NewSession().Context(ctx)
	if err := session.Begin(); err != nil {
		return nil, err
	}
	return &Transaction{session: session}, nil
}

// Stats 返回数据库连接池统计信息。
//
// 返回值:
//   - data.DBStats: 数据库统计信息
func (d *DB) Stats() data.DBStats {
	db := d.engine.DB()
	stats := db.Stats()
	return data.DBStats{
		MaxOpenConnections: stats.MaxOpenConnections,
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
	}
}

// Close 关闭数据库连接，释放资源。
//
// 返回值:
//   - error: 关闭错误
func (d *DB) Close() error {
	return d.engine.Close()
}

// Transaction 是 XORM 事务实现了 data.Transaction 接口。
//
// 字段说明:
//   - session: XORM 会话实例
type Transaction struct {
	session *xorm.Session
}

// Transactor 返回 data.Transactor 接口，用于嵌套事务操作。
//
// 返回值:
//   - data.Transactor: 事务执行器
func (t *Transaction) Transactor() data.Transactor {
	return t
}

// Query 在事务中执行查询。
//
// 参数:
//   - ctx: 上下文
//   - query: SQL 查询语句
//   - args: 查询参数
//
// 返回值:
//   - data.Rows: 查询结果行
//   - error: 错误
func (t *Transaction) Query(ctx context.Context, query string, args ...any) (data.Rows, error) {
	session := t.session.Context(ctx)
	rows, err := session.Query(convertArgs(args)...)
	if err != nil {
		return nil, err
	}
	return &xormRows{rows: rows}, nil
}

// QueryRow 在事务中执行查询并返回单行。
//
// 参数:
//   - ctx: 上下文
//   - query: SQL 查询语句
//   - args: 查询参数
//
// 返回值:
//   - data.Row: 单行查询结果
func (t *Transaction) QueryRow(ctx context.Context, query string, args ...any) data.Row {
	return &xormRow{session: t.session.Context(ctx), query: query, args: args}
}

// Exec 在事务中执行 SQL。
//
// 参数:
//   - ctx: 上下文
//   - query: SQL 语句
//   - args: 参数
//
// 返回值:
//   - data.Result: 执行结果
//   - error: 错误
func (t *Transaction) Exec(ctx context.Context, query string, args ...any) (data.Result, error) {
	result, err := t.session.Context(ctx).Exec(convertArgs(args)...)
	if err != nil {
		return nil, err
	}
	return &xormResult{result: result}, nil
}

// Begin 在事务中开始嵌套事务。
//
// 参数:
//   - ctx: 上下文
//
// 返回值:
//   - data.Transaction: 嵌套事务实例
//   - error: 错误
func (t *Transaction) Begin(ctx context.Context) (data.Transaction, error) {
	session := t.session.Engine().NewSession()
	if err := session.Begin(); err != nil {
		return nil, err
	}
	return &Transaction{session: session}, nil
}

// Stats 返回事务统计信息。
//
// 返回值:
//   - data.DBStats: 数据库统计信息
func (t *Transaction) Stats() data.DBStats {
	return data.DBStats{}
}

// Close 关闭事务，若未提交则自动回滚。
//
// 返回值:
//   - error: 关闭错误（回滚错误）
func (t *Transaction) Close() error {
	return t.session.Rollback()
}

// Commit 提交事务。
func (t *Transaction) Commit() error {
	return t.session.Commit()
}

// Rollback 回滚事务。
func (t *Transaction) Rollback() error {
	return t.session.Rollback()
}

var _ data.Transaction = (*Transaction)(nil)

// convertArgs 将 any 切片转换为 any 切片。
func convertArgs(args []any) []any {
	result := make([]any, len(args))
	copy(result, args)
	return result
}

// xormRows 包装 xorm 的查询结果实现 data.Rows 接口。
type xormRows struct {
	rows  []map[string][]byte
	index int
}

func (r *xormRows) Next() bool {
	return r.index < len(r.rows)
}

func (r *xormRows) Scan(dest ...any) error {
	if r.index >= len(r.rows) {
		return sql.ErrNoRows
	}
	row := r.rows[r.index]
	r.index++
	for i, d := range dest {
		if destBytes, ok := row[fmt.Sprintf("column_%d", i)]; ok {
			// XORM 返回列数据作为字节切片
			// 对常见类型进行简单字符串转换
			switch v := d.(type) {
			case *string:
				*v = string(destBytes)
			case *[]byte:
				*v = destBytes
			case *int, *int64, *int32:
				// 解析整数
				_, _ = fmt.Sscanf(string(destBytes), "%d", v)
			}
		}
	}
	return nil
}

func (r *xormRows) Close() error {
	r.rows = nil
	return nil
}

func (r *xormRows) Err() error {
	return nil
}

// xormRow 包装 xorm 的单行查询实现 data.Row 接口。
type xormRow struct {
	session *xorm.Session
	query   string
	args    []any
	queried bool
	rows    []map[string][]byte
}

func (r *xormRow) Scan(dest ...any) error {
	if !r.queried {
		var err error
		r.rows, err = r.session.Query(append([]any{r.query}, convertArgs(r.args)...)...)
		if err != nil {
			return err
		}
		r.queried = true
	}
	if len(r.rows) == 0 {
		return sql.ErrNoRows
	}
	// 获取第一行并扫描到目标字段
	row := r.rows[0]
	// 对于 SELECT 1，列名通常是 "1" 或 "column_0"
	// 尝试通过常见列名模式查找值
	for i, d := range dest {
		// 尝试不同的列名可能性
		var valBytes []byte
		for _, key := range []string{fmt.Sprintf("column_%d", i), "1", "0", fmt.Sprintf("%d", i)} {
			if b, ok := row[key]; ok {
				valBytes = b
				break
			}
		}
		if valBytes == nil {
			// 尝试第一个可用的键
			for _, b := range row {
				valBytes = b
				break
			}
		}
		if valBytes != nil {
			switch v := d.(type) {
			case *string:
				*v = string(valBytes)
			case *[]byte:
				*v = valBytes
			case *int:
				_, _ = fmt.Sscanf(string(valBytes), "%d", v)
			case *int64:
				_, _ = fmt.Sscanf(string(valBytes), "%d", v)
			case *int32:
				_, _ = fmt.Sscanf(string(valBytes), "%d", v)
			}
		}
	}
	return nil
}

// xormResult 包装 xorm 的执行结果实现 data.Result 接口。
type xormResult struct {
	result sql.Result
}

func (r *xormResult) LastInsertId() (int64, error) {
	return r.result.LastInsertId()
}

func (r *xormResult) RowsAffected() (int64, error) {
	return r.result.RowsAffected()
}

// Repository 是 XORM 泛型 Repository。
//
// 字段说明:
//   - engine: XORM 引擎实例
//
// 类型参数:
//   - T: 实体类型
type Repository[T any] struct {
	engine *xorm.Engine
}

// NewRepository 创建新的泛型 Repository。
//
// 参数:
//   - engine: XORM 引擎实例
//
// 返回值:
//   - *Repository[T]: Repository 实例
func NewRepository[T any](engine *xorm.Engine) *Repository[T] {
	return &Repository[T]{engine: engine}
}

// SessionRepository 是 XORM 会话 Repository。
//
// 字段说明:
//   - session: XORM 会话实例
//
// 类型参数:
//   - T: 实体类型
type SessionRepository[T any] struct {
	session *xorm.Session
}

// NewSessionRepository 创建新的会话 Repository。
//
// 参数:
//   - session: XORM 会话实例
//
// 返回值:
//   - *SessionRepository[T]: Repository 实例
func NewSessionRepository[T any](session *xorm.Session) *SessionRepository[T] {
	return &SessionRepository[T]{session: session}
}

// Create 在会话中插入一个实体。
//
// 参数:
//   - bean: 实体对象指针
//
// 返回值:
//   - error: 错误
func (sr *SessionRepository[T]) Create(bean *T) error {
	_, err := sr.session.Insert(bean)
	return err
}

// CreateBatch 在会话中批量插入实体。
//
// 参数:
//   - beans: 实体切片
//
// 返回值:
//   - error: 错误
func (sr *SessionRepository[T]) CreateBatch(beans []T) error {
	if len(beans) == 0 {
		return nil
	}
	_, err := sr.session.Insert(&beans)
	return err
}

// Delete 在会话中根据 ID 删除实体。
//
// 参数:
//   - id: 实体 ID
//
// 返回值:
//   - error: 错误
func (sr *SessionRepository[T]) Delete(id any) error {
	var t T
	_, err := sr.session.ID(id).Delete(&t)
	return err
}

// DeleteByCondition 在会话中根据条件删除实体。
//
// 参数:
//   - where: WHERE 条件
//   - args: 条件参数
//
// 返回值:
//   - error: 错误
func (sr *SessionRepository[T]) DeleteByCondition(where any, args ...any) error {
	var t T
	_, err := sr.session.Where(where, args...).Delete(&t)
	return err
}

// Update 在会话中更新实体。
//
// 参数:
//   - bean: 实体对象指针
//
// 返回值:
//   - error: 错误
func (sr *SessionRepository[T]) Update(bean *T) error {
	_, err := sr.session.Update(bean)
	return err
}

// UpdateByCondition 在会话中根据条件更新实体。
//
// 参数:
//   - where: WHERE 条件
//   - args: 条件参数
//
// 返回值:
//   - int64: 受影响的行数
//   - error: 错误
func (sr *SessionRepository[T]) UpdateByCondition(where any, args ...any) (int64, error) {
	var t T
	result, err := sr.session.Where(where, args...).Update(&t)
	return result, err
}

// FindByID 在会话中根据 ID 查询实体。
//
// 参数:
//   - id: 实体 ID
//
// 返回值:
//   - *T: 实体指针
//   - error: 错误
func (sr *SessionRepository[T]) FindByID(id any) (*T, error) {
	var t T
	has, err := sr.session.ID(id).Get(&t)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil // 查询不到记录时返回 (nil, nil) 而非错误
	}
	return &t, nil
}

// FindOne 在会话中根据条件查询单个实体。
//
// 参数:
//   - where: WHERE 条件
//   - args: 条件参数
//
// 返回值:
//   - *T: 实体指针
//   - error: 错误
func (sr *SessionRepository[T]) FindOne(where any, args ...any) (*T, error) {
	var t T
	has, err := sr.session.Where(where, args...).Get(&t)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("no entity found matching the condition")
	}
	return &t, nil
}

// FindAll 在会话中查询所有符合条件的实体。
//
// 参数:
//   - where: WHERE 条件，传入 nil 查询所有记录
//   - args: 条件参数
//
// 返回值:
//   - []T: 实体列表
//   - error: 错误
func (sr *SessionRepository[T]) FindAll(where any, args ...any) ([]T, error) {
	var results []T

	if where == nil && len(args) == 0 {
		err := sr.session.Find(&results)
		return results, err
	} else {
		err := sr.session.Where(where, args...).Find(&results)
		return results, err
	}
}

// Count 在会话中统计符合条件的实体数量。
//
// 参数:
//   - where: WHERE 条件
//   - args: 权限参数
//
// 返回值:
//   - int64: 数量
//   - error: 错误
func (sr *SessionRepository[T]) Count(where any, args ...any) (int64, error) {
	var t T

	count, err := sr.session.Where(where, args...).Count(&t)
	return count, err
}

// Raw 在会话中执行原生 SQL 查询。
//
// 参数:
//   - sql: SQL 语句
//   - args: 查询参数
//
// 返回值:
//   - []T: 结果切片
//   - error: 错误
func (sr *SessionRepository[T]) Raw(sql string, args ...any) ([]T, error) {
	var results []T

	err := sr.session.SQL(sql, args...).Find(&results)
	return results, err
}

// Create 插入一个实体。
//
// 参数:
//   - bean: 实体对象指针
//
// 返回值:
//   - error: 错误
func (r *Repository[T]) Create(bean *T) error {
	_, err := r.engine.Insert(bean)
	return err
}

// CreateBatch 批量插入实体。
//
// 参数:
//   - beans: 实体切片
//
// 返回值:
//   - error: 错误
func (r *Repository[T]) CreateBatch(beans []T) error {
	if len(beans) == 0 {
		return nil
	}
	_, err := r.engine.Insert(&beans)
	return err
}

// Delete 根据 ID 删除实体。
//
// 参数:
//   - id: 实体 ID
//
// 返回值:
//   - error: 错误
func (r *Repository[T]) Delete(id any) error {
	var t T
	_, err := r.engine.ID(id).Delete(&t)
	return err
}

// DeleteByCondition 根据条件删除实体。
//
// 参数:
//   - where: WHERE 条件
//   - args: 条件参数
//
// 返回值:
//   - error: 错误
func (r *Repository[T]) DeleteByCondition(where any, args ...any) error {
	var t T
	_, err := r.engine.Where(where, args...).Delete(&t)
	return err
}

// Update 更新实体。
//
// 参数:
//   - bean: 实体对象指针
//
// 返回值:
//   - error: 错误
func (r *Repository[T]) Update(bean *T) error {
	// 尝试同步表结构
	var t T
	if err := r.engine.Sync2(&t); err != nil {
		return err
	}
	_, err := r.engine.Update(bean)
	return err
}

// UpdateByCondition 根据条件更新实体。
//
// 参数:
//   - where: WHERE 条件
//   - args: 条件参数
//
// 返回值:
//   - int64: 受影响的行数
//   - error: 错误
func (r *Repository[T]) UpdateByCondition(where any, args ...any) (int64, error) {
	var t T
	result, err := r.engine.Where(where, args...).Update(&t)
	return result, err
}

// FindByID 根据 ID 查询实体。
//
// 参数:
//
// - id: 实体 ID
//
// 返回值:
//
// - *T: 实体指针
// - error: 错误
func (r *Repository[T]) FindByID(id any) (*T, error) {
	var t T
	has, err := r.engine.ID(id).Get(&t)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil // 查询不到记录时返回 (nil, nil) 而非错误
	}
	return &t, nil
}

// FindOne 根据条件查询单个实体。
//
// 参数:
//   - where: WHERE 条件
//   - args: 条件参数
//
// 返回值:
//   - *T: 实体指针
//   - error: 错误
func (r *Repository[T]) FindOne(where any, args ...any) (*T, error) {
	var t T
	has, err := r.engine.Where(where, args...).Get(&t)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &t, nil
}

// FindAll 查询所有符合条件的实体。
//
// 参数:
//   - where: WHERE 条件，传入 nil 查询所有记录
//   - args: 条件参数
//
// 返回值:
//   - []T: 实体列表
//   - error: 错误
func (r *Repository[T]) FindAll(where any, args ...any) ([]T, error) {
	var results []T
	var err error

	// 尝试同步表结构
	var t T
	if err := r.engine.Sync2(&t); err != nil {
		return nil, err
	}

	if where == nil && len(args) == 0 {
		err = r.engine.Find(&results)
	} else {
		err = r.engine.Where(where, args...).Find(&results)
	}
	return results, err
}

// Count 统计符合条件 的实体数量。
//
// 参数:
//   - where: WHERE 条件
//   - args: 条件参数
//
// 返回值:
//   - int64: 数量
//   - error: 错误
func (r *Repository[T]) Count(where any, args ...any) (int64, error) {
	var t T

	// 尝试同步表结构
	if err := r.engine.Sync2(&t); err != nil {
		return 0, err
	}

	count, err := r.engine.Where(where, args...).Count(&t)
	return count, err
}

// Raw 执行原生 SQL 查询。
//
// 参数:
//   - sql: SQL 语句
//   - args: 查询参数
//
// 返回值:
//   - []T: 结果切片
//   - error: 错误
func (r *Repository[T]) Raw(sql string, args ...any) ([]T, error) {
	var results []T

	// 尝试同步表结构
	var t T
	if err := r.engine.Sync2(&t); err != nil {
		return nil, err
	}

	err := r.engine.SQL(sql, args...).Find(&results)
	return results, err
}

// Client 是 XORM 数据库客户端。
//
// 字段说明:
//   - engine: XORM 引擎实例
type Client struct {
	engine *xorm.Engine
}

// NewClient 创建新的数据库客户端。
//
// 参数:
//   - engine: XORM 引擎实例
//
// 返回值:
//   - *Client: 客户端实例
func NewClient(engine *xorm.Engine) *Client {
	return &Client{engine: engine}
}

// Engine 返回底层的 XORM 引擎实例。
//
// 返回值:
//   - *xorm.Engine: XORM 引擎实例
func (c *Client) Engine() *xorm.Engine {
	return c.engine
}

// Begin 开始一个新事务。
//
// 返回值:
//   - *Transaction: 事务实例
//   - error: 错误
func (c *Client) Begin() (*Transaction, error) {
	session := c.engine.NewSession()
	if err := session.Begin(); err != nil {
		return nil, err
	}
	return &Transaction{session: session}, nil
}

// BeginTx 在事务中执行回调函数。
//
// 参数:
//   - ctx: 上下文
//   - engine: XORM 引擎实例
//   - fc: 回调函数
//
// 返回值:
//   - error: 错误
func BeginTx(ctx context.Context, engine *xorm.Engine, fc func(session *xorm.Session) error) error {
	session := engine.NewSession()
	if err := session.Begin(); err != nil {
		return err
	}
	if err := fc(session); err != nil {
		if rbErr := session.Rollback(); rbErr != nil {
			fmt.Printf("[go-boot] session rollback failed: %v (original error: %v)\n", rbErr, err)
		}
		return err
	}
	return session.Commit()
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
