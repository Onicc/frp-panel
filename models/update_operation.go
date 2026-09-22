package models

import "time"

// UpdateOperation survives Master/Server restarts and records the actual
// outcome rather than treating a remote RPC acknowledgement as success.
type UpdateOperation struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	Kind         string     `json:"kind" gorm:"index:idx_update_target,priority:1;not null"`
	TargetID     string     `json:"targetId" gorm:"index:idx_update_target,priority:2;not null"`
	TenantID     int        `json:"-" gorm:"index;not null"`
	UserID       int        `json:"-" gorm:"not null"`
	Mode         string     `json:"mode" gorm:"not null"`
	TargetCommit string     `json:"targetCommit" gorm:"not null"`
	Version      string     `json:"version" gorm:"not null"`
	State        string     `json:"state" gorm:"not null"`
	Error        string     `json:"error,omitempty"`
	StartedAt    time.Time  `json:"startedAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
}
