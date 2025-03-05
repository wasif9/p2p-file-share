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

/***********************
* This function inserts a record into the registry database
* Input:
*		*gorm.DB db
* 		struct Record record
* Returns:
* 		None
*
************************/
func insertRecord(db *gorm.DB, record *Manifest) (newRecord Manifest, Error error) {

	result := db.Create(&record)

	return *record, result.Error

}

/***********************
* This function deletes a record by filename from the registry database.
* Input:
*		*gorm.DB db
* 		string filename
* Returns:
* 		None
*
************************/
func deleteRecord(db *gorm.DB, filename string) {

	result := db.Where("filename = ?", filename).Delete(&Manifest{})

	if result.RowsAffected == 0 {
		fmt.Println("No record found to delete")
	} else {
		fmt.Println("Record deleted successfully")
	}

}

/***********************
* This function queries the database for all the records. It returns an array of
* all the records
* Input:
*		*gorm.DB db
* Returns:
* 		[]Record record
*
************************/
func queryRecord(db *gorm.DB) (record []Manifest, Error error) {

	result := db.Find(&record)

	return record, result.Error

}

/***********************
* This function queries the database for a file and returns the record for the file.
* Input:
*		*gorm.DB db
* 		filename
* Returns:
* 		Record record
*
************************/
func querySingleRecord(db *gorm.DB, filename string) (record Manifest, Error error) {

	result := db.Where("filename = ?", filename).First(&record)

	return record, result.Error

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
