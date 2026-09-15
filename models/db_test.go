package models

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestRunMigrationsIsIdempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatal(err)
	}

	if err := runMigrations(db); err != nil {
		t.Fatalf("first migration: %v", err)
	}
	if err := runMigrations(db); err != nil {
		t.Fatalf("second migration: %v", err)
	}

	var count int64
	if err := db.Model(&schemaMigration{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != int64(len(migrations)) {
		t.Fatalf("migration records = %d, want %d", count, len(migrations))
	}
	for _, model := range []any{&User{}, &Client{}, &AgentEnrollment{}, &Server{}, &ProxyConfig{}, &WireGuardLink{}} {
		if !db.Migrator().HasTable(model) {
			t.Fatalf("missing table for %T", model)
		}
	}
}
