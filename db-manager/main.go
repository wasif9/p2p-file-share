package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// todo: move this to a models.go file/package
type Record struct {
	gorm.Model
	//ID   uint
	Name string
}

func handlePostRecord(w http.ResponseWriter, r *http.Request) {
	var response string = "handling POST\n"

	w.WriteHeader(http.StatusCreated)
	// write the response
	fmt.Fprintln(w, response)
}
func handleGetRecord(w http.ResponseWriter, r *http.Request) {
	s := strings.Split(r.URL.Path, "/")

	var response string = "handling GET " + s[len(s)-1]

	fmt.Fprintln(w, response)

}
func handleDeleteRecord(w http.ResponseWriter, r *http.Request) {
	var response string = "handling DELETE\n"
	fmt.Fprintln(w, response)

}

// Handles http requests to the route '/records'
func recordHandler(w http.ResponseWriter, r *http.Request) {
	handlers := map[string]func(http.ResponseWriter, *http.Request){
		http.MethodPost:   handlePostRecord,
		http.MethodGet:    handleGetRecord,
		http.MethodDelete: handleDeleteRecord,
	}

	handlers[r.Method](w, r)
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
func insertRecord(db *gorm.DB, record *Record) {

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

// Server Instance Configuration:
// TODO: read this from a config.json or similar
const index = 0 // the index of this DB instance
const PORT = "8080"
const VERSION = "v1"

func main() {
	var err error // re-useable error

	basePath, err := url.JoinPath("/", "api", VERSION, "records", "/")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc(basePath, recordHandler)

	// Connection to the database
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

	insertRecord(db, &myRecord)

	deleteRecord(db, 5)

	fmt.Printf("Server starting on port %s...\n", PORT)
	err = http.ListenAndServe(net.JoinHostPort("localhost", PORT), nil)
	if err != nil {
		log.Fatal(err)
	}
}
