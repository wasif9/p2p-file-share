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

type Node struct {
	IP     string `json:"server-ip"`
	Port   string `json:"port"`
	Index  int    `json:"node-index"`
	Status string `json:"node-status"`
}
