package main

import (
	types "github.com/wasif9/p2p-file-share/pkg/models"
	"gorm.io/gorm"
)

/***********************
* This function inserts a record into the registry database
* Input:
*		*gorm.DB db
* 		struct Record record
* Returns:
* 		None
*
************************/
func insertRecord(db *gorm.DB, record *types.Manifest) (newRecord types.Manifest, Error error) {

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
func deleteRecord(db *gorm.DB, filename string) (Error error) {

	result := db.Where("name = ?", filename).Delete(&types.Manifest{})

	if result.RowsAffected == 0 {
		return result.Error
	} else {
		return nil
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
func queryRecord(db *gorm.DB) (record []types.Manifest, Error error) {

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
func querySingleRecord(db *gorm.DB, filename string) (record types.Manifest, Error error) {

	result := db.Where("name = ?", filename).First(&record)

	return record, result.Error

}
