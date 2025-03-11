package types

import "time"

// todo: move this to a models.go file/package
/****************************************
* Struct for the database records
*
* **************************************/
type Manifest struct {
	Name string `gorm:"primaryKey" json:"name"`
	Hash string `json:"hash"`
}

// TODO: move this to common package so that client can use too
type Heartbeat struct {
	Index       int           `json:"node-index"`
	Uptime      time.Duration `json:"uptime"`
	Utilization int           `json:"utilization"`
}
