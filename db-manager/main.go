package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	types "github.com/wasif9/p2p-file-share/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var err error

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Connection to the database
	dsn := fmt.Sprintf("host=localhost user=postgres password=%s dbname=registry%d port=5432 sslmode=disable TimeZone=UTC", cfg.Pg_password, cfg.Index)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&types.Manifest{})
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/"+cfg.Version+"/records/", createRecordHandler(db))
	http.HandleFunc("/api/"+cfg.Version+"/records", createRecordsHandler(db))
	http.HandleFunc("/api/"+cfg.Version+"/kill", killHandler)
	http.HandleFunc("/api/"+cfg.Version+"/heartbeat", heartbeatHandler)

	log.Printf("Server starting on port %s...\n", cfg.Port)
	err = http.ListenAndServe(net.JoinHostPort("localhost", cfg.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
