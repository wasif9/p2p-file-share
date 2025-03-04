package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// todo: move this to a models.go file/package
/****************************************
* Struct for the database records
*
* **************************************/
type Record struct {
	//gorm.Model
	ID        uint      `gorm:"primarykey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	Filename  string    `gorm:"not null"` //`gorm:"unique;not null"`
	Filehash  string    `gorm:"not null"` //`gorm:"unique;not null"`
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
func insertRecord(db *gorm.DB, record *Record) {

	result := db.Create(&record)

	if result.Error != nil {
		log.Fatal(result.Error)
	}

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

	result := db.Where("filename = ?", filename).Delete(&Record{})

	if result.RowsAffected == 0 {
		fmt.Println("No record found to delete")
	} else {
		fmt.Println("Record deleted successfully")
	}

	// if result.Error != nil {
	// 	log.Fatal(result.Error)
	// }

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
func queryRecord(db *gorm.DB) (record []Record) {

	result := db.Find(&record)

	if result.Error != nil {
		log.Fatal(result.Error)
	}

	return record

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
func querySingleRecord(db *gorm.DB, filename string) (record Record) {

	//result := db.Select("filehash").Where("filename = ?", filename).First(&record)
	result := db.Where("filename = ?", filename).First(&record)

	if result.Error != nil {
		log.Fatal(result.Error)
	}

	return record

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

	// myRecord1 := Record{
	// 	Filename: "textfile_1.txt",
	// 	Filehash: "9a1c5ec4f901367900c362ff69e22fb4802e3919116388a191fd2c54034f8089",
	// }

	// myRecord2 := Record{
	// 	Filename: "textfile_2.txt",
	// 	Filehash: "5a7cb2181dc560db23c644d6c2b3fe26146bb26ce4579d6fc6c5b136fbd6242c",
	// }

	// fmt.Println(myRecord1.ID)
	// fmt.Println(myRecord1.Filename)
	// fmt.Println(myRecord1.Filehash)

	err = db.AutoMigrate(&Record{})
	if err != nil {
		log.Fatal(err)
	}

	// insertRecord(db, &myRecord1)
	// insertRecord(db, &myRecord2)
	// record := querySingleRecord(db, "textfile_1.txt")
	// records := queryRecord(db)
	// for i := 0; i < len(records); i++ {
	// 	fmt.Println(records[i].Filename)
	// }
	// deleteRecord(db, "textfile_1.txt")

	fmt.Printf("Server starting on port %s...\n", PORT)
	err = http.ListenAndServe(net.JoinHostPort("localhost", PORT), nil)
	if err != nil {
		log.Fatal(err)
	}
}
