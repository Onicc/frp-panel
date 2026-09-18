package models

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Onicc/frp-panel/utils"
	v1 "github.com/fatedier/frp/pkg/config/v1"
	"gorm.io/gorm"
)

type Server struct {
	*ServerEntity
}

type ServerEntity struct {
	ServerID      string            `json:"server_id" gorm:"uniqueIndex;not null;primaryKey"`
	TenantID      int               `json:"tenant_id" gorm:"not null,index"`
	UserID        int               `json:"user_id" gorm:"not null"`
	ServerIP      string            `json:"server_ip"`
	ConfigContent []byte            `json:"config_content"`
	ConnectSecret string            `json:"-" gorm:"not null"`
	Comment       string            `json:"comment"`
	FrpsUrls      GormArray[string] `json:"frps_urls"`
	EnrolledAt    *time.Time        `json:"enrolled_at" gorm:"index"`
	LastSeenAt    *time.Time        `json:"last_seen_at" gorm:"index"`
	LastSeenIP    string            `json:"last_seen_ip"`
	BindPort      int               `json:"bind_port"`
	ServerAPIPort int               `json:"server_api_port" gorm:"not null;default:8999"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (*Server) TableName() string {
	return "servers"
}

func (s *ServerEntity) GetConfigContent() (*v1.ServerConfig, error) {
	if len(s.ConfigContent) == 0 {
		return nil, fmt.Errorf("config content is empty")
	}

	var cfg v1.ServerConfig
	err := json.Unmarshal(s.ConfigContent, &cfg)
	return &cfg, err
}

func (s *ServerEntity) SetConfigContent(cfg *v1.ServerConfig) error {
	raw, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	s.ConfigContent = raw
	return nil
}

func (s *ServerEntity) ConfigEqual(cfg *v1.ServerConfig) bool {
	raw, err := json.Marshal(cfg)
	if err != nil {
		return false
	}

	return utils.SHA256(s.ConfigContent) == utils.SHA256(raw)
}
