package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	types "github.com/wasif9/p2p-file-share/pkg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var err error
var node types.Node
var nodeArr = [10]types.Node{}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	node.IP, node.Port, node.Index, node.Status = "localhost", cfg.Port, cfg.Index, "Follower"
	nodeArr[node.Index] = node

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

	http.HandleFunc("/api/"+cfg.Version+"/manifests/", createManifestHandler(db))
	http.HandleFunc("/api/"+cfg.Version+"/manifests", createManifestsHandler(db))
	http.HandleFunc("/api/"+cfg.Version+"/kill", killHandler)
	http.HandleFunc("/api/"+cfg.Version+"/heartbeat", heartbeatHandler)
	http.HandleFunc("/api/"+cfg.Version+"/election/", electionHandler)
	http.HandleFunc("/api/"+cfg.Version+"/election/", leaderHandler)

	log.Printf("Server starting on port %s...\n", cfg.Port)
	err = http.ListenAndServe(net.JoinHostPort("0.0.0.0", cfg.Port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
