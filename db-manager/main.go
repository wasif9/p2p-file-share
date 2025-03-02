package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// todo: move this to a models.go file/package
type Record struct {
	gorm.Model
	Name string
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "You requested: %s\n", r.URL.Path)
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
func insertRecord(db *gorm.DB, record Record) {
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
	result := db.Delete(&Record{}, id)

	if result.Error != nil {
		log.Fatal(result.Error)
	}

}

const PORT = "8080"

func main() {
	var err error

	// prefix-based pattern matching means any route with
	// the / prefix (all of them) will be served by function 'handler'
	http.HandleFunc("/", handler)

	// TODO: read this from config
	index := 0 // the index of this DB instance

	dsn := fmt.Sprintf("host=localhost user=postgres password=password dbname=registry%d port=5432 sslmode=disable TimeZone=UTC", index)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n\n", db.Config)

	myRecord := Record{
		Name: "pete",
	}

	fmt.Println(myRecord.Model.ID)
	fmt.Println(myRecord.Name)

	err = db.AutoMigrate(&Record{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Server starting on port %s...\n", PORT)
	err = http.ListenAndServe(net.JoinHostPort("localhost", PORT), nil)
	if err != nil {
		log.Fatal(err)
	}
}
