package gorm

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"

	"github.com/ao9911/go-matrix/util/xtime"
)

type Config struct {
	DebugMode    bool           `toml:"debug_mode"`
	DriverName   string         `toml:"driver_name"`
	DSN          string         `toml:"dsn"`
	ReadDSN      []string       `toml:"read_dsn"`
	MaxIdleConns int            `toml:"max_idle_conns"`
	MaxOpenConns int            `toml:"max_open_conns"`
	MaxLifetime  xtime.Duration `toml:"max_lifetime"`
	MaxIdleTime  xtime.Duration `toml:"max_idle_time"`
}

func NewORM(c *Config) (orm *gorm.DB) {
	orm, err := Open(c)
	if err != nil {
		panic(fmt.Errorf("open gorm mysql: %w", err))
	}
	return
}

func Open(c *Config) (*gorm.DB, error) {
	if c == nil {
		return nil, errors.New("gorm: nil config")
	}
	if c.DSN == "" {
		return nil, errors.New("gorm: empty DSN")
	}

	gormConfig := &gorm.Config{}
	if c.DebugMode {
		gormConfig.Logger = logger.Default.LogMode(logger.Info)
	}
	newDialector := func(dsn string) gorm.Dialector {
		return mysql.New(mysql.Config{
			DriverName:                c.DriverName,
			DSN:                       dsn,
			DefaultStringSize:         256,   // add default size for string fields, by default, will use db type `longtext` for fields without size, not a primary key, no index defined and don't have default values
			DisableDatetimePrecision:  true,  // disable datetime precision support, which not supported before MySQL 5.6
			DontSupportRenameIndex:    false, // drop & create index when rename index, rename index not supported before MySQL 5.7, MariaDB
			DontSupportRenameColumn:   true,  // use change when rename column, rename rename not supported before MySQL 8, MariaDB
			SkipInitializeWithVersion: false, // smart configure based on used version
		})
	}
	orm, err := gorm.Open(newDialector(c.DSN), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}
	if len(c.ReadDSN) > 0 {
		replicas := make([]gorm.Dialector, 0, len(c.ReadDSN))
		for _, dsn := range c.ReadDSN {
			replicas = append(replicas, newDialector(dsn))
		}
		err = orm.Use(
			dbresolver.Register(dbresolver.Config{
				Replicas: replicas,
				Policy:   dbresolver.RandomPolicy{},
			}).
				SetMaxIdleConns(c.MaxIdleConns).
				SetConnMaxLifetime(time.Duration(c.MaxLifetime)).
				SetMaxOpenConns(c.MaxOpenConns).
				SetConnMaxIdleTime(time.Duration(c.MaxIdleTime)),
		)
		if err != nil {
			closeGORM(orm)
			return nil, fmt.Errorf("register gorm dbresolver: %w", err)
		}
	} else {
		db, err := orm.DB()
		if err != nil {
			return nil, fmt.Errorf("get gorm sql db: %w", err)
		}
		db.SetMaxIdleConns(c.MaxIdleConns)
		db.SetMaxOpenConns(c.MaxOpenConns)
		db.SetConnMaxLifetime(time.Duration(c.MaxLifetime))
		db.SetConnMaxIdleTime(time.Duration(c.MaxIdleTime))
	}
	return orm, nil
}

func closeGORM(orm *gorm.DB) {
	db, err := orm.DB()
	if err == nil {
		_ = db.Close()
	}
}
