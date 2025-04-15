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
type ManifestData struct {
	FileName string         `json:"fileName"`
	FileCID  string         `json:"fileCID"`
	Chunks   map[int]string `json:"chunks"` // index => chunk CID
}

type DownloadRequest struct {
	FileName   string `json:"filename"`
	ChunkIndex int    `json:"chunk_index"`
	Type       string `json:"type"`
}

type Heartbeat struct {
	Index       int           `json:"node-index"`
	Uptime      time.Duration `json:"uptime"`
	LeaderIndex int           `json:"leader-index"`
	Timestamp   int           `json:"timestamp"`
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
