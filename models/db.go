package models

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Onicc/frp-panel/defs"
	"github.com/Onicc/frp-panel/utils/logger"
	"github.com/google/uuid"
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
	{
		version: 3,
		name:    "managed_resource_lifecycle",
		up: func(tx *gorm.DB) error {
			// v1 creates tables from the current model definitions while this
			// migration upgrades databases created by older releases. Migrator
			// operations are idempotent and work across SQLite/Postgres.
			// Add the fields individually so this remains compatible with
			// databases that predate the fields. GORM infers the SQL type.
			for _, item := range []struct {
				model  any
				column string
			}{
				{&Client{}, "Enabled"}, {&Client{}, "EnrolledAt"},
				{&Server{}, "EnrolledAt"}, {&Server{}, "LastSeenAt"}, {&Server{}, "BindPort"},
				{&ProxyConfig{}, "PublicID"}, {&ProxyConfig{}, "ManagedBy"},
				{&ProxyConfig{}, "DesiredRevision"}, {&ProxyConfig{}, "LastAppliedAt"}, {&ProxyConfig{}, "LastError"},
			} {
				if !tx.Migrator().HasColumn(item.model, item.column) {
					if err := tx.Migrator().AddColumn(item.model, item.column); err != nil {
						return err
					}
				}
			}
			// Existing rows were created before Enabled existed; make the
			// desired state explicit and infer enrollment from known config.
			if err := tx.Model(&Client{}).Where("enabled = ?", false).Where("stopped = ?", false).Update("enabled", true).Error; err != nil {
				return err
			}
			if err := tx.Model(&Client{}).Where("enrolled_at IS NULL AND COALESCE(length(config_content), 0) > 0").Update("enrolled_at", gorm.Expr("updated_at")).Error; err != nil {
				return err
			}
			if err := tx.Model(&Server{}).Where("enrolled_at IS NULL AND COALESCE(length(config_content), 0) > 0").Update("enrolled_at", gorm.Expr("updated_at")).Error; err != nil {
				return err
			}
			var proxies []ProxyConfig
			if err := tx.Find(&proxies).Error; err != nil {
				return err
			}
			for _, proxy := range proxies {
				updates := map[string]any{}
				if proxy.PublicID == "" {
					updates["public_id"] = uuid.NewString()
				}
				if proxy.ManagedBy == "" {
					if proxy.Type == "tcp" || proxy.Type == "udp" {
						updates["managed_by"] = "tunnel"
					} else {
						updates["managed_by"] = "legacy"
					}
				}
				if len(updates) > 0 {
					if err := tx.Model(&ProxyConfig{}).Where("id = ?", proxy.ID).Updates(updates).Error; err != nil {
						return err
					}
				}
			}
			var servers []Server
			if err := tx.Find(&servers).Error; err != nil {
				return err
			}
			for _, server := range servers {
				if server.BindPort != 0 || len(server.ConfigContent) == 0 {
					continue
				}
				var config struct {
					BindPort int `json:"bindPort"`
				}
				if json.Unmarshal(server.ConfigContent, &config) == nil && config.BindPort != 0 {
					if err := tx.Model(&Server{}).Where("server_id = ?", server.ServerID).Update("bind_port", config.BindPort).Error; err != nil {
						return err
					}
				}
			}
			for _, item := range []struct {
				model       any
				field, name string
			}{
				{&Client{}, "EnrolledAt", "idx_clients_enrolled_at"},
				{&Server{}, "EnrolledAt", "idx_servers_enrolled_at"},
				{&Server{}, "LastSeenAt", "idx_servers_last_seen_at"},
				{&ProxyConfig{}, "ManagedBy", "idx_proxy_config_managed_by"},
				{&ProxyConfig{}, "PublicID", "idx_proxy_config_public_id"},
			} {
				if !tx.Migrator().HasIndex(item.model, item.name) {
					if err := tx.Migrator().CreateIndex(item.model, item.field); err != nil {
						return err
					}
				}
			}
			return nil
		},
	},
	{
		version: 4,
		name:    "server_api_port",
		up: func(tx *gorm.DB) error {
			if !tx.Migrator().HasColumn(&Server{}, "ServerAPIPort") {
				if err := tx.Migrator().AddColumn(&Server{}, "ServerAPIPort"); err != nil {
					return err
				}
			}
			return tx.Model(&Server{}).Where("server_api_port = 0").Update("server_api_port", defs.DefaultServerAPIPort).Error
		},
	},
	{
		version: 5,
		name:    "endpoint_last_seen_ip",
		up: func(tx *gorm.DB) error {
			for _, model := range []any{&Client{}, &Server{}} {
				if !tx.Migrator().HasColumn(model, "LastSeenIP") {
					if err := tx.Migrator().AddColumn(model, "LastSeenIP"); err != nil {
						return err
					}
				}
			}
			return nil
		},
	},
	{
		version: 6,
		name:    "client_location_sources",
		up: func(tx *gorm.DB) error {
			for _, column := range []string{"ReportedIP", "ReportedAt", "LocationIPOverride"} {
				if !tx.Migrator().HasColumn(&Client{}, column) {
					if err := tx.Migrator().AddColumn(&Client{}, column); err != nil {
						return err
					}
				}
			}
			return nil
		},
	},
	{
		version: 7,
		name:    "unique_tunnel_name_per_client",
		up: func(tx *gorm.DB) error {
			return tx.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_tunnel_client_name ON proxy_config (tenant_id, user_id, origin_client_id, name) WHERE managed_by = 'tunnel' AND deleted_at IS NULL").Error
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
