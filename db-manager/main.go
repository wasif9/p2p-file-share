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
var configuration Configuration

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	configuration = Configure()

	// Connection to the database
	dsn := fmt.Sprintf("host=localhost user=postgres password=%s dbname=registry%d port=5432 sslmode=disable TimeZone=UTC", configuration.PG_PASSWORD, configuration.index)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to database as %s", db.Name())

	err = db.AutoMigrate(&types.Manifest{})
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/"+configuration.VERSION+"/records/", createRecordHandler(db))
	http.HandleFunc("/api/"+configuration.VERSION+"/records", createRecordsHandler(db))
	http.HandleFunc("/api/"+configuration.VERSION+"/kill", killHandler)
	http.HandleFunc("/api/"+configuration.VERSION+"/heartbeat", heartbeatHandler)

	log.Printf("Server starting on port %s...\n", configuration.PORT)
	err = http.ListenAndServe(net.JoinHostPort("localhost", configuration.PORT), nil)
	if err != nil {
		log.Fatal(err)
	}
}
