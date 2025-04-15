package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/peer"
)

func DownloadManifest(fileName, fileCID, filePath string) {
	var providers []peer.ID

	var err error
	if providers, err = dhtLookup1(fileCID); err != nil {
		PopupMessage("Fail to download file due to " + err.Error())
	}
	for {

		for _, provider := range providers {
			if err := requestFile1(provider, fileName+".json", filePath, fileCID); err != nil {
				continue
			}
			return
		}
	}
}

func dhtLookup1(fileCID string) ([]peer.ID, error) {
	ctx := context.Background()

	// Decode the CID from the string
	c, err := cid.Decode(fileCID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert CID: %v", err)
	}

	providerChan := KadDHT.FindProvidersAsync(ctx, c, 10)

	// Keep looking for providers until find at least 1
	var foundProvider []peer.ID
	for len(foundProvider) == 0 {
		for p := range providerChan {
			log.Println("Discovered provider:", p.ID.String())
			if p.ID != Node.ID() {
				foundProvider = append(foundProvider, p.ID)
			}
		}
	}

	return foundProvider, nil
}

func requestFile1(providerID peer.ID, fileName, dirPath, expectedCID string) error {
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
	log.Printf("Requesting file: %s\n", fileName)
	if _, err := s.Write([]byte(fileName + "\n")); err != nil {
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
