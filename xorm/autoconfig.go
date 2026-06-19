// Package xorm 提供 XORM ORM 的自动配置。
//
// 当 xorm.enabled=true 时自动启用，从 Environment 中读取 xorm.type、xorm.host、xorm.port、
// xorm.username、xorm.password、xorm.database、xorm.max-open-conns 等配置项，
// 创建并注册 XORM DB Bean 到 IoC 容器中（Bean ID: xormDB），实现 data.Transactor 接口。
//
// 同时会自动注册数据库健康指标（Bean ID: xormDatabaseHealthIndicator），
// 使用 PingContext 进行数据库连接检查。
package xorm

import (
	"context"

	xormcore "github.com/xudefa/go-boot-xorm"

	"github.com/xudefa/go-boot/actuator"
	"github.com/xudefa/go-boot/boot"
	"github.com/xudefa/go-boot/condition"
	"github.com/xudefa/go-boot/constants"
	"github.com/xudefa/go-boot/core"
	"github.com/xudefa/go-boot/data"
)

// init 注册 XORM 自动配置，由 xorm.enabled=true 条件控制。
func init() {
	boot.RegisterAutoConfig(&XormAutoConfiguration{},
		condition.OnProperty(constants.XORMEnabled, constants.ConditionTrue),
	)
}

// XormAutoConfiguration XORM ORM 的自动配置。
//
// 从 Environment 中读取 xorm.type、xorm.host、xorm.port、xorm.username、xorm.database 等配置项，
// 创建 XORM DB 连接（支持 MySQL / PostgreSQL）并注册到 IoC 容器中，实现 data.Transactor 接口。
// 启用条件：xorm.enabled=true
type XormAutoConfiguration struct{}

// Configure 执行自动配置逻辑，创建 XORM DB 连接并注册为 Bean。
//
// 同时注册 XORM 数据库健康指标，用于监控数据库连接状态。
func (x *XormAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	env := ctx.Environment()

	dbType := env.GetString(constants.XORMType, constants.DefaultXORMType)
	var db *xormcore.DB
	var err error

	opts := []xormcore.Option{
		xormcore.WithHost(env.GetString(constants.XORMHost, constants.DefaultXORMHost)),
		xormcore.WithPort(env.GetInt(constants.XORMPort, constants.DefaultXORMPort)),
		xormcore.WithUser(env.GetString(constants.XORMUsername, constants.DefaultXORMUsername)),
		xormcore.WithPassword(env.GetString(constants.XORMPassword, constants.DefaultXORMPassword)),
		xormcore.WithDBName(env.GetString(constants.XORMDatabase, constants.DefaultXORMDatabase)),
		xormcore.WithMaxOpenConns(env.GetInt(constants.XORMMaxOpenConns, constants.DefaultXORMMaxOpenConns)),
		xormcore.WithMaxIdleConns(env.GetInt(constants.XORMMaxIdleConns, constants.DefaultXORMMaxIdleConns)),
		xormcore.WithShowSQL(env.GetBool(constants.XORMShowSQL, constants.DefaultXORMShowSQL)),
	}
	if charset := env.GetString(constants.XORMCharset, ""); charset != "" {
		opts = append(opts, xormcore.WithCharset(charset))
	}

	switch dbType {
	case "postgres":
		db, err = xormcore.OpenPostgreSQL(opts...)
	default:
		db, err = xormcore.OpenMySQL(opts...)
	}
	if err != nil {
		panic(err)
	}

	if err := ctx.Register(constants.XORMDBBeanID,
		core.Bean(db),
		core.Singleton(),
	); err != nil {
		return err
	}

	xormHealthIndicator := actuator.NewDatabaseHealthIndicator(func(ctx context.Context) error {
		sqlDB := db.Engine().DB()
		if sqlDB == nil {
			return nil
		}
		return sqlDB.PingContext(ctx)
	})

	if err := ctx.Register(constants.XORMDatabaseHealthIndicatorBeanID,
		core.Bean(xormHealthIndicator),
		core.Singleton(),
	); err != nil {
		return err
	}

	return nil
}

var _ data.Transactor = (*xormcore.DB)(nil)
