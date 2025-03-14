package main

import (
	types "github.com/wasif9/p2p-file-share/pkg/models"
	"gorm.io/gorm"
)

/***********************
* This function inserts a manifest into the registry database
* Input:
*		*gorm.DB db
* 		struct Manifest manifest
* Returns:
* 		None
*
************************/
func insertManifest(db *gorm.DB, manifest *types.Manifest) (newManifest types.Manifest, Error error) {

	result := db.Create(&manifest)

	return *manifest, result.Error

}

/***********************
* This function deletes a manifest by filename from the registry database.
* Input:
*		*gorm.DB db
* 		string filename
* Returns:
* 		None
*
************************/
func deleteManifest(db *gorm.DB, filename string) (Error error) {

	result := db.Where("name = ?", filename).Delete(&types.Manifest{})

	if result.RowsAffected == 0 {
		return result.Error
	} else {
		return nil
	}

}

/***********************
* This function queries the database for all the manifests. It returns an array of
* all the manifests
* Input:
*		*gorm.DB db
* Returns:
* 		[]Manifest manifest
*
************************/
func queryManifest(db *gorm.DB) (manifest []types.Manifest, Error error) {

	result := db.Find(&manifest)

	return manifest, result.Error

}

/***********************
* This function queries the database for a file and returns the manifest for the file.
* Input:
*		*gorm.DB db
* 		filename
* Returns:
* 		Manifest manifest
*
************************/
func querySingleManifest(db *gorm.DB, filename string) (manifest types.Manifest, Error error) {

	result := db.Where("name = ?", filename).First(&manifest)

	return manifest, result.Error

}
