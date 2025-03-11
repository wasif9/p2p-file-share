package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var err error

// todo: move this to a models.go file/package
/****************************************
* Struct for the database records
*
* **************************************/
type Manifest struct {
	Name string `gorm:"primaryKey" json:"name"`
	Hash string `json:"hash"`
}

// Server Instance Configuration:
// TODO: read this from a config.json or similar
const index = 0 // the index of this DB instance
const PORT = "8081"
const VERSION = "v1"
const PG_PASSWORD = "password"

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Connection to the database
	dsn := fmt.Sprintf("host=localhost user=postgres password=%s dbname=registry%d port=5432 sslmode=disable TimeZone=UTC", PG_PASSWORD, index)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connected to database as %s", db.Name())

	err = db.AutoMigrate(&Manifest{})
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/api/"+VERSION+"/records/", createRecordHandler(db))
	http.HandleFunc("/api/"+VERSION+"/records", createRecordsHandler(db))
	http.HandleFunc("/api/"+VERSION+"/kill", killHandler)
	http.HandleFunc("/api/"+VERSION+"/heartbeat", heartbeatHandler)

	log.Printf("Server starting on port %s...\n", PORT)
	err = http.ListenAndServe(net.JoinHostPort("localhost", PORT), nil)
	if err != nil {
		log.Fatal(err)
	}
}
