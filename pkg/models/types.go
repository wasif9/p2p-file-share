package types

import (
	"time"

	"gorm.io/gorm"
)

type Manifest struct {
	Timestamp uint           `gorm:"primarykey"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Name string `gorm:"unique" json:"name"`
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

type Heartbeat struct {
	Index       int           `json:"node-index"`
	Uptime      time.Duration `json:"uptime"`
	LeaderIndex int           `json:"leader-index"`
	Timestamp   uint          `json:"timestamp"`
}

type DbMgrConfig struct {
	Address     string `json:"address"`
	Index       int    `json:"index"`
	Version     string `json:"version"`
	Pg_host     string `json:"pg-host"`
	Pg_user     string `json:"pg-user"`
	Pg_password string `json:"pg-password"`
	Pg_database string `json:"pg-database"`
	Pg_port     string `json:"pg-port"`
}

type RevProxyConfig struct {
	Address string `json:"address"`
}

type SuperConfig struct {
	RpConfig         RevProxyConfig `json:"reverse-proxy"`
	DbManagerConfigs []DbMgrConfig  `json:"db-managers"`
}
