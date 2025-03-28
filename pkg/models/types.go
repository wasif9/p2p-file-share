package types

import "time"

type Manifest struct {
	Name string `gorm:"primaryKey" json:"name"`
	Hash string `json:"hash"`
	Size int64  `json:"size"`
}

type Heartbeat struct {
	Index       int           `json:"node-index"`
	Uptime      time.Duration `json:"uptime"`
	Utilization int           `json:"utilization"`
}
