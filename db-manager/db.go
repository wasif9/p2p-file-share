package main

import (
	"errors"

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
func insertManifest(db *gorm.DB, manifest *types.Manifest) (types.Manifest, error) {
	found, err := querySingleManifest(db, manifest.Name)

	if err == nil && found.Name == manifest.Name {
		return types.Manifest{}, errors.New("manifest already exists in the DB")
	}

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
func deleteManifest(db *gorm.DB, filename string) error {

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
func queryManifest(db *gorm.DB) (manifests []types.Manifest, Error error) {

	result := db.Find(&manifests)

	return manifests, result.Error

}

func queryManifestByPrefix(db *gorm.DB, prefix string) (manifests []types.Manifest, Error error) {

	result := db.Where("name LIKE ?", prefix+"%").Find(&manifests)

	return manifests, result.Error

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
