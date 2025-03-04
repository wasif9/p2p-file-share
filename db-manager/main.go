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

func replaceWithGarethsInsertMethod(manifest Manifest) (Manifest, error) {
	return manifest, nil
}

func replaceWithGarethsSelectMethod(name string) (Manifest, error) {
	return Manifest{
		Name: name,
		Hash: "xyz",
	}, nil
}

func replaceWithGarethsDeleteMethod() error {
	return nil
}

// todo: move this to a models.go file/package
type Manifest struct {
	Name string `gorm:"primaryKey" json:"name"`
	Hash string `json:"hash"`
}

/***********************
* This function inserts a record into the registry database
* Input:
*		*gorm.DB db
* 		struct Record record
* Returns:
* 		None
*
************************/
func insertRecord(db *gorm.DB, record *Manifest) {

	result := db.Create(&record)

	if result.Error != nil {
		log.Fatal(result.Error)
	}

}

/***********************
* This function deletes a record from the registry database
* Input:
*		*gorm.DB db
* 		filename
* Returns:
* 		None
*
************************/
func deleteRecord(db *gorm.DB, id uint) {
	result := db.Delete(&Manifest{}, id)

	if result.Error != nil {
		log.Fatal(result.Error)
	}

}

// Server Instance Configuration:
// TODO: read this from a config.json or similar
const index = 0 // the index of this DB instance
const PORT = "8080"
const VERSION = "v1"
const PG_PASSWORD = "password"

func main() {

	http.HandleFunc("/api/"+VERSION+"/records/", recordHandler)
	http.HandleFunc("/api/"+VERSION+"/records", recordsHandler)

	// Connection to the database
	dsn := fmt.Sprintf("host=localhost user=postgres password=%s dbname=registry%d port=5432 sslmode=disable TimeZone=UTC", PG_PASSWORD, index)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&Manifest{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server starting on port %s...\n", PORT)
	err = http.ListenAndServe(net.JoinHostPort("localhost", PORT), nil)
	if err != nil {
		log.Fatal(err)
	}
}
