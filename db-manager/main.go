package main

import (
	"fmt"
	"log"
	"net/http"

	types "github.com/wasif9/p2p-file-share/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var err error

func main() {
	// Call election to determine if there needs to be a new leader.

	// Connection to the database
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=registry%d port=%s sslmode=disable TimeZone=America/Edmonton",
		cfg.Pg_host, cfg.Pg_user, cfg.Pg_password, cfg.Index, cfg.Pg_port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&types.Manifest{})
	if err != nil {
		log.Fatal(err)
	}

	go monitorLeader(db)

	log.Printf("Max timestamp: %d\n", getTimestamp(db))

	http.HandleFunc("/api/"+cfg.Version+"/manifests/", manifestHandler(db))
	http.HandleFunc("/api/"+cfg.Version+"/manifests", manifestsHandler(db))
	http.HandleFunc("/api/"+cfg.Version+"/kill", killHandler)
	http.HandleFunc("/api/"+cfg.Version+"/heartbeat", heartbeatHandler(db))
	http.HandleFunc("/api/"+cfg.Version+"/election/", electionHandler)
	http.HandleFunc("/api/"+cfg.Version+"/leader", leaderHandler)
	http.HandleFunc("/api/"+cfg.Version+"/catchup", catchupHandler(db))

	log.Printf("Node %d serving on %s...\n", cfg.Index, cfg.Address)
	err = http.ListenAndServe(cfg.Address, nil)
	if err != nil {
		log.Fatal(err)
	}

}
