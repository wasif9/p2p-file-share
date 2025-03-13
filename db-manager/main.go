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
	dsn := fmt.Sprintf("host=localhost user=postgres password=%s dbname=registry%d port=5432 sslmode=disable TimeZone=UTC", cfg.PG_PASSWORD, cfg.index)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to database as %s", db.Name())

	err = db.AutoMigrate(&types.Manifest{})
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/"+cfg.VERSION+"/records/", createRecordHandler(db))
	http.HandleFunc("/api/"+cfg.VERSION+"/records", createRecordsHandler(db))
	http.HandleFunc("/api/"+cfg.VERSION+"/kill", killHandler)
	http.HandleFunc("/api/"+cfg.VERSION+"/heartbeat", heartbeatHandler)

	log.Printf("Server starting on port %s...\n", cfg.PORT)
	err = http.ListenAndServe(net.JoinHostPort("localhost", cfg.PORT), nil)
	if err != nil {
		log.Fatal(err)
	}
}
