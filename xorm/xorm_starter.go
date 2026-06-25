package xorm

import src "github.com/xudefa/go-boot"

// Starter 实现 xorm 数据库连接启动器。
//
// 用于在应用启动时初始化数据库连接并运行迁移。
type Starter struct {
	engine interface {
		Sync2(beans ...any) error
	}
	autoMigrate []any
}

// NewStarter 创建新的数据库启动器。
//
// 参数:
//   - engine: xorm 引擎
//   - models: 要迁移的模型
//
// 返回值:
//   - *Starter: 启动器实例
func NewStarter(engine interface {
	Sync2(beans ...any) error
}, models []any) *Starter {
	return &Starter{
		engine:      engine,
		autoMigrate: models,
	}
}

// Starter 实现 boot.Starter 接口。
func (s *Starter) Starter() error {
	if s.engine == nil {
		return nil
	}
	if len(s.autoMigrate) > 0 {
		return s.engine.Sync2(s.autoMigrate...)
	}
	return nil
}

// NewAutoMigrateStarter 创建自动迁移启动器，在应用启动时自动同步数据库表结构。
//
// 参数:
//   - engine: 支持 Sync2 方法的 xorm 引擎
//   - models: 需要同步的模型列表
//
// 返回值:
//   - boot.Starter: 启动器实例
func NewAutoMigrateStarter(engine interface {
	Sync2(beans ...any) error
}, models ...any) src.Starter {
	return &Starter{
		engine:      engine,
		autoMigrate: models,
	}
}
