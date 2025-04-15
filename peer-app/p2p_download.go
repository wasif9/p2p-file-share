package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
	types "github.com/wasif9/p2p-file-share/pkg/models"
)

func DownloadManifest(fileName, fileCID, filePath string) {
	var providers []peer.ID
	var err error
	if providers, err = dhtLookup(fileCID); err != nil {
		PopupMessage("Fail to download file due to " + err.Error())
	}
	for {

		for _, provider := range providers {
			if err := requestFile(provider, fileName+".json", filePath, fileCID, "manifest", 0); err != nil {
				log.Println(err)
				continue
			}
			return
		}
	}
}

func DownloadFile(fileName, filePath string) {
	// Open the file
	file, err := os.Open(filepath.Join(filePath, ".p2p", fileName+".json"))
	if err != nil {
		log.Fatalf("failed to open file: %v", err)
	}
	defer file.Close()

	// Read all content
	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("failed to read file: %v", err)
	}

	// Unmarshal into struct
	var manifest types.ManifestData
	if err := json.Unmarshal(data, &manifest); err != nil {
		log.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// Download file from providers
	// TODO P2P chunking download
	var providers []peer.ID
	if providers, err = dhtLookup(manifest.Chunks[0]); err != nil {
		PopupMessage("Fail to download file due to " + err.Error())
		return
	}
	for {

		for _, provider := range providers {
			// 0 here is the chunk ID....
			if err := requestFile(provider, fileName, filePath, manifest.FileCID, "file", 0); err != nil {
				log.Println(err)
				continue
			}
			return
		}
	}
}

func dhtLookup(fileCID string) ([]peer.ID, error) {
	ctx := context.Background()

	// Decode the CID from the string
	c, err := cid.Decode(fileCID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert CID: %v", err)
	}

	log.Print("FIND PROVIDER")
	providerChan := KadDHT.FindProvidersAsync(ctx, c, 10)
	log.Print("FIND PROVIDER DONE")

	// Keep looking for providers until find at least 1
	var foundProvider []peer.ID
	for len(foundProvider) == 0 {
		for p := range providerChan {
			log.Println("Discovered provider:", p.ID.String())
			if p.ID != Node.ID() {
				log.Printf("FIND PEER %v\n", p.ID.ShortString())
				foundProvider = append(foundProvider, p.ID)
			}
		}
	}

	return foundProvider, nil
}

func requestFile(providerID peer.ID, fileName, dirPath, expectedCID, downloadType string, chunkID int) error {
	// Open a new stream to the provider
	ctx := context.Background()

	// Log which peer we're contacting
	log.Println("Attempting to request file from provider:", providerID.String())

	s, err := Node.NewStream(ctx, providerID, Protocol)
	if err != nil {
		return err
	}
	defer func() {
		if err := s.Close(); err != nil {
			// Closed by the other side (provider)
			log.Println("Peer app error when close request file request ", err)
		}
	}()

	// Send the file request
	request := types.DownloadRequest{
		FileName:   fileName,
		ChunkIndex: chunkID,
		Type:       downloadType,
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		log.Println("Error encoding JSON:", err)
		PopupMessage("Cannot download file due to JSON error")
		return err
	}
	log.Printf("Requesting file: %v\n", request)
	if _, err := s.Write(jsonData); err != nil {
		return err
	}

	// Read the response
	data, err := io.ReadAll(s)
	if err != nil {
		return err
	}

	// Check if the response indicates an error
	response := string(data)
	if strings.HasPrefix(response, "Error:") || strings.HasPrefix(response, "File not found") {
		log.Printf("Failed to get file %s from the provider %s, %s\n", fileName, providerID.String(), response)
		return fmt.Errorf("file Not Found from the provider")
	}

	// Construct the correct save path in the peer's directory
	savePath := filepath.Join(dirPath, fileName)

	// Write the file data to correct directory
	err = os.WriteFile(savePath, data, 0644)
	if err != nil {
		return err
	}

	log.Printf("Received and saved file as %s\n", fileName)

	// Integrity Check: Compare computed CID with expected CID
	computedCID, err := cidFromFile(savePath)
	if err != nil {
		log.Println("Error computing CID for downloaded file:", err)
		return err
	}
	if computedCID.String() == expectedCID {
		log.Println("✅ Integrity check passed: File matches expected CID.")
	} else {
		log.Println("❌ Integrity check failed: File does NOT match expected CID.")
		log.Printf("Expected CID: %s\n", expectedCID)
		log.Printf("Computed CID: %s\n", computedCID.String())
		return fmt.Errorf("integrity check failed")
	}

	return nil
}
