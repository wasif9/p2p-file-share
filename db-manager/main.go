package main

import (
	"fmt"
	"log"
	"net/http"

	types "github.com/wasif9/p2p-file-share/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var err error
var timestamp uint = 0

func main() {
	// Call election to determine if there needs to be a new leader.
	go monitorLeader()

	// Connection to the database
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=registry%d port=%s sslmode=disable TimeZone=UTC",
		cfg.Pg_host, cfg.Pg_user, cfg.Pg_password, cfg.Index, cfg.Pg_port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&types.Manifest{})
	if err != nil {
		log.Fatal(err)
	}

	// query for the largest entry in the timestamp column of the manifest table
	// first upsert a row with timestamp 0
	err = db.Exec("INSERT INTO manifests (name, hash, size, timestamp) VALUES ('', '', 0, 0) ON CONFLICT DO NOTHING").Error
	if err != nil {
		log.Fatal(err)
	}
	err = db.Model(&types.Manifest{}).Select("max(timestamp)").Scan(&timestamp).Error
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Max timestamp: %d\n", timestamp)

	http.HandleFunc("/api/"+cfg.Version+"/manifests/", createManifestHandler(db))
	http.HandleFunc("/api/"+cfg.Version+"/manifests", createManifestsHandler(db))
	http.HandleFunc("/api/"+cfg.Version+"/kill", killHandler)
	http.HandleFunc("/api/"+cfg.Version+"/heartbeat", heartbeatHandler)
	http.HandleFunc("/api/"+cfg.Version+"/election/", electionHandler)
	http.HandleFunc("/api/"+cfg.Version+"/leader", leaderHandler)

	log.Printf("Node %d serving on %s...\n", cfg.Index, cfg.Address)
	err = http.ListenAndServe(cfg.Address, nil)
	if err != nil {
		log.Fatal(err)
	}

}
