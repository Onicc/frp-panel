package models

import "time"

// AgentEnrollment stores only a one-way token hash. The permanent agent
// credential is deterministically derived during the single redemption and is
// also stored only as a hash on Client.
type AgentEnrollment struct {
	ID        string     `gorm:"primaryKey;type:varchar(36)"`
	TokenHash string     `gorm:"uniqueIndex;not null;type:varchar(80)"`
	ClientID  string     `gorm:"uniqueIndex;not null;type:varchar(255)"`
	UserID    int        `gorm:"index;not null"`
	TenantID  int        `gorm:"index;not null"`
	ExpiresAt time.Time  `gorm:"index;not null"`
	UsedAt    *time.Time `gorm:"index"`
	CreatedAt time.Time
}

func (*AgentEnrollment) TableName() string { return "agent_enrollments" }

// ServerEnrollment is the FRPS equivalent of AgentEnrollment. The bootstrap
// token is single-use and the long-lived Server credential is stored only as a
// hash, so neither credential can be recovered from the Master database.
type ServerEnrollment struct {
	ID        string     `gorm:"primaryKey;type:varchar(36)"`
	TokenHash string     `gorm:"uniqueIndex;not null;type:varchar(80)"`
	ServerID  string     `gorm:"uniqueIndex;not null;type:varchar(255)"`
	UserID    int        `gorm:"index;not null"`
	TenantID  int        `gorm:"index;not null"`
	ExpiresAt time.Time  `gorm:"index;not null"`
	UsedAt    *time.Time `gorm:"index"`
	CreatedAt time.Time
}

func (*ServerEnrollment) TableName() string { return "server_enrollments" }
