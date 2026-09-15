package models

import (
	"context"
	"fmt"
	"time"

	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/utils/logger"
	"gorm.io/gorm"
)

type dbManagerImpl struct {
	DBs           map[string]map[string]*gorm.DB // map[db type]map[db role]*gorm.DB
	defaultDBType string
	debug         bool
}

func (dbm *dbManagerImpl) Init() {
	for _, dbGroup := range dbm.DBs {
		for _, db := range dbGroup {
			ctx := context.Background()
			if err := runMigrations(db); err != nil {
				logger.Logger(ctx).WithError(err).Fatal("cannot migrate database")
			}
		}
	}
}

type schemaMigration struct {
	Version   uint `gorm:"primaryKey"`
	Name      string
	AppliedAt time.Time
}

type migration struct {
	version uint
	name    string
	up      func(*gorm.DB) error
}

var migrations = []migration{
	{
		version: 1,
		name:    "create_v2_schema",
		up: func(tx *gorm.DB) error {
			models := []any{
				&User{}, &UserGroup{}, &Client{}, &AgentEnrollment{}, &Server{}, &Cert{},
				&Worker{}, &ProxyConfig{}, &ProxyStats{}, &HistoryProxyStats{},
				&Network{}, &WireGuard{}, &Endpoint{}, &WireGuardLink{},
			}
			for _, model := range models {
				if tx.Migrator().HasTable(model) {
					continue
				}
				if err := tx.Migrator().CreateTable(model); err != nil {
					return err
				}
			}
			return nil
		},
	},
	{
		version: 2,
		name:    "create_server_enrollments",
		up: func(tx *gorm.DB) error {
			if tx.Migrator().HasTable(&ServerEnrollment{}) {
				return nil
			}
			return tx.Migrator().CreateTable(&ServerEnrollment{})
		},
	},
}

func runMigrations(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if !tx.Migrator().HasTable(&schemaMigration{}) {
			if err := tx.Migrator().CreateTable(&schemaMigration{}); err != nil {
				return fmt.Errorf("create migration table: %w", err)
			}
		}

		for _, item := range migrations {
			var count int64
			if err := tx.Model(&schemaMigration{}).Where("version = ?", item.version).Count(&count).Error; err != nil {
				return fmt.Errorf("read migration %d: %w", item.version, err)
			}
			if count != 0 {
				continue
			}
			if err := item.up(tx); err != nil {
				return fmt.Errorf("apply migration %d (%s): %w", item.version, item.name, err)
			}
			if err := tx.Create(&schemaMigration{Version: item.version, Name: item.name, AppliedAt: time.Now().UTC()}).Error; err != nil {
				return fmt.Errorf("record migration %d: %w", item.version, err)
			}
		}
		return nil
	})
}

func NewDBManager(defaultDBType string) *dbManagerImpl {
	dbs := map[string]map[string]*gorm.DB{}
	return &dbManagerImpl{
		DBs:           dbs,
		defaultDBType: defaultDBType,
	}
}

func (dbm *dbManagerImpl) GetDB(dbType string, dbRole string) *gorm.DB {
	return dbm.DBs[dbType][dbRole]
}

func (dbm *dbManagerImpl) SetDB(dbType string, dbRole string, db *gorm.DB) {
	if dbm.DBs[dbType] == nil {
		dbm.DBs[dbType] = map[string]*gorm.DB{}
	}
	dbm.DBs[dbType][dbRole] = db
}

func (dbm *dbManagerImpl) RemoveDB(dbType string, dbRole string) {
	if dbm.DBs[dbType] == nil {
		return
	}
	delete(dbm.DBs[dbType], dbRole)
}

func (dbm *dbManagerImpl) GetDefaultDB() *gorm.DB {
	dbGroup := dbm.DBs[dbm.defaultDBType]
	if dbm.debug {
		return dbGroup[defs.DBRoleDefault].Debug()
	}
	return dbGroup[defs.DBRoleDefault]
}

func (dbm *dbManagerImpl) SetDebug(debug bool) {
	dbm.debug = debug
}
