package types

import "time"

type Manifest struct {
	Name      string `gorm:"primaryKey" json:"name"`
	Hash      string `json:"hash"`
	Size      int64  `json:"size"`
	Timestamp uint   `gorm:"unique" json:"timestamp"`
}

type Heartbeat struct {
	Index       int           `json:"node-index"`
	Uptime      time.Duration `json:"uptime"`
	Utilization int           `json:"utilization"`
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
